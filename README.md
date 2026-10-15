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

## 内蔵ブザーでBGMの旋律を試す

`buzzer` タグ付きビルドでは、Wio Terminalの内蔵ブザー（`WIO_BUZZER` / PD11）をTCC0のPWMで駆動します。
内蔵ブザーはPWMで発音するパッシブ型です（[Seeed公式資料](https://wiki.seeedstudio.com/Wio-Terminal-Buzzer/)）。
従来のDACへの1フレーム1回の書き込みを置き換え、PWMがエミュレーションや画面転送の間も音程を維持します。

```bash
PATH="/opt/homebrew/opt/go/bin:$PATH" \
GOROOT=/opt/homebrew/opt/go/libexec GOTOOLCHAIN=local \
tinygo build -target=wioterminal -tags=buzzer,perf \
  -gc=conservative -opt=2 -stack-size=16KB \
  -scheduler=none -panic=trap -nobounds \
  -o tinygo-nes-buzzer.uf2 .
```

1. `tinygo-nes-buzzer.uf2` をブートローダーモードのWio Terminalへコピー。
2. 起動時に短い確認音が2音鳴ることを確認（約523Hz → 約784Hz）。
3. ゲームを開始し、BGMの鳴る場面へ進む。`tinygo monitor` では音声有効の起動メッセージと性能ログを確認可能。

`buzzer` ビルドでは、**Select（左上のBUTTON_1）を押すたびに消音／解除**を切り替えます。
長押し中は1回だけ切り替わります。ゲーム側にもSelect入力が届くため、ゲームのSelect操作と同時に音声も切り替わります。
消音中もAPUと曲の進行は続き、解除するとその時点の音から再開します。再起動すると音声は有効に戻ります。
`tinygo monitor` には `audio: muted` / `audio: unmuted` を表示します。

APUのパルス音源1を優先し、無音ならパルス音源2、それも無音なら三角波の音程を使います。
音量・発音可否は有効フラグ、長さカウンター、エンベロープなどから判断し、波形の瞬間値では判断しません。
音程と音量は各エミュレーションフレーム後に更新します。PWMの最大High幅は周期の1/8に抑えています。

これは**単音の試験出力**です。和音、元のデューティ比・音色、ノイズ、DMC/PCMは再現しません。
三角波もブザー用の矩形波に置き換わります。ゲーム速度が約13FPSの場合、曲の進行も通常より遅くなり、
短い音やフレーム途中の音量変化は取りこぼす場合があります。音程自体はPWMで設定した周波数です。
高い音、低い音、音量によって内蔵ブザーでの聞こえ方は異なります。
起動音だけ鳴ってゲーム音が出ない場合は、BGM開始後かどうかとROM側の音源使用状況を確認します。

`buzzer` タグを外すと音声処理は無効です。外付けDAC、外部スピーカー、PCの音声出力は今回の対象外です。
音声用の割り込みやgoroutineは追加していません。実装は `audio_buzzer.go` と `nes/buzzer.go`、
無効時の処理は `audio_silent.go`。音程・休符・音源選択と、APU状態を変更しないことをテストしています。
実機では起動確認音と、下記の修正後のドラクエのBGMの発音を確認済みです。処理速度に伴うテンポの遅さは残っています。

### ドラクエでBGMが無音だった問題の修正

起動音は鳴る一方、タイトル・ゲーム開始後のBGMが鳴らない状態を調査しました。
APUの `$4015` 読み出しにフレームIRQの状態ビット（bit 6）がなく、CPUが割り込みを禁止している間は
フレームIRQ要求も失われていました。手元のROMをホストで実行すると、音源の初期化後に音程・音量の設定が進まず、発音データが出ない状態を再現しました。

- フレームIRQを `frameIRQPending` に保持し、CPUのIフラグに関係なく状態を記録。
- `$4015` 読み出しでbit 6を返し、その読み出しでフレームIRQをクリア。
- `$4017` のIRQ禁止指定で保持しているフレームIRQをクリア。
- CPUの割り込み禁止解除後に保留中のAPU IRQを処理。NMIを優先し、IRQハンドラーへの移行だけではAPU側の要求をクリアしない。

レジスタの確認仕様は [NESdevのAPU Status](https://www.nesdev.org/wiki/APU_Status) を参照。
修正後のホスト実行では、タイトル相当の区間で392Hz・440Hz、Start入力後の区間で658Hzなどの発音データを確認しました。
修正後、実機でもBGMが聞こえることを確認しました。

`nes/apu_irq_test.go` にはIRQ状態の保持・読み出しクリア・禁止・NMI優先のテストと、
CPUのIRQを禁止したまま `$4015` をポーリングし、440Hzの音程を設定する自作プログラムの回帰テストを追加しました。
ROMを使った一時的な調査コードは削除し、リポジトリのテストはROMなしで実行できます。

## 速度優先版を試す

`speed` タグを追加すると、画面描画・LCD転送を通常の2フレームに1回から**4フレームに1回**へ減らします。
ゲーム内部のCPU/APU、ボタン入力、ブザー更新は毎フレーム実行し、表示を省略するフレームでもPPUの状態は進みます。
ゲームとBGMの進行を速める試験設定ですが、画面更新は粗くなります。60FPSを保証するものではありません。

```bash
PATH="/opt/homebrew/opt/go/bin:$PATH" \
GOROOT=/opt/homebrew/opt/go/libexec GOTOOLCHAIN=local \
tinygo build -target=wioterminal -tags=buzzer,perf,speed \
  -gc=conservative -opt=2 -stack-size=16KB \
  -scheduler=none -panic=trap -nobounds \
  -o tinygo-nes-speed.uf2 .
```

`tinygo-nes-speed.uf2` を書き込み、同じ場面で通常版 `tinygo-nes-buzzer.uf2` と比較します。
Selectでの消音／解除も使えます。`perf frames=120 displayed=30` になり、表示FPSは実行FPSの約1/4です。
比較するのはゲームの進行速度を示す `emu_fps_x10` と `window_ms` です。
`perf_core` は通常版と同じく、描画省略／描画ありの両方を各区間1回ずつ計測します。
通常版のUF2を書き戻すと、2フレームに1回の表示へ戻せます。

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
通常版は画面更新を2フレームに1回にしているため、画面更新FPSは実行FPSの約半分です。`speed` 版では約1/4です。

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
- この段階で残った主な負荷はCPU＋APU（実機で約50ms／フレーム）。描画省略の実装時点では、CPU命令実行とAPUの処理方式自体は変更していない。
- 下記スペックのメモリ数値は以前の概算であり、今回の変更後の使用量は再計測していない。

再確認用コマンド（ROMやGUIは不要）:

```bash
go test nes_loader.go nes_loader_test.go
go test ./nes
go test -tags=perf ./nes
go test -tags=perf performance_enabled.go performance_test.go
```

### 追加改善：APUタイマーのまとめ処理

CPU＋APUの負荷を減らすため、CPU命令1回分のサイクルを `APU.StepCycles()` でまとめて進める処理を追加しました。
CPU命令の実行順序やレジスタ書き込みのタイミングは変更せず、命令と命令の間でAPUを進めます。

- Pulse / Triangleのタイマーは残り時間と期限到達回数を計算し、毎サイクルの確認を削減。
- Noiseは期限に達した回数だけシフトレジスターを更新。
- DMCは読み出し・シフト・CPU停止の順序を維持し、イベントのないカウントダウンをまとめて減算。
- APUフレームカウンターの境界で処理を分割し、エンベロープ・長さカウンター・スイープ・IRQの更新順序を維持。
- 通常実行と詳細計測の双方に適用。従来の `APU.Step()` を比較用として残し、差分テストで挙動を検証。

変更箇所は `nes/apu_batch.go`、`nes/console.go`、`nes/frame_profile.go`。
テストは `nes/apu_batch_test.go` に追加しました。ランダムなレジスタ書き込み、ゼロ／最大タイマー値、
APUサイクルカウンターの周回、複数のフレームカウンター境界、DMCのアドレス周回・ループ・CPU停止、
実際にAPUレジスタを書き換える自作プログラムを含め、従来処理との状態・出力一致を確認しています。
これは既存APU実装との同等性確認であり、NES実機との音声精度を新たに保証するものではありません。

Apple M1上のGoベンチマーク（CPU命令相当の2〜7サイクル／呼び出し、各0.5秒）:

| APU条件 | 従来のサイクル処理 | まとめ処理 |
|---|---:|---:|
| 初期状態 | 16.45ns/op | 9.20ns/op |
| 音源タイマー設定あり | 16.01ns/op | 7.08ns/op |
| DMC有効 | 19.97ns/op | 10.32ns/op |

この数値はホスト上のAPU単体の比較です。追加で提供されたWio Terminalのログでは、次の小幅な改善が見られました。

| 指標 | APUまとめ処理前 | まとめ処理後 |
|---|---:|---:|
| 120フレームの経過時間（提供された全体ログ各1区間） | 9,633ms | 9,484ms |
| 実行FPS（経過時間から算出） | 約12.46 | 約12.65 |
| CPU＋APU（詳細サンプル各4件の平均） | 約49.83ms | 約48.71ms |
| 本体平均・詳細計測を除く（各2区間） | 64.92〜65.66ms | 63.69〜64.35ms |

実行速度の差は約1.6%、CPU＋APUの時間差は約2.3%でした。
PC上でのAPU単体の改善率が、そのまま実機全体の改善率になるわけではありません。
提供区間数は少なく、場面差や計測の揺れも含むため、改善の再現性は追加計測で確認します。
計測ログの形式は変更していません。音声の周期的なDAC出力は引き続き未対応です。

```bash
go test ./nes -run 'Test(APUBatch|AdvanceTimer)'
go test ./nes -run '^$' -bench=BenchmarkAPUStepCycles -benchtime=500ms
```

### CPUのメモリアクセスを高速化

ドラクエのホスト実行を計測し、頻繁に呼ばれるCPUのメモリ読み出しを改善しました。

- RAMとカートリッジへのアクセスを先に判定し、PPU・APU・コントローラーのレジスタ判定を別関数へ分離。
- スタックの読み書きは、常に内部RAMの `$0100〜$01FF` に収まることを使って直接アクセス。
- ゼロページの間接アドレス読み出しもRAMへ直接アクセス。`$FF → $00` の折り返しとページ跨ぎの追加サイクルを維持。

変更は `nes/cpu.go`。APUの更新頻度、フレームスキップ率、LCD転送、ボタンでの消音切り替えは従来どおりです。
`nes/cpu_memory_test.go` でRAMミラー、スタックの折り返し、間接アドレス、PPU/APUレジスタの副作用、
コントローラー、DMA、PRGバンク切り替え・境界読み出しを確認しています。

Apple M1・Goの一時的なROMベンチマーク（180フレーム実行後、隔フレーム描画、各2秒×3回、変更前後を順番に実行）:

| 指標 | 変更前 | 変更後 |
|---|---:|---:|
| 1フレームのコア処理時間・中央値 | 317.5µs | 296.6µs |

ホスト上では処理時間を約6.6%短縮しました。LCD転送やブザー出力は含まない値で、Wio TerminalのFPS改善率は実機で確認します。
ROMと一時的な計測コードはリポジトリに追加していません。

実機では `tinygo-nes-buzzer.uf2`（`buzzer,perf`）を書き込み、同じ場面・同じ音声設定で、
起動直後を除いた `perf` / `perf_core` ログを3区間以上比較します。
`emu_us/frame` と `cpu_apu_us` の短縮、`window_ms` と `emu_fps_x10` の改善を確認します。
BGM修正前のログとはゲームの進行状況が異なる可能性があるため、比較対象は音声修正後のビルドに揃えます。

```bash
go test -tags=buzzer,perf ./nes
go test ./nes -run '^$' -bench=BenchmarkFrameOutput -benchtime=1s -count=3
```

### Mapper 0・2・3のCPU読み出しを高速化

前項の変更後、実機ログで120フレームが8,753ms（約13.7FPS）、CPU＋APUの詳細サンプル平均が約42.8msとなりました。
この状態を起点に、ROM読み出しのマッパー呼び出しを調査しました。

CPUの `$6000〜$FFFF` 読み出しでは、Mapper 0・2（`Mapper2`）とMapper 3（`Mapper3`）を具体的な型で判定し、
CPU専用の `readCPU` を直接呼びます。汎用の間接呼び出しと、CPUには不要なCHR領域の判定を減らします。
PRGバンク番号は毎回現在値を参照するため、書き込み直後のバンク切り替えも反映されます。
他のマッパーは従来の `Mapper.Read` を使い、PPU側のCHR読み出しと全マッパーの書き込み処理は維持しています。

Apple M1上のドラクエによる比較（前項と同じ180フレーム後・隔フレーム描画、各2秒×3回の中央値）:

| 試行 | コア処理時間／フレーム |
|---|---:|
| 今回の変更前 | 295.5µs |
| CPU専用読み出し | 285.2µs |
| 専用読み出し＋PRGバンクのスライスをキャッシュ | 289.0µs |

CPU専用読み出しで約3.5%短縮しました。キャッシュの追加は改善が小さくなったため採用していません。
これはホスト上の結果です。実機でのFPSと `cpu_apu_us` の変化は、新しい `tinygo-nes-buzzer.uf2` で確認します。

`nes/mapper_cpu_test.go` では、16/32/64KB PRGと複数のバンク指定について `$6000〜$FFFF` の全アドレスを
汎用読み出しと比較しています。また、バンク切り替えを繰り返す自作プログラムで、フレームごとのCPU/APU/PPU・
メモリ・画面出力を従来経路と比較しています。手元のドラクエでもStart入力を含む1,200フレームの
画面・RAM・SRAM・CPUレジスタ・サイクル数・ブザー音程のハッシュが変更前後で一致しました。
一時的なROM計測コードとROMはリポジトリには追加していません。

### PPUの背景描画を高速化

描画するフレームのPPU処理が実機で約31〜33msかかっているため、`renderScanlineBackground` の
ピクセルごとの処理を削減しました。

- 16バイトの変換表でタイル1行の2枚のビットプレーンをまとめ、各ピクセルを固定のシフトで取り出す。
- 背景パレットの16色をスキャンラインごとに解決し、透明色を共通背景色へ統一。
- 左端8ピクセルの表示マスクを行末にまとめて適用し、全256ピクセルでの条件判定を削減。
- フレームバッファの行を参照し、ピクセルごとの行位置計算を削減。

背景の不透明度を保存する `backgroundColors`、スクロール更新、マッパーの読み出し順序は維持しています。
パレットは行ごとに更新するため、CPUが行の間に書き換えた色も反映されます。
描画頻度、CPU/APUの実行、ブザー出力、LCD転送の設定は変更していません。

Apple M1上のGoベンチマーク（各3回の中央値）:

| 指標 | 変更前 | 変更後 |
|---|---:|---:|
| 背景1行・fine X=0（各1秒） | 834.9ns | 598.7ns |
| 背景1行・fine X=7（各1秒） | 827.9ns | 615.2ns |
| ドラクエ1フレーム・隔フレーム描画（各2秒） | 285.6µs | 268.7µs |

ホスト上では背景描画単体を約26〜28%、ROMのコア処理全体を約5.9%短縮しました。
実機での効果は更新した `tinygo-nes-buzzer.uf2` の `frame_output=1` の `ppu_us`、
`emu_us/frame`、`window_ms` を同じ場面で比較して確認します。

`nes/ppu_background_test.go` に従来の背景描画を比較用として保持し、対応する全マッパー、
ミラー設定、fine Xの全8値、背景パターンテーブル、表示マスク、パレット変更、ネームテーブル境界で
画面・スクロール状態・スプライト合成が一致することを確認しています。
手元のドラクエでもStart入力を含む1,200フレームの画面・RAM・SRAM・CPU状態・音程のハッシュが一致しました。
ROMと一時的な計測コードはリポジトリには追加していません。

```bash
go test -tags=buzzer,perf ./nes
go test ./nes -run '^$' -bench=BenchmarkBackgroundRenderer -benchtime=1s -count=3
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
| 音声 | 内蔵ブザーの単音PWM出力（`buzzer` タグ、実験的） |

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
wioterminal.go   - Wio Terminal固有 (ILI9341, ボタン)
audio_buzzer.go  - 内蔵ブザーのPWM出力（buzzerタグ）
audio_silent.go  - 音声無効時の処理
generic.go       - PC用ビルド (tinydisplay)
nes/             - NESエミュレーターコア (fogleman/nes フォーク)
  cpu.go         - 6502 CPU (256-case switch)
  ppu.go         - PPU (スキャンラインレンダラー)
  apu.go         - APU (フレームカウンター + IRQ)
  buzzer.go      - APU状態からブザー用の音程・音量を取得
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
