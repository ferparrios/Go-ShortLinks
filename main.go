package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

var linkStore = struct {
    sync.Mutex
    links map[string]string
}{links: make(map[string]string)}

const baseURL = "https://fer.link/"


type LinkRequest struct {
    URL string `json:"url"`
}


type LinkResponse struct {
    ShortURL string `json:"short_url"`
}


func generateShortCode() string {
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    var shortCode string
    for i := 0; i < 6; i++ {
        shortCode += string(letters[rand.Intn(len(letters))])
    }
    return shortCode
}


func enableCORS(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
        next(w, r)
    }
}


func shortenLink(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
        return
    }
		
    var req LinkRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    shortCode := generateShortCode()

    linkStore.Lock()
    linkStore.links[shortCode] = req.URL
    linkStore.Unlock()

    shortURL := baseURL + shortCode
    res := LinkResponse{ShortURL: shortURL}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}


func redirectLink(w http.ResponseWriter, r *http.Request) {
    shortCode := r.URL.Path[len("/"):]
    linkStore.Lock()
    originalURL, exists := linkStore.links[shortCode]
    linkStore.Unlock()

    if !exists {
        http.Error(w, "Short URL not found", http.StatusNotFound)
        return
    }
    http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
    
    http.HandleFunc("/shorten", enableCORS(shortenLink))
    http.HandleFunc("/", redirectLink)
    fmt.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
