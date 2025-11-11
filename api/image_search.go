package handler

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "time"
)

// Local helpers: Upstash Redis REST API helper (duplicate of messages.go helpers)
// These are kept local to this handler so the Vercel build (which may compile
// handlers independently) does not fail due to missing symbols.
func redisGet(key string) (interface{}, error) {
    urlStr := os.Getenv("KV_REST_API_URL")
    token := os.Getenv("KV_REST_API_TOKEN")
    if urlStr == "" || token == "" {
        return nil, fmt.Errorf("redis credentials not set")
    }

    reqBody, _ := json.Marshal([]interface{}{"GET", key})
    req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(reqBody))
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

    body, _ := io.ReadAll(resp.Body)
    var result struct{
        Result interface{} `json:"result"`
    }
    json.Unmarshal(body, &result)
    return result.Result, nil
}

func redisSet(key string, value string, ttlSeconds int) error {
    urlStr := os.Getenv("KV_REST_API_URL")
    token := os.Getenv("KV_REST_API_TOKEN")
    if urlStr == "" || token == "" {
        return fmt.Errorf("redis credentials not set")
    }

    cmd := []interface{}{"SET", key, value}
    if ttlSeconds > 0 {
        cmd = append(cmd, "EX", fmt.Sprintf("%d", ttlSeconds))
    }
    reqBody, _ := json.Marshal(cmd)
    req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(reqBody))
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("redis set failed: %s", string(body))
    }
    return nil
}

// /api/search_image?q=... -> { "imageUrl": "..." }
func Handler(w http.ResponseWriter, r *http.Request) {
    // Quick probe endpoint to verify handler is reachable and env vars exist.
    if r.URL.Query().Get("_probe") == "1" {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "_from": "image_search_handler",
            "GOOGLE_CSE_KEY_set": os.Getenv("GOOGLE_CSE_KEY") != "",
            "GOOGLE_CSE_CX_set": os.Getenv("GOOGLE_CSE_CX") != "",
            "KV_REST_API_URL_set": os.Getenv("KV_REST_API_URL") != "",
            "KV_REST_API_TOKEN_set": os.Getenv("KV_REST_API_TOKEN") != "",
        })
        return
    }

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
        // Read response body for debugging and return it plus the exact reqUrl
        bodyBytes, _ := io.ReadAll(resp.Body)
        var parsed interface{}
        if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
            parsed = string(bodyBytes)
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadGateway)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "reqUrl": reqUrl,
            "google": parsed,
        })
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
