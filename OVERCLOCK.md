# Wio Terminal 150 MHz 実験

**通常120 MHz・音声＋計測版は実機で動作確認済み。OC版は未確認。**
初版/r2はHomebrewのラッパーがTINYGOROOTを上書きしており、OC未反映だった。
これらは周波数比較に使用しない。以下の設定はラッパー対策後のr4に対するもの。スタック指定の訂正は末尾参照。

TinyGo 0.43.0向け。通常の120 MHzビルドは変更しません。
インストール済みTinyGoの代わりに、一時ディレクトリに必要なソースだけを
差し替えたTINYGOROOTを作り、ビルド後に破棄します。

```sh
python3 tools/build-oc.py 120
python3 tools/build-oc.py 150
```

`roms/game.nes` が必要です。生成物:

- `tinygo-nes-oc120-r4.uf2`: 比較用120 MHz
- `tinygo-nes-oc150-r4.uf2`: 実験用150 MHz

どちらも `buzzer,perf,ocbench`、LCD SPI 24 MHz、Flash待ち7サイクルです。
`ocbench` タグだけではCPUは高速化しません。必ず上のスクリプトを使ってください。
以前の通常版はSPI実効60 MHzであり、この比較用120 MHz版とは条件が違います。

## 変更内容

- DPLL0の2 MHz入力に対する倍率を60倍／75倍（LDR=59／74）に設定。
- `machine.CPUFrequency()` とSPIのGCLK0周波数定数も実際の設定に合わせる。
- TCC PWMの `period * 3 / 25`（120 MHz前提）を周波数に合わせて変更。
  150 MHz版では `period * 150 / 1000`。ブザーの音程が1.25倍になるのを防ぐ。
- Flashは自動待ち設定を解除してRWS=7に固定。両版で同じ余裕を持たせた実験条件。
  この設定は実機安定性の保証ではなく、待ち時間増加による性能への影響もある。
- LCDは独立した48 MHz GCLK1の分周により24 MHzに固定。
  CPUと一緒にLCDまで高速化することを避ける。
- USB用48 MHz、RTC用32768 Hz、PLL1、電源電圧の設定は変更しない。
- `time.Now()` 等はTinyGoのRTCを使うため、CPU周波数変更で時刻が1.25倍にはならない。
- LCD初期化前にはUSBログを出力しない。周波数はビルド時の設定値であり実測ではない。

## 実機での比較

1. 復帰用として既存の通常UF2を保管する。
2. まずoc120をブートローダーモードでコピーし、表示、入力、USBログ、起動音とBGMを確認。
3. 同じROM・場面・音声状態で、起動直後を除く `perf` / `perf_core` を3区間以上記録。
4. oc150で同じ操作を行い、`window_ms`、`emu_us/frame`、`cpu_apu_us`、`ppu_us` を比較。
5. 10分程度動かし、画面乱れ・フリーズ・リセット・USB切断・音程変化を確認。
   問題が出たらブートローダーから通常UF2へ戻す。

ビルド時に自動フラッシュはしません。120 MHz超は実験扱いで、個体・温度などによる
安定性の差があります。CPUクロック倍率1.25倍は、ゲーム全体のFPS倍率の保証ではありません。

## 参考

