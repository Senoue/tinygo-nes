# TinyGo NES エミュレーター開発記録

## プロジェクト概要

- Wio Terminal (ATSAMD51P19A, 120MHz Cortex-M4, 192KB SRAM, 512KB Flash) 上でTinyGoベースのファミコン(NES)エミュレーターを動かすプロジェクト
- fogleman/nes ライブラリをベースにTinyGo向けに最適化
- 現在のテストROM: DragonQuest.nes

## 解決した課題

### 1. TinyGo/Go バージョン互換性

- TinyGo 0.40.1がGo 1.26未対応 → Go 1.25をインストール
- `GOROOT=/opt/homebrew/Cellar/go@1.25/1.25.8/libexec` で指定

### 2. SPI型変更 (tinydisplay互換性)

- `tinydisplay/initdisplay` の SPI型が `drivers` v0.34.0 と非互換
- → `initdisplay` を使わず `wioterminal.go` で直接 ILI9341 初期化に変更
- → `tinygo.org/x/drivers` を v0.28.0 → v0.34.0 に更新
- SPI3を60MHzで初期化し、ILI9341をRotation270で使用

### 3. fogleman/nes のローカルフォーク

- `encoding/gob` (セーブステート), `os`/`path` (ファイルI/O), `image` パッケージはTinyGo baremetal非対応
- → `nes/` ディレクトリにフォーク、gob/Save/Load関連を全除去
- → `ines.go` をヘッダ定数 (`iNESFileMagic`) のみに簡略化
- → ROMパース処理は `nes_loader.go` (`NewConsoleFromString`) に移動

### 4. PPU メモリ最適化 (image.RGBA → パレットインデックス)

- 元: `*image.RGBA` × 2 (front/back) = 256×240×4×2 = **492KB** → SRAM 192KB に収まらない
- 修正: `*[256*240]byte` パレットインデックス方式に変更 (61,440バイト)
- `renderPixel()` で RGBA を書く代わりにパレットインデックス(1byte)を書き込み
- RGB565変換は描画時に `PaletteRGB565` テーブル(128バイト)で行う

### 5. PPU シングルバッファ化

- ダブルバッファ (front/back 120KB) → シングルバッファ (60KB)
- `StepFrame()` は同期的なのでダブルバッファ不要
- `setVerticalBlank()` のスワップ処理を削除 (コメント: `// single buffer - no swap needed`)

### 6. 行単位描画 (フレームバッファ削減)

- 元: 320×240×2 = 153KB の全画面バッファ
- 修正: `[320*2]byte` の1行バッファで行単位描画 (640バイト)
- `renderLine()` で256ピクセルをRGB565に変換、X方向32ピクセルオフセットで中央配置
- `display.DrawRGBBitmap8()` で1行ずつ転送

### 7. ROM読込方式 (embed.FS → 直接embed)

- `embed.FS` の `ReadFile()` がTinyGo baremetalで不安定
- → `//go:embed roms/DragonQuest.nes` で直接 `string` 型に埋め込み
- `[]byte` だとSRAMに配置される → `string` でFlashに配置
- ライセンス画面も同様に `//go:embed img/qr_and_license.bin` で `string` 型に埋め込み

### 8. ライセンス画面の150KBコピー問題

- `[]byte(licenseData)` で150KBのRAMコピーが発生
- TinyGoの保守的GCが回収できず、PPUバッファ確保時にOOM
- → ライセンス画面を一時スキップして対処 (`license()` 関数は定義済みだが未呼出し)

### 9. ROMデータの直接パース (binary.Read回避)

- `encoding/binary.Read` + `bytes.NewReader` の代わりに文字列から直接バイト読み取り
- `NewConsoleFromString()` でiNESヘッダを手動パース
- `runtime.GC()` をPRG/CHRコピー前後に挿入してメモリ断片化を抑制

### 10. APU float64 除算の排除

- APU.Step() が毎サイクル `float64(cycle) / frameCounterRate` を実行
- ATSAMD51はfloat64 FPU非搭載 → ソフトウェア演算で1フレーム29,780回の除算が致命的に遅い
- → 整数剰余 `cycle % frameCounterPeriod` に置換 (`frameCounterPeriod = CPUFrequency / 240 = 7457`)
- オーディオサンプリングも完全スキップ (音声出力なし)

## 現在のメモリ使用量

```
静的RAM: ~12KB
動的:
  PRG ROM:   32KB
  CHR ROM:   32KB
  SRAM:       8KB
  NES RAM:    2KB
  PPU buffer: 60KB
  合計:     ~146KB / 192KB
```

## 今後のチャレンジ

