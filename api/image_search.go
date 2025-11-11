package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "time"
)

// /api/search_image?q=... -> { "imageUrl": "..." }
func Handler(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")
    if q == "" {
        http.Error(w, "missing q", http.StatusBadRequest)
        return
    }
    // Try Redis cache first (keyed by escaped query)
    cacheKey := "image_search:" + url.QueryEscape(q)
    if res, err := redisGet(cacheKey); err == nil && res != nil {
        var s string
        switch v := res.(type) {
        case string:
            s = v
        case []byte:
            s = string(v)
        }
        if s != "" {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]string{"imageUrl": s})
            return
        }
    }

    apiKey := os.Getenv("GOOGLE_CSE_KEY")
    cx := os.Getenv("GOOGLE_CSE_CX")
    if apiKey == "" || cx == "" {
        http.Error(w, "image search not configured", http.StatusNotImplemented)
        return
    }

    // Build Google Custom Search URL (image search)
    base := "https://www.googleapis.com/customsearch/v1"
    vals := url.Values{}
    vals.Set("key", apiKey)
    vals.Set("cx", cx)
    vals.Set("q", q)
    vals.Set("searchType", "image")
    vals.Set("num", "1")

    reqUrl := fmt.Sprintf("%s?%s", base, vals.Encode())

    client := &http.Client{Timeout: 8 * time.Second}
    resp, err := client.Get(reqUrl)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        http.Error(w, "image search failed", http.StatusBadGateway)
        return
    }

    // Parse JSON response
    var body struct {
        Items []struct {
            Link string `json:"link"`
        } `json:"items"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        http.Error(w, "failed to parse search response", http.StatusInternalServerError)
        return
    }

    if len(body.Items) == 0 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"imageUrl": ""})
        return
    }

    imageUrl := body.Items[0].Link

    // Cache in Redis with TTL (24h)
    // Use redisSet helper from messages.go
    _ = redisSet(cacheKey, imageUrl, 86400)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"imageUrl": imageUrl})
}