- [Seeed仕様: 120 MHz / Boost up to 200 MHz](https://wiki.seeedstudio.com/Wio-Terminal-Getting-Started/)
- [Seeed SAMD起動処理](https://github.com/Seeed-Studio/ArduinoCore-samd/blob/master/cores/arduino/startup.c)
- [SAM D5x/E5xデータシート](https://files.seeedstudio.com/wiki/Wio-Terminal/res/ATSAMD51.pdf)

ローカルTinyGo 0.43.0の `runtime_atsamd51.go` と `machine_atsamd51.go` を確認して実装。
他のバージョンではスクリプトが停止します。ソースの置換箇所が一致しない場合も停止します。

## r2: 両方の実験版で画面が出ない報告への対処

初版はAUTOWSとRWSをクリアしてから別の書き込みでRWS=7にしていました。
この中間状態（自動待ち無効・RWS=0）を作らず、1回のレジスター書き込みで設定するよう修正。
LCDより前に実行されるUSBログのinit関数も削除しました。原因確定・実機での改善確認は未完了です。
初版と区別するためUF2名に `-r2` を付けています。

まず `tinygo-nes-recovery120.uf2`（通常ランタイム・従来のLCD速度）で表示を確認し、
その後 `tinygo-nes-oc120-r2.uf2` を試してください。120 MHzで表示できるまで150 MHzは試さないでください。

## r2の120 MHz版も表示されない場合の切り分け

実機から `oc120-r2` でも表示されない報告あり。初期化修正だけでは未解決。
次は `tinygo-nes-diag120-spi24.uf2` を使う。この版はインストール済みTinyGoを
そのまま使用し、復帰用通常版との差はLCD SPI指定が24 MHzになる点のみ。
Flash待ち、PLL、PWMのパッチや実験用TINYGOROOTは使わない。

```sh
tinygo build -target=wioterminal -tags=buzzer,perf,ocbench \
  -gc=conservative -opt=2 -stack-size=16KB -scheduler=none \
  -panic=trap -nobounds -o tinygo-nes-diag120-spi24.uf2 .
```

- 復帰用通常版で表示でき、これでも表示できる場合: 実験用ランタイムの変更側を調べる。
- 復帰用通常版では表示でき、これでは表示できない場合: SPIの速度・クロック源切り替え側を調べる。
- 復帰用通常版も表示できない場合: OC以外のROM・書き込み・起動状態から確認する。

原因確定まで150 MHz版での比較は保留する。

## 通常版も白画面になる報告と起動診断

通常版も白いとの報告あり。白いバックライトだけではLCDの初期化成功を断定できない。
`tinygo-nes-bootdiag.uf2` は通常クロック・従来LCD設定で、音声初期化とUSBログを使わない。
0.7秒ずつ段階色を表示し、最後に1行バッファでゲーム画面を転送する。

| 色 | 到達した処理 |
|---|---|
| 青 | mainに到達、LCDのinitから戻った |
| 緑 | ROM読み込み・コンソール作成完了 |
| 黄 | 最初のフレームを実行する直前 |
| 紫 | 最初のフレーム実行から戻った |
| 赤 | ROMローダーがエラーを返した |

白いままなら、LCD通信・mainより前の初期化も候補に残る。
色が変わるがゲーム画面が出ない場合は、最後に見えた色で切り分ける。

```sh
tinygo build -target=wioterminal -tags=bootdiag \
  -gc=conservative -opt=2 -stack-size=16KB -scheduler=none \
  -panic=trap -o tinygo-nes-bootdiag.uf2 .
```

r3ビルドはHomebrewのlibexec内の実行ファイルを呼び、TINYGOROOTの反映を確認する。
TinyGoが標準ライブラリを合成できるよう、src以下のディレクトリは実ディレクトリで作る。

### 実機確認: bootdiagはゲーム画面まで到達

ユーザー確認で青〜紫を経てゲーム画面を表示できた。
この構成ではROMロード・フレーム生成・LCD描画が動作する。
通常版との差（音声、性能計測、24行バッファ、描画頻度など）の切り分けを続ける。

次の `tinygo-nes-diag120-normal-silent.uf2` は通常main・24行バッファ・隔フレーム描画に戻し、
音声と性能計測を外した通常120 MHz版。OCパッチは使わない。

```sh
tinygo build -target=wioterminal -gc=conservative -opt=2 \
  -stack-size=16KB -scheduler=none -panic=trap -nobounds \
  -o tinygo-nes-diag120-normal-silent.uf2 .
```

これで表示できれば音声と性能計測を一つずつ戻す。
白い場合はバッファ・スタック使用量など通常main側の違いを優先して調べる。

### 実機確認: 通常描画・音声なし・計測なし版でプレー可能

`diag120-normal-silent` はユーザー確認で通常プレー可能。
通常の24行バッファと隔フレーム描画だけで白画面になるわけではない。
次は同じ設定で以下を個別に試す（両版ともOCなし）。

- `tinygo-nes-diag120-audio-only.uf2`: `-tags=buzzer`、性能計測なし。
- `tinygo-nes-diag120-perf-only.uf2`: `-tags=perf`、音声なし。

共通ビルドオプションは `-target=wioterminal -gc=conservative -opt=2
-stack-size=16KB -scheduler=none -panic=trap -nobounds`。
音声のみ版では起動音・ゲーム表示・BGMを確認。
計測のみ版ではゲーム表示と、USBモニター接続の有無による違いを確認。
両方単独で動けば、併用時のメモリ使用量や相互作用も候補に残る。

### 実機確認: 音声のみ・性能計測のみは両方動作

ユーザー確認で `diag120-audio-only` と `diag120-perf-only` はともに動作。
音声または計測が単独で必ず停止する状況ではない。

次は併用時のスタック余裕を比較する `tinygo-nes-diag120-combined-stack32.uf2`。
通常120 MHz・従来LCD設定・`buzzer,perf`を使用し、復帰用通常版に対する
ビルド設定の差は `-stack-size=32KB`（従来16KB）のみ。
これは仮説の検証であり、スタック不足の確定ではない。
静的なスタック予約を増やす分、ヒープに使えるRAMは減る。

```sh
tinygo build -target=wioterminal -tags=buzzer,perf \
  -gc=conservative -opt=2 -stack-size=32KB -scheduler=none \
  -panic=trap -nobounds -o tinygo-nes-diag120-combined-stack32.uf2 .
```

画面・起動音・BGMを確認し、USBモニターを接続した場合にも
`perf` / `perf_core` が複数区間出てゲームが止まらないことを確認する。

### 実機確認: 音声＋計測・スタック32 KB版は動作

ユーザー確認で `diag120-combined-stack32` は問題なく動作。
併用16 KB版との差からスタック余裕が関係する可能性があるが、原因確定ではない。
OCビルドスクリプトも32 KBに揃え、生成物をr4とする。

次は `oc120-r4` を先に試す。これは既に動いた通常版とはFlash待ち設定と
LCDクロックが異なるので、150 MHz版より先に確認が必要。
それが動いたら `oc150-r4` を試し、同じ場面・同じUSB接続状態で
`perf` / `perf_core` を3区間以上比較する。両版とも実機周波数・安定性は未検証。

### 訂正: stack-size指定による改善は確認できていない

生成物を検証すると `recovery120` と `diag120-combined-stack32` はSHA-256まで同一だった。
r3/r4も各周波数で同一。従って「スタックを増やして改善した」とは結論できない。
このscheduler=none構成ではCLIのstack-size変更は生成物に反映されなかった。
ターゲットの `atsamd51.ld` は `_stack_size = 4K` を指定している。
前段の32 KBに関する説明はCLIに渡した値であり、実際のメインスタック予約量ではない。
通常版の表示が以前と異なった理由は未確定（起動・接続条件なども要確認）。

一方r4の120 MHzと150 MHzのUF2はハッシュが異なり、実験用TINYGOROOTの反映も
スクリプト内で検証している。周波数の実測と実機動作確認は引き続き必要。

### 実機確認: r4は両周波数ともBGMあり・画面は白

CPU/音声処理は進むが、120 MHz・150 MHzともゲーム画面は出ないとの報告。
LCD 24 MHz指定（GCLK1選択）と通常の80 MHz指定（GCLK0選択）の差を調べる。
Flash待ち設定やランタイムの他の変更はこの試行では維持する。

```sh
python3 tools/build-oc.py 120 --lcd-stock
python3 tools/build-oc.py 150 --lcd-stock
```

生成物は `tinygo-nes-oc120-r5-lcd-stock.uf2` と `tinygo-nes-oc150-r5-lcd-stock.uf2`。
まず120 MHz版を試す。r4との差はLCDの指定を通常設定に戻す点のみ。
120 MHzで表示できたら150 MHzを試す。
SPIの計算上の実効速度は120 MHz版で60 MHz、150 MHz版で75 MHzとなる。
両版のSPI速度は同一ではないため、FPS差をCPUだけの効果とは解釈しない。
実測・表示改善の確認は未完了。120 MHz版でも白いならFlash/ランタイム側の切り分けを続ける。

## r6: explicit LCD transmit completion

The project now uses a local copy of drivers v0.34.0's ILI9341 package in
`internal/ili9341` (upstream license included). All six ATSAMD51 SPI write
methods wait for TXC before returning, before the caller can change DC/CS.
Zero-length writes return immediately. The previous code waited for DRE
before each byte but did not explicitly wait for the last bit to leave the
shifter. This is a candidate explanation for the 24 MHz white screen, not
an experimentally confirmed cause. DMA is not included in this revision.

Comparison builds (CPU 150 MHz, same flags as r5):

- `python3 tools/build-oc.py 150 --lcd-stock`: `tinygo-nes-oc150-r6-txc-lcd-stock.uf2`, expected SPI 75 MHz.
- `python3 tools/build-oc.py 150`: `tinygo-nes-oc150-r6-txc-lcd24.uf2`, expected SPI 24 MHz.

実機確認: ユーザーから75 MHz版・24 MHz版とも問題なさそうとの報告あり。
TXC待ちの不足が白画面の原因だった可能性は高まったが、原因の断定はしない。

## r7: synchronous LCD DMA experiment

r6 was reported working on hardware at both 75 MHz and 24 MHz LCD SPI.
The `lcddma` build tag now uses DMAC channel 0 for bulk byte slices on
Wio Terminal SPI3 (SERCOM7 TX trigger 19). Commands and startup fills retain
polling. Descriptor and writeback tables are statically allocated, aligned
to 16 bytes, with byte beats, source increment and fixed SPI DATA destination.
Transfers larger than 65535 bytes are split. Completion is polled with a
100 ms per-block timeout and transfer-error detection; SPI TXC is still
waited before changing CS/DC. The source buffer stays alive until completion.
The experiment requires exclusive ownership of DMAC and refuses to replace
an already enabled controller's tables.

Build: `python3 tools/build-oc.py 150 --lcd-stock --lcd-dma`
Output: `tinygo-nes-oc150-r7-dma-lcd-stock.uf2`.
Optional slower SPI diagnosis: omit `--lcd-stock`.
Omit `--lcd-dma` to rebuild r6's polling implementation.

This revision preserves the 24-line buffer and every-other-frame display.
It does not yet overlap conversion/emulation with DMA; the CPU polls block
completion. Compare `transfer_us/display`, `convert_us/display`, and FPS
with r6 in the same game scene. Initial user-provided measurements are
recorded below; extended stability testing remains pending.

Register setup reference: SAM D5x/E5x data sheet, DMAC chapter;
SERCOM7 request IDs verified against
https://raw.githubusercontent.com/adafruit/asf4/master/samd51/include/instance/sercom7.h


## LCD転送改修まとめ・実測結果（2026-10-20）

### 変更内容

1. **r6: 送信完了待ちを修正**。ILI9341ドライバーをプロジェクト内に配置し、
   ATSAMD51の6種類のSPI書き込み処理でTXCを待つようにした。
   DRE（次のデータを書ける状態）と実際の送信完了を区別し、最後のビットを
   送り終えてからDC・CSを変更する。空データ・非正の繰り返し回数では待たない。
   ユーザー確認ではSPI 75 MHz・24 MHzの両版が動作した。
2. **r7: 画像データの転送をDMA化**。Wio TerminalのSPI3＝SERCOM7に対し、
   DMACチャネル0でバイト単位に送信する。コマンドや起動時の塗りつぶしは
   従来のポーリングを維持。DMA完了後もTXC待ちを行う。
3. **比較条件を維持**。CPU 150 MHz、SPI計算値75 MHz、24行バッファ、
   2エミュレーションフレームに1回の表示。今回は同期式DMAであり、
   CPUは完了をポーリングする。色変換やエミュレーションとの並行処理は未実装。

主な変更ファイル:

- `internal/ili9341/spi_atsamd51.go`: TXC待ちと画像転送のDMA分岐。
- `internal/ili9341/dma_wioterminal.go`: DMA初期化、記述子、転送、エラー検出。
- `internal/ili9341/dma_disabled.go`: DMAを使わないビルドの経路。
- `wioterminal.go`: ローカルILI9341ドライバーの参照。
- `tools/build-oc.py`: `--lcd-dma`とr7出力名を追加。

GoモジュールキャッシュやTinyGoのインストール本体は変更していない。
ドライバーの由来・ライセンスは `internal/ili9341/README.md` と `LICENSE` に記録。

### ユーザー提供の測定値

各区間は120エミュレーションフレーム・60表示。
以前の150 MHz版のログとの比較であり、同一場面の厳密なA/B測定ではない。
特にエミュレーション時間・FPSの差をすべてDMAの効果とは扱わない。

| 指標 | 以前の150 MHz版（2区間） | r7 DMA版（2区間） |
| --- | --- | --- |
| 計測窓 ms | 6859 / 6567 | 5971 / 6484 |
| emu_fps_x10 | 174 / 182 | 200 / 185 |
| display_fps_x10 | 87 / 91 | 100 / 92 |
| emu_us/frame | 44593 / 42255 | 40375 / 44312 |
| convert_us/display | 3031 / 3029 | 3369 / 3363 |
| transfer_us/display | 19411 / 19421 | 14841 / 14857 |

LCD転送は平均 **19.416 ms → 14.849 ms**、表示1回あたり **4.567 ms（約23.5%）短縮**。
RGB565の256×240画素＝122,880バイトを75 MHzで送る理論時間は約13.107 ms。
実測との差は約1.742 msまで縮まった。ただしSPI周波数の実測値ではなく計算値を使用。

色変換は平均3.030 msから3.366 msへ増加しており、理由は未確定。
色変換と転送の合計は22.446 msから18.215 msへ、表示1回あたり約4.231 ms短縮した。
2フレームに1回表示するため、エミュレーション1フレームあたりの削減相当は約2.116 ms。

r7の詳細サンプル:

| 区間 | frame_output | cpu_apu_us | ppu_us | total_us |
| --- | --- | --- | --- | --- |
| 1 | 0 | 35400 | 2807 | 38696 |
| 1 | 1 | 35522 | 19897 | 56030 |
| 2 | 0 | 34667 | 3662 | 38818 |
| 2 | 1 | 35156 | 20263 | 56030 |

`perf_core`は各1サンプルで、通常窓の平均とは別の測定。
継続した性能ログの取得は確認できたが、r7の表示品質・BGM・長時間安定性は
別途確認が必要。DMA版・DMA無効版のビルドと `git diff --check` は成功済み。

### 再現方法と次の候補

```sh
# r7: CPU 150 MHz / LCD SPI計算値75 MHz / DMA
python3 tools/build-oc.py 150 --lcd-stock --lcd-dma

# r6: 比較・復帰用のポーリング版
python3 tools/build-oc.py 150 --lcd-stock
```

次の候補は24行バッファを2つにし、DMA送信と次のバッファの色変換を重ねること。
バッファ再利用前の完了待ちと、最終転送のTXC待ちを維持する必要がある。
現在はエミュレーション本体が約40〜44 ms/フレームを占めるため、
さらに大きな改善を狙うにはCPU/APU・PPU処理の最適化も必要。

DMA実装はコントローラーの記述子テーブルを専有する。今後ほかのDMA利用を
追加する場合は共有管理が必要。現在は既に有効なDMACを検出すると停止し、
他の利用者の設定を上書きしない。転送エラー・100 msタイムアウトも停止扱い。
