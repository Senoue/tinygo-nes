# TinyGo `-scheduler=none` での埋め込みインターフェースのメソッドディスパッチ検証結果

## 概要

TinyGoの`-scheduler=none`（スケジューラ無効化）環境で、埋め込みインターフェース（embedded interface）のメソッドディスパッチが正常に動作するかを検証しました。

## テスト環境

- **TinyGo バージョン**: 0.40.1
- **Go バージョン**: 1.25.8
- **ターゲット**: wasm
- **プラットフォーム**: darwin/arm64
- **LLVM バージョン**: 20.1.1

## 検証対象のコードパターン

このプロジェクトでは、[nes/filter.go:48-57](nes/filter.go#L48-L57)で以下のような**埋め込みインターフェース**パターンを使用しています：

```go
type Filter interface {
    Step(x float32) float32
}

type FilterChain []Filter  // Filterのスライス

// FilterChain自身もFilterインターフェースを実装
func (fc FilterChain) Step(x float32) float32 {
    if fc != nil {
        for i := range fc {
            x = fc[i].Step(x)
        }
    }
    return x
}
```

このパターンは[nes/apu.go:94](nes/apu.go#L94)で使用されており、音声フィルタリング処理の重要な部分を担っています。

## テスト内容

以下の7つのシナリオをテストしました：

### 1. 基本的なインターフェースメソッド呼び出し
- 構造体のメソッドをインターフェース経由で呼び出し

### 2. 埋め込みインターフェース（FilterChain）
- `[]Filter`型のスライスが`Filter`インターフェースを実装するパターン
- 複数のフィルタをチェーン接続して順次処理

### 3. 構造体内のインターフェース
- 構造体フィールドとしてインターフェースを保持し、メソッド経由で呼び出し

### 4. 動的なインターフェース生成
- 実行時に異なる具象型を返す関数を通じたインターフェースの使用

### 5. 型アサーション
- インターフェースから具象型への型アサーション
- 成功ケースと失敗ケースの両方

### 6. ネストされたFilterChain
- FilterChainの中にFilterChainを含むネスト構造
- 3層以上の深いインターフェース呼び出し

### 7. Nil インターフェース処理
- nilのFilterChainに対するStep()呼び出し

## テスト結果

### ビルド結果

両方のスケジューラオプションで**正常にビルド成功**：

```bash
✓ tinygo build -target=wasm test_interface_advanced.go
✓ tinygo build -target=wasm -scheduler=none test_interface_advanced.go
```

### 実行結果

**デフォルトスケジューラ** と **`-scheduler=none`** の両方で、すべてのテストが**完全に同一の出力**を生成：

```
=== TinyGo Interface Dispatch Advanced Tests ===

Test 1: Basic interface method call
Result: 50.000000

Test 2: Embedded interface (FilterChain)
Result: 30.000002

Test 3: Interface stored in struct
Result: 100.000000

Test 4: Dynamic interface creation
Result: 50.000000

Test 5: Type assertion
Type assertion successful: B0=0.500000
Type assertion failed

Test 6: Nested FilterChain
Result: 80.000000

Test 7: Nil FilterChain
Result: 100.000000

All tests completed successfully!
```

### バイナリサイズ比較

| スケジューラオプション | ファイルサイズ | 削減率 |
|---------------------|------------|--------|
| デフォルト           | 603 KB     | -      |
| `-scheduler=none`   | 498 KB     | 17.4%  |

`-scheduler=none`を使用することで、約**105 KB (17.4%)**のサイズ削減が可能です。

## TinyGoのインターフェース実装について

### 標準Goとの違い

TinyGoは標準Goコンパイラとは異なるインターフェース実装を採用しています：

- **標準Go**: itableベースの実装
- **TinyGo**: typecode（型固有の小さな数値）ベースの実装

### TinyGoの最適化手法

TinyGoコンパイラは、インターフェースメソッド呼び出しを実装数に応じて最適化します：

1. **単一実装**: 直接呼び出しに変換（インターフェースオーバーヘッドを完全に除去）
2. **複数実装**: typecodeベースのtype switchに変換し、LLVMが効率的な機械語に最適化
3. **実装なし**: unreachableとマークし、デッドコード除去

この実装により：
- 暗黙的なメモリアロケーションを回避
- 実行時のitable検索を排除
- インターフェース境界を越えた最適化が可能
- 型アサーションが単純な定数比較に簡略化

### `-scheduler=none`の制限事項

公式ドキュメントによると、`-scheduler=none`の制限は以下の通りです：

- ❌ **goroutine** は使用不可
- ❌ **channel** は使用不可
- ✅ **インターフェースメソッドディスパッチ** は正常に動作（制限なし）

## 結論

### 主要な発見

1. **埋め込みインターフェースのメソッドディスパッチは`-scheduler=none`環境で正常に動作する**
2. すべてのテストケース（基本的な呼び出し、型アサーション、ネスト構造など）で期待通りの動作を確認
3. デフォルトスケジューラと`-scheduler=none`で実行結果に差異なし
4. バイナリサイズが約17%削減される

### NESエミュレータプロジェクトへの影響

現在のプロジェクトコード（特に[nes/filter.go](nes/filter.go)の`FilterChain`実装）は、`-scheduler=none`環境でも問題なく動作します。

ただし、[wasm/main.go:44](wasm/main.go#L44)で`select {}`を使用しているため、現状のWASMビルドで`-scheduler=none`を適用するには、このコードの修正が必要です。

### 推奨事項

- **インターフェースメソッドディスパッチに関しては、`-scheduler=none`は安全に使用可能**
- goroutineやchannelを使用していないコード部分では、`-scheduler=none`によるサイズ削減のメリットを享受できる
- TinyGoのインターフェース実装は効率的で、スケジューラの有無に関わらず正常に動作する

## テストファイル

検証に使用したテストファイル：
- [test_interface.go](test_interface.go) - 基本テスト
- [test_interface_advanced.go](test_interface_advanced.go) - 包括的なテスト
- [test_interface.html](test_interface.html) - ブラウザでの実行テスト用HTML

## 参考資料

- [TinyGo Important Build Options](https://tinygo.org/docs/reference/usage/important-options/)
- [Interfaces in TinyGo - Ayke van Laethem](https://aykevl.nl/2018/12/tinygo-interface)
- TinyGo GitHub Issues: scheduler関連の議論
