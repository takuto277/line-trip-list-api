package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"
)

// Message represents a stored LINE message
type Message struct {
	ID        int    `json:"id"`
	GroupID   string `json:"group_id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	UserName  string `json:"user_name"`
	Timestamp int64  `json:"timestamp"`
	CreatedAt string `json:"created_at"`
}

// 一時的にメモリ内ストレージ（後でデータベースに置き換え）
var messages []Message
var messageID int = 1

func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS設定
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		// HTMLとJSONの両方に対応
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			serveHTML(w, r)
		} else {
			serveJSON(w, r)
		}

	case "POST":
		// 新しいメッセージを保存
		var msg Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// IDと作成日時を設定
		msg.ID = messageID
		messageID++
		msg.CreatedAt = time.Now().Format(time.RFC3339)

		// メッセージを保存
		messages = append(messages, msg)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success", 
			"message": "Message saved",
			"data":    msg,
		})

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func serveJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
		"note":     "Note: Messages are stored in memory and will be lost on restart. Use a database for production.",
	}
	json.NewEncoder(w).Encode(response)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	htmlContent := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LINE Bot - 受信メッセージ一覧</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            min-height: 100vh;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
        }
        .header {
            background: white;
            padding: 30px;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
            margin-bottom: 20px;
            text-align: center;
        }
        .header h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 2em;
        }
        .header .emoji {
            font-size: 3em;
            margin-bottom: 10px;
        }
        .stats {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin-top: 20px;
        }
        .stat-box {
            background: #f8f9fa;
            padding: 15px 30px;
            border-radius: 10px;
            text-align: center;
        }
        .stat-box .number {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        .stat-box .label {
            color: #666;
            font-size: 0.9em;
            margin-top: 5px;
        }
        .message-card {
            background: white;
            padding: 20px;
            border-radius: 15px;
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.1);
            margin-bottom: 15px;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .message-card:hover {
            transform: translateY(-3px);
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
        }
        .message-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #f0f0f0;
        }
        .user-info {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .user-avatar {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
        }
        .user-name {
            font-weight: 600;
            color: #333;
        }
        .timestamp {
            color: #999;
            font-size: 0.85em;
        }
        .message-text {
            color: #555;
            line-height: 1.6;
            font-size: 1.05em;
            padding: 10px;
            background: #f8f9fa;
            border-radius: 8px;
            margin-bottom: 10px;
        }
        .message-meta {
            display: flex;
            gap: 15px;
            font-size: 0.85em;
            color: #666;
            margin-top: 10px;
        }
        .meta-item {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .empty-state {
            background: white;
            padding: 60px 30px;
            border-radius: 15px;
            text-align: center;
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.1);
        }
        .empty-state .emoji {
            font-size: 4em;
            margin-bottom: 20px;
        }
        .empty-state h2 {
            color: #333;
            margin-bottom: 10px;
        }
        .empty-state p {
            color: #666;
        }
        .refresh-btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            font-size: 1em;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            margin-top: 20px;
        }
        .refresh-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }
        .note {
            background: rgba(255, 255, 255, 0.9);
            padding: 15px;
            border-radius: 10px;
            margin-top: 20px;
            text-align: center;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="emoji">💬</div>
            <h1>LINE Bot 受信メッセージ</h1>
            <p style="color: #666; margin-top: 10px;">グループで受信したメッセージ一覧</p>
            <div class="stats">
                <div class="stat-box">
                    <div class="number">` + fmt.Sprintf("%d", len(messages)) + `</div>
                    <div class="label">受信メッセージ数</div>
                </div>
            </div>
            <button class="refresh-btn" onclick="location.reload()">🔄 更新</button>
        </div>
`

	if len(messages) == 0 {
		htmlContent += `
        <div class="empty-state">
            <div class="emoji">📭</div>
            <h2>メッセージがありません</h2>
            <p>LINEグループでメッセージを送信してください</p>
            <p style="margin-top: 10px; color: #999;">Webhookが正しく設定されているか確認してください</p>
        </div>`
	} else {
		// メッセージを新しい順に表示
		for i := len(messages) - 1; i >= 0; i-- {
			msg := messages[i]
			timestamp := time.Unix(msg.Timestamp/1000, 0).Format("2006/01/02 15:04:05")
			htmlContent += fmt.Sprintf(`
        <div class="message-card">
            <div class="message-header">
                <div class="user-info">
                    <div class="user-avatar">%s</div>
                    <div class="user-name">%s</div>
                </div>
                <div class="timestamp">%s</div>
            </div>
            <div class="message-text">%s</div>
            <div class="message-meta">
                <div class="meta-item">🆔 ID: %d</div>
                <div class="meta-item">👥 Group: %s</div>
            </div>
        </div>`,
				getInitial(msg.UserName),
				html.EscapeString(msg.UserName),
				timestamp,
				html.EscapeString(msg.Message),
				msg.ID,
				truncate(msg.GroupID, 20))
		}
	}

	htmlContent += `
        <div class="note">
            ⚠️ メッセージはメモリ内に保存されており、サーバー再起動時に消えます。<br>
            本番環境ではデータベースの使用を推奨します。
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(htmlContent))
}

func getInitial(userName string) string {
	if len(userName) > 0 {
		return string([]rune(userName)[0])
	}
	return "?"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}