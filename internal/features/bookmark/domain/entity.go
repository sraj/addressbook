package domain

type Bookmark struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	FaviconURL  string `json:"favicon_url,omitempty"`
	Category    string `json:"category,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
