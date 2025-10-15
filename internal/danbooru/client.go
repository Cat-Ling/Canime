package danbooru

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"booru-aggregator/internal/imageboard"
)

const (
	danbooruAPIURL = "https://danbooru.donmai.us"
)

// DanbooruPost represents the structure of a post from the Danbooru API.
type DanbooruPost struct {
	ID         int    `json:"id"`
	FileURL    string `json:"file_url"`
	PreviewURL string `json:"preview_file_url"`
	TagString  string `json:"tag_string"`
}

// Client is the Danbooru API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	Username   string
}

// NewClient creates a new Danbooru API client.
func NewClient() *Client {
	return &Client{
		BaseURL:    danbooruAPIURL,
		HTTPClient: &http.Client{},
		APIKey:     os.Getenv("DANBOORU_API_KEY"),
		Username:   os.Getenv("DANBOORU_USERNAME"),
	}
}

// SearchPosts searches for posts on Danbooru.
func (c *Client) SearchPosts(tags string, page, limit int) ([]imageboard.Post, error) {
	endpoint := fmt.Sprintf("%s/posts.json", c.BaseURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("tags", tags)
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("limit", fmt.Sprintf("%d", limit))
	if c.APIKey != "" && c.Username != "" {
		q.Set("login", c.Username)
		q.Set("api_key", c.APIKey)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Danbooru API request failed with status: %s", resp.Status)
	}

	var danbooruPosts []DanbooruPost
	if err := json.NewDecoder(resp.Body).Decode(&danbooruPosts); err != nil {
		return nil, err
	}

	var posts []imageboard.Post
	for _, p := range danbooruPosts {
		posts = append(posts, imageboard.Post{
			ID:         p.ID,
			FileURL:    p.FileURL,
			PreviewURL: p.PreviewURL,
			Tags:       strings.Split(p.TagString, " "),
			Source:     "Danbooru",
		})
	}

	return posts, nil
}