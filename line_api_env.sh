#!/bin/zsh

# LINE API設echo "👤 User ID: ${LINE_USER_ID:0:10}..."
# 以下の値を実際のAPIキーに置き換えてください

# Channel Access Token（長い英数字文字列）
export LINE_CHANNEL_TOKEN="uAPviIKPrzgu3fIEdk5DOolX34+AwWVVF5sEmiDwDl+InL7tt+DcYKvlC7KV2DAvdjyDEuHb8hh6Q7JIxEDBL4A847UrXj9EIWHkLaDQoFyIeQoQHukhaG8+0A7gdrIZVOG2vsMoaTsTadYqxegh1QdB04t89/1O/w1cDnyilFU="  

# Channel Secret（Basic settingsから）
export LINE_CHANNEL_SECRET="6f118ccd25ffd18bdbe577c02371e053" 

# Channel ID（数字のみ、Basic settingsから）
export LINE_CHANNEL_ID="2008209395"

# 以下は後で設定
export LINE_USER_ID="Ue41eb551b7025cf64bd42e3c31445832"
export LINE_WEBHOOK_URL="YOUR_WEBHOOK_URL_HERE"  # 例: https://yourserver.com/webhook
export LINE_GROUP_ID="YOUR_GROUP_ID_HERE"       # 監視したいグループのID

echo "✅ LINE API環境変数が設定されました"
echo "🔑 Channel Token: ${LINE_CHANNEL_TOKEN:0:10}..."
echo "🔑 Channel Secret: ${LINE_CHANNEL_SECRET:0:10}..."
echo "🆔 Channel ID: ${LINE_CHANNEL_ID}"
echo "� User ID: ${LINE_USER_ID:0:10}..."
echo "🌐 Webhook URL: ${LINE_WEBHOOK_URL}"
echo "👥 Group ID: ${LINE_GROUP_ID}"
