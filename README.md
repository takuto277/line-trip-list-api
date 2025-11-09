# LINE Trip List API

LINE Messaging APIのWebhookを受信し、iOSアプリとの中継を行うサーバーサイドAPIです。

## プロジェクト構成

このリポジトリには以下が含まれています：

- `webhook-server/` - Go言語で実装されたWebhookサーバー
- `get_group_id.sh` - LINE Group IDを取得するためのスクリプト
- `line_api_env.sh` - LINE API環境変数設定用スクリプト

## セットアップ

### 1. 環境変数設定
```bash
cd webhook-server
cp .env.example .env
# .envファイルを編集して実際の値を設定
```

### 2. 依存関係インストール
```bash
cd webhook-server
go mod tidy
```

### 3. ローカル実行
```bash
cd webhook-server
go run main.go
```

### 4. ngrokでトンネル作成（開発用）
```bash
# 別ターミナルで
ngrok http 8080
```

## API エンドポイント

### Webhook受信
- `POST /webhook` - LINE Messaging APIからのWebhook

### メッセージ送信
- `POST /send` - iOSアプリからのメッセージ送信
```json
{
  "group_id": "GROUP_ID",
  "message": "メッセージ内容"
}
```

### ヘルスチェック
- `GET /health` - サーバー生存確認

## デプロイ

### Vercelへのデプロイ

詳細は `webhook-server/DEPLOY.md` を参照してください。

重要な設定：
1. Root Directoryを`webhook-server`に設定
2. 環境変数（`LINE_CHANNEL_SECRET`、`LINE_CHANNEL_TOKEN`）を設定
3. LINE Developer ConsoleでWebhook URLを設定

## 取得が必要な情報

### LINE Developer Console
1. **Channel Secret** - Basic settings
2. **Channel Access Token** - Messaging API settings（Issue buttonで発行）
3. **Group ID** - ボットをグループに追加後、`get_group_id.sh`を使用して取得

## 関連プロジェクト

- [line-trip-list](https://github.com/takuto277/line-trip-list) - iOSクライアントアプリケーション

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。
