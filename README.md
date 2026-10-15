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

- Go 1.27.2
- TinyGo 0.43.0

### 手順

1. ROMファイルを `roms/game.nes` として配置（別名にする場合は `main.go` の embed パスも変更）
```bash
mkdir -p roms
cp /path/to/your/game.nes roms/game.nes
```

3. ビルド
```bash
PATH="/opt/homebrew/opt/go/bin:$PATH" \
GOROOT=/opt/homebrew/opt/go/libexec GOTOOLCHAIN=local \
tinygo build -target=wioterminal \
  -gc=conservative -opt=2 -stack-size=16KB \
  -scheduler=none -panic=trap -nobounds \
  -o tinygo-nes.uf2 .
```

4. Wio Terminal をブートローダーモードにして `tinygo-nes.uf2` をドラッグ&ドロップ

## 処理時間の計測

通常ビルドでは計測しません。上記のビルドコマンドに `-tags=perf` を追加すると、
120エミュレーションフレームごとに `println` で集計を出力します。
Wio TerminalではUSBシリアルのモニターで確認できます。
PCで実行する場合は `go run -tags=perf .` の端末に出力されます。

| 出力項目 | 意味 |
|---|---|
| `frames` / `displayed` | 計測区間内の実行フレーム数 / 画面更新回数 |
| `window_ms` | 計測区間の経過時間（ミリ秒） |
| `emu_fps_x10` | 実行FPSの10倍（例: 245なら24.5 FPS） |
| `display_fps_x10` | 画面更新FPSの10倍 |
| `emu_us/frame` | `StepFrame()`の平均時間（マイクロ秒 / 実行フレーム） |
| `convert_us/display` | 全24行バッチのRGB565変換の合計平均時間（マイクロ秒 / 画面更新） |
| `transfer_us/display` | 全24行バッチの画面転送呼び出しの合計平均時間（マイクロ秒 / 画面更新） |

FPSの経過時間には入力・音声処理や計測自体の負荷も含みます。
ログ出力時間は集計区間から除外しますが、出力中はエミュレーションが止まります。
転送時間はドライバー呼び出しの所要時間で、通信だけを独立に測った値ではありません。
同じROM・同じ場面で複数区間を比較してください。起動直後の区間は別扱いにします。
現在は画面更新を2フレームに1回にしているため、画面更新FPSは通常、実行FPSの約半分です。

`perf` に続く2行の `perf_core` は、各120フレーム区間の最後の2フレームを測った本体の内訳です。
画素生成を省略したフレームと描画したフレームを1回ずつ計測します。

| 出力項目 | 意味 |
|---|---|
| `samples=1` | 各描画モードで詳細計測したフレーム数（区間平均ではなく1フレームの測定） |
| `frame_output=0` / `frame_output=1` | 画素生成を省略 / 画素生成あり |
| `mode=scanline` | 走査線単位でCPU＋APUとPPUを分けて計測 |
| `cpu_apu_us` | CPU＋APUの合計処理時間（マイクロ秒） |
| `ppu_us` | PPUの処理時間（Mapper更新を含む、マイクロ秒） |
| `mode=cycle_total` | Mapper 40用。フレーム全体のみ計測し、内訳は出力しない |
| `total_us` | 詳細計測フレーム全体の所要時間 |
| `regular_emu_us/frame` | 詳細計測を除く118フレームの本体処理の平均時間 |

CPUとAPUは従来どおり交互に実行し、1走査線分をまとめて測ります。
時計読み取りはCPU命令ループの外に置き、1走査線あたり4回（フレーム全体の計測を除く）に抑えています。
各内訳には少量の時計読み取り負荷が含まれます。以前の命令単位の `cpu_us` / `apu_us` とは直接比較できません。
`regular_emu_us/frame` は描画あり・省略を合わせた平均です。各 `total_us` との差には描画の有無、計測負荷、フレーム内容の差が含まれます。
前のログとの速度比較には `regular_emu_us/frame` を使い、内訳は複数区間で傾向を見てください。
`emu_us/frame` とFPSには詳細計測による遅延も含まれます。
Mapper 40はCPU・PPU・APUを命令ごとに進めるため、低負荷で分離せず全体時間のみ測ります。
通常ビルドでは時計読み取りを行いません。

画面更新をしないフレームは `SetFrameOutput(false)` により画素生成も省略します。
CPU/APU、スクロール、NMI、Mapper IRQ、スプライト評価・オーバーフロー判定は進め続けます。
走査線方式ではSprite 0と重なる背景だけを調べ、画面全体の色変換・合成・バッファ書き込みを省きます。
描画側も元のパターンの透明情報でSprite 0と優先度を判定します（同じ表示色でも不透明画素は衝突します）。
Mapper 40のサイクル方式ではタイミングを維持し、最終的な画素書き込みのみ省略します。
比較には同じ場面で `emu_fps_x10`、`regular_emu_us/frame`、両モードの `ppu_us` を確認してください。

ROMやGUIに依存しない計測ロジックのテスト:

```bash
go test -tags=perf performance_enabled.go performance_test.go
go test -tags=perf ./nes
```

## 修正内容と実機計測結果（2026-10-15）

ROM初期化の不具合を修正した後、処理時間の計測を追加し、画面に出さないフレームの描画計算を削減しました。
作業ブランチは `fix/rom-loading-mapper-init` です。

