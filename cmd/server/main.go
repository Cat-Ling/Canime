package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"booru-aggregator/internal/danbooru"
	"booru-aggregator/internal/gelbooru"
	"booru-aggregator/internal/imageboard"
)

type searchResult struct {
	Posts []imageboard.Post `json:"posts"`
	Total int               `json:"total"`
}

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API endpoint for searching posts
	http.HandleFunc("/api/search", searchHandler)
	// API endpoint for proxying images
	http.HandleFunc("/api/proxy", proxyHandler)

	// Route for the homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/index.html")
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	tags := r.URL.Query().Get("tags")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 40 // Default limit
	}

	danbooruClient := danbooru.NewClient()
	gelbooruClient := gelbooru.NewClient()

	var wg sync.WaitGroup
	var danbooruPosts, gelbooruPosts []imageboard.Post
	var danbooruErr, gelbooruErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		danbooruPosts, danbooruErr = danbooruClient.SearchPosts(tags, page, limit/2)
	}()

	go func() {
		defer wg.Done()
		gelbooruPosts, gelbooruErr = gelbooruClient.SearchPosts(tags, page-1, limit/2) // Gelbooru is 0-indexed
	}()

	wg.Wait()

	if danbooruErr != nil {
		http.Error(w, "Failed to fetch from Danbooru", http.StatusInternalServerError)
		log.Printf("Danbooru API error: %v", danbooruErr)
		return
	}
	if gelbooruErr != nil {
		http.Error(w, "Failed to fetch from Gelbooru", http.StatusInternalServerError)
		log.Printf("Gelbooru API error: %v", gelbooruErr)
		return
	}

	allPosts := append(danbooruPosts, gelbooruPosts...)
	result := searchResult{
		Posts: allPosts,
		Total: len(allPosts),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the headers from the original response to our response
	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	// Stream the image data
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Failed to stream image: %v", err)
	}
}