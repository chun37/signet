# Signet 開発ガイド

## プロジェクト概要

友人間の現金貸し借り記録を共有するプライベートブロックチェーン。Go フルスクラッチ、Ed25519署名、SHA-256ハッシュ、P2Pゴシップ型。

## アーキテクチャ

```
main.go          エントリーポイント (init/start/stop)
cmd/             コマンド実装
config/          設定管理 (TOML)
core/            ブロック・チェーン・トランザクション
crypto/          Ed25519 鍵生成・署名・検証
node/            全コンポーネント統合（server.NodeService 実装）
server/          HTTP API + ハンドラー
p2p/             ブロードキャスト
storage/         永続化 (JSONL/JSON/TOML)
```

## 重要な設計判断

### 型の二重構造: core.Block vs server.Block

- `core.Block`: 内部表現。`Header.CreatedAt` は `time.Time`、`Payload.Data` は `json.RawMessage`
- `server.Block`: HTTP API 表現。`Header.CreatedAt` は `int64` (Unix)、`Payload.Transaction`/`AddNode` は構造体
- 変換関数: `node/node.go` の `convertBlockToServer` / `convertServerToBlock`
- **注意**: 変換時に署名(`FromSignature`/`ToSignature`)とIndex を必ずマッピングすること。欠落するとブロードキャスト受信側でハッシュ不一致になる

### ジェネシスブロック

- 全ノード共通の固定データ (`node_name:"genesis"`, `nick_name:"Signet Network"`)
- **ノード固有情報をジェネシスに入れてはいけない** → チェーンルートが不一致になり P2P が壊れる
- ノード自身の情報は `/register` で記録する

### P2P ブロードキャスト

- `p2p.BroadcastBlock` は `any` 型を受け取る。呼び出し側（`node.go`）で `server.Block` に変換済みのものを渡す
- 受信側 `POST /block` は `server.Block` でデコード → `core.Block` に変換 → 検証・追加

### チェーン同期 (SyncChain)

- `node.Node.SyncChain()` に実装（`p2p/sync.go` ではない）
- `GET /chain` で `[]*server.Block` を取得 → `core.Block` に変換 → 最長チェーンルールで置換
- **置換後は `BlockStore.ReplaceAll()` で永続化が必須**

### トランザクション署名

- 提案時: ノードが自分の秘密鍵で自動署名（API に `from_signature` は不要）
- 承認時: To ノードが自分の秘密鍵で署名を追加

## デプロイ先

| ノード | IP | パス |
|--------|-----|------|
| node-137 | 192.168.120.137 | /root/signet |
| node-138 | 192.168.120.138 | /root/signet |

### プロセス管理 (systemd)

両サーバーに `signet.service` を設置済み（enabled / on-failure restart）。

```bash
systemctl start signet      # 起動
systemctl stop signet       # 停止（SIGTERM → graceful shutdown）
systemctl restart signet    # 再起動
systemctl status signet     # 状態確認
journalctl -u signet -f     # ログをtail
```

unitファイル: `/etc/systemd/system/signet.service`

### デプロイ手順

```bash
# 1. commit & push
git push origin master

# 2. 両サーバーで pull → ビルド → 再起動
ssh root@192.168.120.137 'cd /root/signet && systemctl stop signet; git pull && go build -o signet . && systemctl start signet'
ssh root@192.168.120.138 'cd /root/signet && systemctl stop signet; git pull && go build -o signet . && systemctl start signet'

# 3. 初期化（データリセットする場合のみ）
ssh root@192.168.120.137 'systemctl stop signet; rm -rf /etc/signet && /root/signet/signet init --address 192.168.120.137 --nickname Node137 --nodename node-137 && systemctl start signet'
ssh root@192.168.120.138 'systemctl stop signet; rm -rf /etc/signet && /root/signet/signet init --address 192.168.120.138 --nickname Node138 --nodename node-138 && systemctl start signet'

# 4. ピア相互登録（初期化後に必要）
# 137に138を登録
ssh root@192.168.120.137 'curl -s -X POST http://192.168.120.137:8080/register ...'
# 138に137を登録
ssh root@192.168.120.138 'curl -s -X POST http://192.168.120.138:8080/register ...'
```

### 動作確認

```bash
# サービス状態
ssh root@192.168.120.137 'systemctl status signet --no-pager'

# チェーン確認（サーバー内から実行）
ssh root@192.168.120.137 'curl -s http://192.168.120.137:8080/chain | python3 -m json.tool'

# ピア確認
ssh root@192.168.120.137 'curl -s http://192.168.120.137:8080/peers'

# ログ確認
ssh root@192.168.120.137 'journalctl -u signet --no-pager -n 50'
```

**注意**: ローカルから直接 `curl http://192.168.120.137:8080/...` はタイムアウトする。サーバーが特定 IP にバインドされているため、必ず `ssh` 経由でアクセスすること。

## ビルド・テスト

```bash
go build ./...     # ビルド
go test ./...      # テスト
go vet ./...       # 静的解析
make deploy-all    # 両サーバーにバイナリ転送（scp）
```

## 過去に踏んだバグ

| バグ | 原因 | 対処 |
|------|------|------|
| ブロードキャスト先でハッシュ不一致 | `server.BlockPayload` に `FromSignature`/`ToSignature` がなく変換で消失 | フィールド追加 + 変換でマッピング |
| ノード間でチェーンが同期しない | ノード固有ジェネシスでチェーンルート不一致 | 全ノード共通の固定ジェネシスに統一 |
| ダブルポート `192.168.120.137:8080:8080` | Address にポートが含まれる場合の未処理 | `ParseAddress()` でホスト/ポート分離 |
| ジェネシスブロック二重生成 | `NewNode` で `NewChain()` + 既存ブロック追加 | `NewChainFromBlocks()` で直接構築 |
| チェーン同期後の永続化漏れ | `SyncChain` が `BlockStore.ReplaceAll()` を呼んでいない | `node.SyncChain()` に移動して永続化追加 |
