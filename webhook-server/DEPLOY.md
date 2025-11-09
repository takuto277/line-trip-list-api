# LINE Trip List Webhook Server

Vercelデプロイ済み: https://line-trip-list-api.vercel.app

## 📋 デプロイ手順

### 1. Vercelプロジェクト設定

#### Root Directory設定
- Settings → General → Root Directory
- 値: `webhook-server` を設定

#### 環境変数設定
Settings → Environment Variables で以下を追加:

| 変数名 | 説明 | 取得方法 |
|--------|------|----------|
| `LINE_CHANNEL_SECRET` | LINEチャンネルのシークレット | LINE Developer Console > Basic settings |
| `LINE_CHANNEL_TOKEN` | LINEチャンネルのアクセストークン | LINE Developer Console > Messaging API > Issue token |

### 2. LINE Developer Console設定

#### Webhook URL設定
1. Messaging API settings を開く
2. Webhook settings セクション
   - **Webhook URL**: `https://line-trip-list-api.vercel.app/api/webhook`
   - **Use webhook**: ON に設定
   - **Verify** ボタンをクリックして接続確認

#### 自動返信の無効化（推奨）
1. Messaging API settings を開く
2. LINE Official Account features セクション
   - **Auto-reply messages**: OFF
   - **Greeting messages**: OFF

### 3. 動作確認

#### ステップ1: ヘルスチェック
```bash
curl https://line-trip-list-api.vercel.app/api/health
# 期待する応答: {"status":"ok"}
```

#### ステップ2: Botをグループに追加
1. LINEアプリでグループを作成
2. グループにBotを追加
3. グループでメッセージを送信

#### ステップ3: メッセージ確認
ブラウザで以下にアクセス:
```
https://line-trip-list-api.vercel.app/api/messages
```

送信したメッセージが表示されれば成功！🎉

## エンドポイント

- `GET /api/health` - ヘルスチェック
- `POST /api/webhook` - LINE Webhook受信
- `POST /api/send` - メッセージ送信
- `GET /api/messages` - メッセージ取得
- `POST /api/messages` - メッセージ保存
- `GET /` または `GET /api` - サービス情報

## ローカル開発

```bash
# 環境変数設定
cp .env.example .env
# .env ファイルを編集

# 依存関係インストール
go mod tidy

# サーバー起動
go run main.go
```