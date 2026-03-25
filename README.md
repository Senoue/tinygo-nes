# tinygo-nes

Wio Terminal (ATSAMD51, Cortex-M4 120MHz, 192KB SRAM) で動作する TinyGo 製ファミコン (NES) エミュレーター。

## 対応

- **Mapper 0** (NROM) - Super Mario Bros. など
- **Mapper 1** (MMC1)
- **Mapper 2** (UxROM)
- **Mapper 3** (CNROM) - Dragon Quest など
- **Mapper 4** (MMC3)
- **Mapper 7** (AxROM)

## ボタン配置

```
Wio Terminal          NES
─────────────────────────
BUTTON_1 (左)    →   Select
BUTTON_2 (中)    →   B
BUTTON_3 (右)    →   Start
5方向スティック   →   十字キー
5S_PRESS (押込)  →   A
```

## ビルド

### 必要なもの

- Go 1.25
- TinyGo 0.40.1

### 手順

1. ROMファイルを `roms/game.nes` として配置（別名にする場合は `main.go` の embed パスも変更）
```bash
mkdir -p roms
cp /path/to/your/game.nes roms/game.nes
```

3. ビルド
```bash
PATH="/opt/homebrew/Cellar/go@1.25/1.25.8/bin:$PATH" \
GOROOT=/opt/homebrew/Cellar/go@1.25/1.25.8/libexec \
tinygo build -target=wioterminal \
  -gc=conservative -opt=2 -stack-size=16KB \
  -scheduler=none -panic=trap -nobounds \
  -o tinygo-nes.uf2 .
```

4. Wio Terminal をブートローダーモードにして `tinygo-nes.uf2` をドラッグ&ドロップ

## スペック

| 項目 | 値 |
|---|---|
| Flash | ~109KB |
| 静的RAM | ~8KB |
| 動的RAM | ~134KB (PRG+CHR+PPU等) |
| 合計RAM | ~142KB / 192KB |
| PPU | スキャンラインレンダラー |
| 描画 | 256x240, RGB565, 24行バッチ |
| SPI | 80MHz |
| 音声 | DAC出力 (実験的) |

## 技術的な工夫

- **fogleman/nes をローカルフォーク** - gob/os/image/fmt/log/math パッケージを全除去
- **PPU シングルバッファ** - image.RGBA (492KB) → パレットインデックス byte配列 (60KB)
- **スキャンラインレンダラー** - PPU処理を 89,342回/フレーム → 262回/フレーム に削減
- **Memory インターフェース除去** - TinyGo の scheduler=none でインターフェースディスパッチが動作しない問題を回避
- **CPU 関数テーブル → switch文** - 256個のクロージャを直接ディスパッチに変換
- **APU float64 除去** - 整数カウントダウン方式に変更 (ATSAMD51にfloat64 FPUなし)
- **手動CPU Reset** - リセットベクターをPRG ROMから直接読み込み

## ファイル構成

```
main.go          - エントリポイント、ROM読込、エミュレーションループ
nes_loader.go    - iNESフォーマットパーサー
wioterminal.go   - Wio Terminal固有 (ILI9341, ボタン, DAC)
generic.go       - PC用ビルド (tinydisplay)
nes/             - NESエミュレーターコア (fogleman/nes フォーク)
  cpu.go         - 6502 CPU (256-case switch)
  ppu.go         - PPU (スキャンラインレンダラー)
  apu.go         - APU (フレームカウンター + IRQ)
  console.go     - Console (スキャンラインベースStepFrame)
  mapper*.go     - カートリッジマッパー
  palette.go     - NESパレット (RGBA + RGB565)
  controller.go  - コントローラー入力
  cartridge.go   - カートリッジデータ
  filter.go      - オーディオフィルター
  memory.go      - ミラーリングアドレス計算
```

## ベースプロジェクト

- [fogleman/nes](https://github.com/fogleman/nes) - NES emulator written in Go
