#!/usr/bin/env python3
"""Build isolated TinyGo 0.43.0 clock experiments, without patching the installation."""
import argparse
import os
from pathlib import Path
import subprocess
import tempfile


def replace_once(text, old, new):
    if text.count(old) != 1:
        raise RuntimeError('TinyGo source mismatch: ' + old)
    return text.replace(old, new, 1)


def overlay(src, dst, patched):
    dst.mkdir()
    for entry in src.iterdir():
        rel = entry.relative_to(root)
        if rel in patched:
            (dst / entry.name).write_text(patched[rel])
        elif entry.is_dir() and (rel.parts[0] == 'src' or any(rel in p.parents for p in patched)):
            overlay(entry, dst / entry.name, patched)
        else:
            (dst / entry.name).symlink_to(entry, target_is_directory=entry.is_dir())


parser = argparse.ArgumentParser()
parser.add_argument('mhz', type=int, choices=[120, 150])
parser.add_argument('--lcd-stock', action='store_true', help='Use the normal LCD SPI request (80 MHz) instead of 24 MHz')
parser.add_argument('--lcd-dma', action='store_true', help='Use synchronous LCD DMA (Wio Terminal channel 0)')
parser.add_argument('--apu-lite', action='store_true', help='Skip waveform state unused by the buzzer')
args = parser.parse_args()
project = Path(__file__).resolve().parents[1]
version = subprocess.check_output(['tinygo', 'version'], text=True)
if not version.startswith('tinygo version 0.43.0 '):
    raise SystemExit('This experiment requires TinyGo 0.43.0')
root = Path(subprocess.check_output(['tinygo', 'env', 'TINYGOROOT'], text=True).strip()).resolve()
# Homebrew's bin/tinygo wrapper forcibly resets TINYGOROOT. Use the actual binary.
compiler = root / 'libexec/tinygo'
if not compiler.is_file():
    import shutil
    compiler = Path(shutil.which('tinygo')).resolve()
runtime_path = Path('src/runtime/runtime_atsamd51.go')
machine_path = Path('src/machine/machine_atsamd51.go')
runtime = (root / runtime_path).read_text()
runtime = replace_once(runtime, '(59 << sam.OSCCTRL_DPLL_DPLLRATIO_LDR_Pos)',
                       f'({args.mhz // 2 - 1} << sam.OSCCTRL_DPLL_DPLLRATIO_LDR_Pos)')
# Conservative experimental setting shared by both comparison builds.
runtime = replace_once(runtime, 'sam.NVMCTRL.CTRLA.SetBits(0 << sam.NVMCTRL_CTRLA_RWS_Pos)',
    'sam.NVMCTRL.CTRLA.Set((sam.NVMCTRL.CTRLA.Get() &^ '
    '(sam.NVMCTRL_CTRLA_AUTOWS | sam.NVMCTRL_CTRLA_RWS_Msk)) | '
    '(7 << sam.NVMCTRL_CTRLA_RWS_Pos))')
machine = (root / machine_path).read_text()
machine = replace_once(machine, 'return 120000000', f'return {args.mhz * 1000000}')
machine = replace_once(machine, 'SERCOM_FREQ_REF_GCLK0 = 120000000',
                       f'SERCOM_FREQ_REF_GCLK0 = {args.mhz * 1000000}')
machine = replace_once(machine, 'top = period * 3 / 25',
                       f'top = period * {args.mhz} / 1000')
with tempfile.TemporaryDirectory(prefix='tinygo-oc-') as temp:
    target = Path(temp) / 'root'
    overlay(root, target, {runtime_path: runtime, machine_path: machine})
    env = os.environ.copy()
    env['TINYGOROOT'] = str(target)
    actual_root = subprocess.check_output([str(compiler), 'env', 'TINYGOROOT'], env=env, text=True).strip()
    if Path(actual_root).resolve() != target.resolve():
        raise RuntimeError('Compiler ignored experimental TINYGOROOT; refusing build')
    suffix = 'r6-txc-lcd-stock' if args.lcd_stock else 'r6-txc-lcd24'
    tags = 'buzzer,perf' if args.lcd_stock else 'buzzer,perf,ocbench'
    if args.lcd_dma:
        suffix = 'r7-dma-lcd-stock' if args.lcd_stock else 'r7-dma-lcd24'
        tags += ',lcddma'
    if args.apu_lite:
        suffix = 'r8-apulite-' + ('dma-' if args.lcd_dma else '') + ('lcd-stock' if args.lcd_stock else 'lcd24')
        tags += ',apulite'
    suffix = 'r9-core-' + suffix.split('-', 1)[1]
    output = project / f'tinygo-nes-oc{args.mhz}-{suffix}.uf2'
    subprocess.run([str(compiler), 'build', '-target=wioterminal', '-tags=' + tags,
        '-gc=conservative', '-opt=2', '-stack-size=32KB', '-scheduler=none',
        '-panic=trap', '-nobounds', '-o', str(output), '.'], cwd=project, env=env, check=True)
    spi_mhz = args.mhz / 2 if args.lcd_stock else 24
    print(f'Built {output}: CPU {args.mhz} MHz / expected LCD SPI {spi_mhz:g} MHz / flash RWS 7')
