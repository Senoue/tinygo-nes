# ブラウザ版

NESコアを標準Goの `js/wasm` 向けにビルドし、256×240のCanvasへ表示します。
Wio Terminal向けのエントリポイント・LCDドライバーはこのビルドには入りません。

リポジトリのルートで実行:

```sh
sh web/build.sh
python3 -m http.server 8080 --directory web --bind 127.0.0.1
```

ブラウザで http://localhost:8080 を開き、手元の `.nes` ファイルを選択します。
HTMLを `file://` で直接開く方法は使えません。ROMをビルドへ埋め込む必要はなく、
選択したROMのデータはサーバーへ送信されません。

- 矢印キー: 十字キー
- Z / X: A / B
- Enter: Start、右Shift: Select
- 一時停止・再開、リセットに対応。フォーカスを失うと一時停止します。
- 60フレーム/秒を目標に実行します。遅い端末では実行速度が落ちます。
- 音声・バッテリーバックアップの保存は未実装。リセット・再読み込みでゲーム状態は失われます。
- NES 2.0は非対応。対応マッパーは既存のNESコアに準じます。

`main.wasm` と `wasm_exec.js` は生成物です。Goを更新した場合は両方を再生成してください。
ビルド方法は [Go WebAssembly](https://go.dev/wiki/WebAssembly) に準じています。

自作の最小ROMによるWASM実行テスト（Node.jsが必要）:

```sh
node web/smoke-test.cjs
```
