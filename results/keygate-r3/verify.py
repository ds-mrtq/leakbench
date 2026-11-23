"""Verify committed r3 part-1 evidence from any working directory."""

from pathlib import Path
import runpy

out = Path(__file__).resolve().parent
helpers = runpy.run_path(str(out.parent / 'keygate-r2r3/verify.py'))
helpers['verify'](out)
