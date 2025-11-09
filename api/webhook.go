package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

const redisKey = "line_messages"

// Upstash Redis REST API helper
func redisCommand(command []interface{}) (interface{}, error) {
	url := os.Getenv("KV_REST_API_URL")
	token := os.Getenv("KV_REST_API_TOKEN")
	
	if url == "" || token == "" {
		return nil, fmt.Errorf("redis credentials not set")
	}
	
	reqBody, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result struct {
		Result interface{} `json:"result"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return result.Result, nil
}
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
	"github.com/upstash/redis-go"
)
)

// AppMessage represents a LINE message for the iOS app
type AppMessage struct {
	GroupID   string `json:"group_id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	UserName  string `json:"user_name"`
}

const redisKey = "line_messages"

func getRedisClient() *redis.Client {
	url := os.Getenv("KV_REST_API_URL")
	token := os.Getenv("KV_REST_API_TOKEN")
	
	if url == "" || token == "" {
		log.Printf("Warning: Redis credentials not set")
		return nil
	}
	
	return redis.NewClient(&redis.Config{
		Addr:  url,
		Token: token,
	})
}

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
	
	// Redisに保存
	saveToRedis(message)
}

func saveToRedis(message AppMessage) {
	// 既存のメッセージを取得
	var messages []AppMessage
	result, err := redisCommand([]interface{}{"GET", redisKey})
	if err == nil && result != nil {
		if str, ok := result.(string); ok {
			json.Unmarshal([]byte(str), &messages)
		}
	}
	
	// 新しいメッセージを追加
	messages = append(messages, message)
	
	// Redisに保存
	jsonData, err := json.Marshal(messages)
	if err != nil {
		log.Printf("❌ Error marshaling messages: %v", err)
		return
	}
	
	_, err = redisCommand([]interface{}{"SET", redisKey, string(jsonData)})
	if err != nil {
		log.Printf("❌ Error saving to Redis: %v", err)
		return
	}
	
	log.Printf("✅ Message saved to Redis. Total messages: %d", len(messages))
}