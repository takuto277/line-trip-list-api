package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

// AppMessage represents a LINE message for the iOS app
type AppMessage struct {
	GroupID   string `json:"group_id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	UserName  string `json:"user_name"`
}

// 共有メモリストレージ（本番環境ではデータベースを使用すること）
var (
	receivedMessages []AppMessage
	mu               sync.Mutex
)

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Line-Signature")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	if channelSecret == "" {
		log.Printf("LINE_CHANNEL_SECRET not set")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cb, err := webhook.ParseRequest(channelSecret, r)
	if err != nil {
		log.Printf("Webhook parse error: %v", err)
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range cb.Events {
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			case webhook.TextMessageContent:
				handleTextMessage(e, message)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func handleTextMessage(event webhook.MessageEvent, message webhook.TextMessageContent) {
	// グループソースを取得
	groupSource, ok := event.Source.(webhook.GroupSource)
	if !ok {
		log.Printf("Not a group message, skipping")
		return
	}

	// ユーザー情報を取得
	userName := "Unknown User"
	userID := ""
	
	if groupSource.UserId != "" {
		userID = groupSource.UserId
		if len(userID) > 8 {
			userName = fmt.Sprintf("User-%s", userID[:8])
		} else {
			userName = fmt.Sprintf("User-%s", userID)
		}
	}

	appMessage := AppMessage{
		GroupID:   groupSource.GroupId,
		UserID:    userID,
		Message:   message.Text,
		Timestamp: event.Timestamp,
		UserName:  userName,
	}

	// TODO: ここでデータベースに保存
	notifyiOSApp(appMessage)

	log.Printf("📱 Group Message: %s from %s in group %s", 
		message.Text, userName, groupSource.GroupId)
}

func notifyiOSApp(message AppMessage) {
	// メッセージログ出力
	messageJSON, _ := json.MarshalIndent(message, "", "  ")
	log.Printf("📲 Received LINE Message:\n%s", messageJSON)
	
	// メモリ内に保存
	saveToMemory(message)
}

func saveToMemory(message AppMessage) {
	mu.Lock()
	defer mu.Unlock()
	
	receivedMessages = append(receivedMessages, message)
	log.Printf("✅ Message saved to memory. Total messages: %d", len(receivedMessages))
}

// GetReceivedMessages returns all received messages (used by messages.go)
func GetReceivedMessages() []AppMessage {
	mu.Lock()
	defer mu.Unlock()
	
	// コピーを返す
	messages := make([]AppMessage, len(receivedMessages))
	copy(messages, receivedMessages)
	return messages
}