#!/usr/bin/env bash
# Set up the local sprite-generation environment (Apple Silicon / MPS).
# Uses a uv-managed Python 3.12 venv (torch/diffusers have no 3.14 wheels yet),
# installs deps, and pre-downloads the SDXL model so first generation is fast.
set -euo pipefail
cd "$(dirname "$0")"

echo "== 1/3  create venv (uv-managed Python 3.12) =="
uv venv --python 3.12 --clear .venv

echo "== 2/3  install deps =="
uv pip install --python .venv -r requirements.txt

echo "== 3/3  pre-download SDXL base (fp16) into the HF cache =="
KMP_DUPLICATE_LIB_OK=TRUE .venv/bin/python - <<'PY'
import torch
from diffusers import StableDiffusionXLPipeline
print("torch", torch.__version__, "| mps:", torch.backends.mps.is_available())
# Constructing the pipeline downloads exactly the fp16 weights we need, then we drop it.
p = StableDiffusionXLPipeline.from_pretrained(
    "stabilityai/stable-diffusion-xl-base-1.0",
    torch_dtype=torch.float16, variant="fp16", use_safetensors=True)
del p
print("model cached OK")
PY

echo
echo "Ready. Generate with:"
echo "  .venv/bin/python generate.py --list"
echo "  .venv/bin/python generate.py --only sheep-running,sheep-pending,sheep-failed,sheep-succeeded"
echo "  .venv/bin/python generate.py            # whole catalog"
