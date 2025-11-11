package handler

import (
    "encoding/json"
    "net/http"
    "os"
)

// Simple probe endpoint to verify function routing
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "_from": "sanity_search",
        "GOOGLE_CSE_KEY_set": os.Getenv("GOOGLE_CSE_KEY") != "",
        "GOOGLE_CSE_CX_set": os.Getenv("GOOGLE_CSE_CX") != "",
    })
}
