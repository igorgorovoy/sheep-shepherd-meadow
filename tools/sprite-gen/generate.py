#!/usr/bin/env python3
"""
Generate WC3 / steampunk cluster sprites (DESIGN-0002) locally on Apple Silicon.

Pipeline per sprite: SDXL text->image (MPS) -> transparent cutout (rembg) ->
trim + fit into the target canvas -> save PNG into web/public/sprites/.

Examples:
  python generate.py --list
  python generate.py --only sheep-running,sheep-failed
  python generate.py --only dwarf-idle --n 4          # 4 variants, pick the best
  python generate.py                                   # whole catalog
  python generate.py --update-manifest                 # point manifest.json at the .png files
"""
import argparse, json, os, sys, pathlib

# Keep a messy global OpenMP from crashing the process on macOS.
os.environ.setdefault("KMP_DUPLICATE_LIB_OK", "TRUE")
os.environ.setdefault("TOKENIZERS_PARALLELISM", "false")
os.environ.setdefault("PYTORCH_ENABLE_MPS_FALLBACK", "1")

HERE = pathlib.Path(__file__).resolve().parent
REPO = HERE.parent.parent  # tools/sprite-gen -> repo root
DEFAULT_OUT = REPO / "web" / "public" / "sprites"
MANIFEST = DEFAULT_OUT / "manifest.json"


def load_catalog():
    with open(HERE / "catalog.json") as f:
        return json.load(f)


def build_pipe(model, device):
    import torch
    from diffusers import AutoPipelineForText2Image
    # fp16 on MPS produces NaN/blank output; SDXL fp32 is too heavy for 16 GB.
    # Default model is a light SD1.5 checkpoint run in fp32 on MPS.
    dtype = torch.float16 if device == "cuda" else torch.float32
    print(f"→ loading {model} ({dtype}) on {device} …")
    try:
        pipe = AutoPipelineForText2Image.from_pretrained(
            model, torch_dtype=dtype, use_safetensors=True, variant="fp16")
    except Exception:
        pipe = AutoPipelineForText2Image.from_pretrained(
            model, torch_dtype=dtype, use_safetensors=True)
    pipe = pipe.to(device)
    pipe.set_progress_bar_config(disable=True)
    # SD1.5 ships a safety checker that can blank a sprite on a false positive.
    if hasattr(pipe, "safety_checker"):
        pipe.safety_checker = None
        pipe.requires_safety_checker = False
    try:
        pipe.enable_attention_slicing()
    except Exception:
        pass
    return pipe


def cutout(img):
    """Transparent-background cutout via rembg -> RGBA."""
    from rembg import remove, new_session
    if not hasattr(cutout, "_sess"):
        cutout._sess = new_session("u2net")
    return remove(img.convert("RGBA"), session=cutout._sess, post_process_mask=True)