- SDカードからのROM読込対応
- ライセンス画面の復活（Flash直接読みで150KBコピー回避）
- パフォーマンス計測と最適化（目標: 10fps以上）
- APUフレームカウンターの精度改善
- `fmt` パッケージ除去（cpu.go / mapper.go の `fmt.Errorf`）でFlash節約
- `math` パッケージ除去（filter.go の `math.Pi`）の検討
- `log` パッケージ除去（memory.go / mapper*.go の `log.Fatalf`）の検討

## ビルド方法

```bash
PATH="/opt/homebrew/Cellar/go@1.25/1.25.8/bin:$PATH" \
GOROOT=/opt/homebrew/Cellar/go@1.25/1.25.8/libexec \
tinygo build -target=wioterminal -o tinygo-nes.uf2 .
```

デスクトップ（tinydisplay）でのテスト:

```bash
go run .
```

## ファイル構成

### トップレベル

| ファイル | 説明 |
|---------|------|
| `main.go` | エントリポイント。ROMとライセンスデータの埋め込み、エミュレーションループ（フレーム実行→行単位描画）、フレームカウンター表示 |
| `nes_loader.go` | iNESフォーマットのROMパーサー。`string`型の埋め込みデータから直接ヘッダ読み取り、PRG/CHR ROMをコピーしてCartridgeを生成 |
| `wioterminal.go` | Wio Terminal固有コード（ビルドタグ: `wioterminal`）。SPI3経由のILI9341ディスプレイ初期化、ハードウェアボタン（5方向スイッチ+3ボタン）からのNESコントローラー入力読み取り |
| `generic.go` | デスクトップ向けコード（ビルドタグ: `!baremetal`）。tinydisplayによるウィンドウ表示、キーボード入力からのNESコントローラー入力読み取り |
| `go.mod` | モジュール定義（`github.com/Senoue/tinygo-nes`）。依存: `tinydisplay` v0.3.0, `drivers` v0.34.0 |
| `go.sum` | 依存パッケージのチェックサム |

### nes/ ディレクトリ（fogleman/nesのTinyGo向けフォーク）

| ファイル | 説明 |
|---------|------|
| `console.go` | NES本体の統合。CPU/APU/PPU/Mapper/Controllerの生成と接続、`Step()`（1CPU命令実行）、`StepFrame()`（1フレーム実行）、`Buffer()`（PPUバッファ参照） |
| `cpu.go` | 6502 CPU実装。全命令のデコード/実行、割り込み処理（NMI/IRQ）、アドレッシングモード |
| `ppu.go` | PPU実装。シングルバッファ方式のパレットインデックスレンダリング、スキャンライン/サイクル単位のタイミング、スプライト評価、背景/スプライト合成 |
| `apu.go` | APU実装。パルス×2/三角波/ノイズ/DMCチャンネル、フレームカウンター（整数剰余方式）、オーディオ出力はスキップ |
| `palette.go` | NESパレットデータ。64色の`color.RGBA`テーブルと、ILI9341向けRGB565変換済みテーブル(`PaletteRGB565`) |
| `cartridge.go` | カートリッジデータ構造。PRG ROM、CHR ROM、SRAM、マッパー番号、ミラーリングモード |
| `controller.go` | NESコントローラー実装。8ボタン（A/B/Select/Start/上下左右）のシリアル読み出し |
| `memory.go` | メモリマップ実装。CPUメモリ（RAM/PPUレジスタ/APU/コントローラー/マッパー）とPPUメモリ（CHR/ネームテーブル/パレット）のアドレスデコード、ミラーリング処理 |
| `mapper.go` | マッパーインターフェースと生成関数。対応マッパー: 0, 1, 2, 3, 4, 7, 40, 225 |
| `mapper1.go` | MMC1 (Mapper 1)。シフトレジスタ方式のバンク切替、PRG/CHRバンクモード対応 |
| `mapper2.go` | UxROM (Mapper 2)。Mapper 0も兼用。16KB PRGバンク切替 |
| `mapper3.go` | CNROM (Mapper 3)。8KB CHRバンク切替 |
| `mapper4.go` | MMC3 (Mapper 4)。IRQカウンター付きバンク切替 |
| `mapper7.go` | AxROM (Mapper 7)。32KB PRGバンク切替、シングルスクリーンミラーリング |
| `mapper40.go` | Mapper 40。IRQカウンター付き |
| `mapper225.go` | Mapper 225。大容量マルチカートリッジ対応 |
| `ines.go` | iNESフォーマット定数（`iNESFileMagic = 0x1a53454e`）のみ |
| `filter.go` | オーディオフィルター。ハイパス/ローパスの1次IIRフィルター（現在は未使用） |

### リソースディレクトリ

| ディレクトリ/ファイル | 説明 |
|---------|------|
| `roms/DragonQuest.nes` | テスト用NES ROM |
| `img/qr_and_license.bin` | ライセンス画面のRGB565バイナリデータ (320×240) |
| `png/qr_and_license.png` | ライセンス画面の元画像 |
