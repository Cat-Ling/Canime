package gelbooru

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"booru-aggregator/internal/imageboard"
)

const (
	gelbooruAPIURL = "https://api.rule34.xxx/index.php?page=dapi&s=post&q=index"
)

// GelbooruPost represents the structure of a post from the Gelbooru API.
type GelbooruPost struct {
	ID        int    `json:"id"`
	FileURL   string `json:"file_url"`
	PreviewURL string `json:"preview_url"`
	Tags      string `json:"tags"`
}

// Client is the Gelbooru API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	UserID     string
}

// NewClient creates a new Gelbooru API client.
func NewClient() *Client {
	return &Client{
		BaseURL:    gelbooruAPIURL,
		HTTPClient: &http.Client{},
		APIKey:     os.Getenv("GELBOORU_API_KEY"), // Or replace with your key
		UserID:     os.Getenv("GELBOORU_USER_ID"), // Or replace with your user ID
	}
}

// SearchPosts searches for posts on Gelbooru.
func (c *Client) SearchPosts(tags string, page, limit int) ([]imageboard.Post, error) {
	endpoint := fmt.Sprintf("%s&tags=%s&pid=%d&limit=%d&json=1", c.BaseURL, tags, page, limit)
	if c.APIKey != "" && c.UserID != "" {
		endpoint = fmt.Sprintf("%s&api_key=%s&user_id=%s", endpoint, c.APIKey, c.UserID)
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gelbooru API request failed with status: %s", resp.Status)
	}

	var gelbooruPosts struct {
		Post []GelbooruPost `json:"post"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gelbooruPosts); err != nil {
		// It's possible the response is an array if there's only one result
		var singlePost []GelbooruPost
		if err := json.NewDecoder(resp.Body).Decode(&singlePost); err != nil {
			return nil, err
		}
		gelbooruPosts.Post = singlePost
	}

	var posts []imageboard.Post
	for _, p := range gelbooruPosts.Post {
		posts = append(posts, imageboard.Post{
			ID:        p.ID,
			FileURL:   p.FileURL,
			PreviewURL: p.PreviewURL,
			Tags:      strings.Split(p.Tags, " "),
			Source:    "Rule34",
		})
	}

	return posts, nil
}