### ROM読み込み・初期化

- `main.go` 内の重複したROM解析と初期化を `NewConsoleFromString()` → `nes.NewConsole()` に統一。
- iNESヘッダー、PRGの有無、データ長、PRG/CHRサイズを検証し、未対応MapperやNES 2.0をエラーとして扱うように変更。
- 既存ローダーのTrainer 512バイトのスキップと、CHR ROMなしの場合の8KB CHR RAM確保を起動経路でも利用。
- Mapper 4/40の更新処理が呼ばれるよう初期化を統一。Mapper 40はPPUサイクル単位でIRQカウンターを進める経路を使用。
- 手動CPUリセットを維持し、初期バンク配置に合わせてMapper 7は先頭32KB、Mapper 40は固定バンクのリセットベクターを参照。

### 計測と描画省略

- `perf` ビルドタグでのみ時間計測とUSBシリアル出力を有効化。120フレームごとにFPS、本体処理、RGB565変換、画面転送を集計。
- CPU命令ごとの詳細計測は負荷が大きかったため、CPU＋APUとPPUを走査線単位で測る方式に変更。
- 変更前は画面転送だけを2フレームに1回省略していた。変更後はそのフレームで `SetFrameOutput(false)` を指定し、PPUの画素生成も省略。
- 非表示フレームでは背景全体の色生成、スプライト合成、フレームバッファ書き込みを省く。Sprite 0と重なる最大8画素の背景パターンだけを調べ、衝突判定を維持。
- CPU/APU、スクロール、NMI、Mapper IRQ、スプライト評価・オーバーフロー判定は継続。表示フレームは通常どおり生成する。
- 表示側の透明判定も最終的な表示色ではなく元のパターン値に変更。Sprite 0判定、重なるスプライトの優先度、描画無効時に古い画素が残る問題を修正。
- 透明情報の保持用に走査線1本分の256バイトを追加。Mapper 40ではサイクル処理を維持し、最終的な画素書き込みだけを省略する。
- 詳細ログは描画省略／通常描画の2種類を出力し、両方を除いた118フレームの平均も表示。

### Wio Terminalでの結果

ユーザー提供の実機ログから、起動直後の区間を除いて比較しました。Go 1.27.2 / TinyGo 0.43.0、
`-tags=perf -gc=conservative -opt=2 -stack-size=16KB -scheduler=none -panic=trap -nobounds` の構成です。

| 指標 | 画素生成省略前 | 省略後 |
|---|---:|---:|
| 実行FPS（ログ表示値） | 10.3〜10.4 | 12.4 |
| 画面更新FPS（ログ表示値） | 5.1〜5.2 | 6.2 |
| 本体平均・詳細計測を除く | 81.85〜82.46ms | 64.92〜65.66ms |
| 非表示フレームのPPU処理 | 34.06〜36.01ms（毎フレーム描画） | 3.42〜3.91ms |
| 表示フレームのPPU処理 | 34.06〜36.01ms | 32.23〜33.33ms |

実行速度は約20%向上し、非表示フレームのPPU処理は約90%減少しました。
PPU時間は各区間で抽出したフレームの測定値であり、全フレームの平均ではありません。
この結果は提供された場面のログに基づき、全ROM・全場面での改善幅を保証するものではありません。

### 検証と残っている作業

- ROM解析・初期化、Mapper IRQ、集計値、サンプル選択の回帰テストが通過。
- 実装済みMapper 0/1/2/3/4/7/40/225について、描画あり／省略でCPUサイクルとエミュレーター状態が一致し、次の表示フレームの画像も一致することを自作テストROMで確認。
- Sprite 0、左端マスク、画面右端、スクロールとネームテーブル境界、スプライト優先度、描画無効時の画面をテスト。
- 計測あり／なしのWio Terminal向けビルドが成功。Homebrew版Goを1.27.2へ更新し、READMEのビルド手順も更新。
- ゲーム操作・スクロール・表示の実機での継続確認は未完了。音声の周期的な出力処理は未対応。
- 次の主な改善候補はCPU＋APU（実機で約50ms／フレーム）。今回、CPU命令実行とAPUの処理方式自体は変更していない。
- 下記スペックのメモリ数値は以前の概算であり、今回の変更後の使用量は再計測していない。

再確認用コマンド（ROMやGUIは不要）:

```bash
go test nes_loader.go nes_loader_test.go
go test ./nes
go test -tags=perf ./nes
go test -tags=perf performance_enabled.go performance_test.go
```

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

## ROM の入手

本リポジトリにROMファイルは含まれていません。以下の方法で合法的に入手できます。

### Homebrew ROM (おすすめ)

NES開発者コミュニティが公開している自作ゲームです。無料で合法的に利用できます。

- [NESdev Homebrew](https://www.nesdev.org/wiki/Homebrew) - コミュニティ作品リスト
- [NES Homebrew on itch.io](https://itch.io/games/tag-nes) - 無料ゲーム多数
- [nestest.nes](https://www.nesdev.org/wiki/Emulator_tests) - CPUテスト用ROM (エミュレーター動作確認に最適)

### カートリッジ吸い出し

自分が所有するファミコン/NESカートリッジから専用機器でROMデータをダンプできます。

## ベースプロジェクト

- [fogleman/nes](https://github.com/fogleman/nes) - NES emulator written in Go
