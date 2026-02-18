# signetの仕様

## ファイル

### 設定: /etc/signet/signet.conf

toml形式

設定可能項目
- RootDir: ファイル類のルートディレクトリ(デフォルト: /etc/signet)

### 秘密鍵: /etc/signet/ed25519.priv

### ノード: /etc/signet/nodes/{node_name}

```txt
NickName = 田中
Address = 10.x.x.x
Ed25519PublicKey = xxxxxxxx
```

### ブロック: /etc/signet/block.jsonl

信頼されているであろうブロックが置かれるファイル

### 承認待ち取引: /etc/signet/pending_transaction.json

## コマンドライン上での操作
- signet init: 初期化
    - --address: 自分のアドレス
    - --nickname: ニックネーム
    - --nodename: ノード名
- signet start: HTTPサーバを起動する
- signet stop: HTTPサーバを停止する

## HTTP JSON API エンドポイント

### POST /transaction/propose
Fromが取引を提案。ToのノードにFrom署名付きトランザクションを送る
### POST /transaction/approve
Toが承認。自分の署名を追加してブロック生成＆ブロードキャスト
### GET /transaction/pending
自分宛の未承認トランザクション一覧を確認
### POST /register
ユーザー登録（registerタイプのトランザクション）
### GET /chain
チェーン全体の取得
### POST /block
他ノードからのブロック受信
### GET /peers
ノードリスト取得

## エンティティ

### BlockHeader

- created_at: 作成日時
- prev_hash: 前ブロックのハッシュ
- hash: このブロックのハッシュ

### BlockPayload

#### Transaction

- from(string): ノード名
- to(string): ノード名
- amount(integer): 金額
- title(string): 題名

#### AddNode

- public_key(string): 公開鍵
- node_name(string): ノード名
- nick_name(string): ニックネーム
- address(string): 宛先アドレス
