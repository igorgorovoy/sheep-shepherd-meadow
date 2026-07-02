# Sprite generation — WC3 / steampunk cluster sprites (DESIGN-0002)

Local text→image pipeline that turns the DESIGN-0002 prompt catalog into transparent PNG sprites and drops them into `web/public/sprites/`, replacing the placeholder vector SVGs.

- **Model:** SDXL base 1.0 (fp16), run on Apple Silicon **MPS**. Open weights, no HF login.
- **Transparency:** `rembg` (u2net) cuts the subject out of a plain background.
- **Post:** trim → fit into the DESIGN-0002 canvas size → save PNG. The floor tile is masked to a 2:1 isometric diamond.

## Setup (once)

```bash
cd tools/sprite-gen
./setup.sh          # uv venv (Python 3.12) + deps + downloads SDXL (~7 GB)
```

Requirements: `uv`, ~10 GB free disk, network to huggingface.co. First download is the slow part; generation afterwards is local.

## Generate

```bash
cd tools/sprite-gen
.venv/bin/python generate.py --list                 # show the catalog
# §8 priority order:
.venv/bin/python generate.py --only sheep-running,sheep-pending,sheep-failed,sheep-succeeded
.venv/bin/python generate.py --only dwarf-idle,dwarf-work,dwarf-offline
.venv/bin/python generate.py --only station-idle,station-active,tile-floor
.venv/bin/python generate.py                        # everything
```

Useful flags:
- `--n 4` — generate 4 variants per sprite (`name.v1.png` …); keep the best, rename to the target name.
- `--seed 1234` — override the base seed.
- `--no-bg` — skip the transparent cutout (debug the raw render).
- `--model <hf-id>` — swap the model (e.g. a WC3-style SDXL LoRA/checkpoint id).

Tune the look by editing `catalog.json` (`style_anchor`, per-sprite `prompt`, `seed`, `steps`, `guidance`, canvas `w`/`h`).

## Wire the PNGs into the game

The scaffold currently loads **SVG** via `Phaser.load.svg`. To consume the generated **PNG**s:

1. Point the manifest at the raster files:
   ```bash
   .venv/bin/python generate.py --update-manifest
   ```
   (rewrites `web/public/sprites/manifest.json` paths `.svg` → `.png`)
2. One-line engine change in `web/src/game/scene.ts` `preload()`: choose the loader by extension —
   `path.endsWith('.svg') ? this.load.svg(key, url) : this.load.image(key, url)`.
   (The manifest key derivation is extension-independent, so keys stay stable.)
3. `cd web && npm run build` — the PNGs ship in `dist/sprites`, and `make dashboard` embeds them into the shepherd binary.

Keys resolve to a placeholder token if a file is missing, so you can migrate sprite-by-sprite.

## Notes

- **Character consistency** across a dwarf's idle/work/offline uses a shared seed + fixed description, but SDXL isn't guaranteed identical. Generate with `--n` and pick a consistent set, or add an IP-Adapter/reference pass later.
- 16 GB RAM is enough for SDXL with attention slicing + VAE tiling (~30–60 s/image on MPS). Flux was skipped (too large for this box).
- `.venv/` and the HF model cache are not committed.