def fit_into(img, w, h, fill=0.92):
    """Trim transparent margin, scale to fit (w,h)*fill, center on a transparent canvas."""
    from PIL import Image
    img = img.convert("RGBA")
    bbox = img.getbbox()
    if bbox:
        img = img.crop(bbox)
    tw, th = int(w * fill), int(h * fill)
    scale = min(tw / img.width, th / img.height)
    img = img.resize((max(1, int(img.width * scale)), max(1, int(img.height * scale))), Image.LANCZOS)
    canvas = Image.new("RGBA", (w, h), (0, 0, 0, 0))
    canvas.paste(img, ((w - img.width) // 2, (h - img.height) // 2), img)
    return canvas


def diamond(img, w, h):
    """Resize to (w,h) and clip to a 2:1 isometric diamond alpha."""
    from PIL import Image, ImageDraw
    img = img.convert("RGBA").resize((w, h), Image.LANCZOS)
    mask = Image.new("L", (w, h), 0)
    ImageDraw.Draw(mask).polygon([(w / 2, 0), (w, h / 2), (w / 2, h), (0, h / 2)], fill=255)
    img.putalpha(mask)
    return img


def _is_blank(img):
    # fp32 on MPS intermittently yields NaN latents that decode to a flat
    # (near-uniform) image. A real sprite has structure → high stddev.
    import numpy as np
    return float(np.asarray(img.convert("RGB"), dtype="float32").std()) < 10.0


def _gen_valid(pipe, prompt, neg, steps, guidance, gen, base, tries=6):
    """Generate, retrying with a fresh seed while the output is blank/NaN."""
    import torch
    for k in range(tries):
        seed = base + k * 7
        g = torch.Generator("cpu").manual_seed(seed)
        img = pipe(prompt=prompt, negative_prompt=neg, num_inference_steps=steps,
                   guidance_scale=guidance, width=gen, height=gen, generator=g).images[0]
        if not _is_blank(img):
            return img, seed
        print(f"      · blank/NaN at seed {seed}; retrying")
    return None, base


def generate(cat, entry, pipe, args):
    import torch  # noqa: F401 (kept for parity / future use)
    d = cat["defaults"]
    # Subject FIRST, style after — SDXL/SD1.5 CLIP truncates at 77 tokens, so
    # the subject must lead or it gets cut off.
    prompt = f"{entry['prompt']}, {cat['style_anchor']}"
    steps = entry.get("steps", d["steps"])
    guidance = entry.get("guidance", d["guidance"])
    gen = d["gen"]
    transparent = entry.get("transparent", d["transparent"])
    post = entry.get("postproc")
    base_seed = args.seed if args.seed is not None else entry["seed"]
    out = DEFAULT_OUT / entry["out"]
    out.parent.mkdir(parents=True, exist_ok=True)

    for i in range(args.n):
        img, seed = _gen_valid(pipe, prompt, cat["negative"], steps, guidance, gen, base_seed + i)
        if img is None:
            print(f"   ✗ {entry['key']} still blank/NaN after retries — skipped")
            continue
        if post == "diamond":
            img = diamond(img, entry["w"], entry["h"])
        elif transparent and not args.no_bg:
            img = fit_into(cutout(img), entry["w"], entry["h"])
        else:
            img = fit_into(img, entry["w"], entry["h"], fill=1.0)
        dest = out if args.n == 1 else out.with_name(f"{out.stem}.v{i+1}{out.suffix}")
        img.save(dest)
        print(f"   ✓ {dest.relative_to(REPO)}  (seed {seed})")


def update_manifest():
    if not MANIFEST.exists():
        print("manifest.json not found; skip"); return
    m = json.loads(MANIFEST.read_text())
    def png(x): return x.replace(".svg", ".png") if isinstance(x, str) else x
    if "sheep" in m: m["sheep"]["path"] = png(m["sheep"].get("path", ""))
    if "dwarves" in m: m["dwarves"]["path"] = png(m["dwarves"].get("path", ""))
    if "tiles" in m and "floor" in m["tiles"]: m["tiles"]["floor"] = png(m["tiles"]["floor"])
    for k, v in m.get("structures", {}).items():
        for kk, vv in list(v.items()):
            v[kk] = png(vv)
    MANIFEST.write_text(json.dumps(m, indent=2) + "\n")
    print("→ manifest.json now points at .png "
          "(NOTE: scene.ts preload must load raster via load.image for .png — see README)")


def main():
    cat = load_catalog()
    keys = [s["key"] for s in cat["sprites"]]
    ap = argparse.ArgumentParser()
    ap.add_argument("--only", help="comma-separated sprite keys")
    ap.add_argument("--n", type=int, default=1, help="variants per sprite")
    ap.add_argument("--seed", type=int, default=None, help="override base seed")
    ap.add_argument("--model", default=cat["model"])
    ap.add_argument("--no-bg", action="store_true", help="skip transparent cutout")
    ap.add_argument("--update-manifest", action="store_true")
    ap.add_argument("--list", action="store_true")
    args = ap.parse_args()

    if args.list:
        for s in cat["sprites"]:
            print(f"  {s['key']:16s} -> {s['out']}  ({s['w']}x{s['h']})")
        return
    if args.update_manifest and not args.only:
        update_manifest(); return

    want = set(args.only.split(",")) if args.only else set(keys)
    todo = [s for s in cat["sprites"] if s["key"] in want]
    if not todo:
        print(f"no matching keys. available: {', '.join(keys)}"); sys.exit(1)

    import torch
    device = "mps" if torch.backends.mps.is_available() else ("cuda" if torch.cuda.is_available() else "cpu")
    pipe = build_pipe(args.model, device)
    for entry in todo:
        print(f"● {entry['key']}")
        try:
            generate(cat, entry, pipe, args)
        except Exception as e:
            print(f"   ✗ {entry['key']} failed: {e}")
    if args.update_manifest:
        update_manifest()
    print("done.")


if __name__ == "__main__":
    main()
