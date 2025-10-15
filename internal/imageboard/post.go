package imageboard

// Post represents a unified structure for posts from different imageboards.
type Post struct {
	ID        int      `json:"id"`
	FileURL   string   `json:"file_url"`
	PreviewURL string  `json:"preview_url"`
	Tags      []string `json:"tags"`
	Source    string   `json:"source"`
}