package application

type CreateRequest struct {
	URL         string `json:"url"         validate:"required,url"`
	Title       string `json:"title"       validate:"required,min=1,max=500"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	FaviconURL  string `json:"favicon_url"  validate:"omitempty"`
	Category    string `json:"category"    validate:"omitempty,max=100"`
}

type UpdateRequest struct {
	URL         string `json:"url"         validate:"required,url"`
	Title       string `json:"title"       validate:"required,min=1,max=500"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	FaviconURL  string `json:"favicon_url"  validate:"omitempty"`
	Category    string `json:"category"    validate:"omitempty,max=100"`
}

type ImportItem struct {
	URL         string `json:"url"         validate:"required,url"`
	Title       string `json:"title"       validate:"required,min=1,max=500"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	FaviconURL  string `json:"favicon_url"  validate:"omitempty"`
	Category    string `json:"category"    validate:"omitempty,max=100"`
}

type ImportRequest struct {
	Bookmarks []ImportItem `json:"bookmarks" validate:"required,min=1,dive"`
}

type ImportResponse struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}

type ImportHTMLRequest struct {
	HTML string `json:"html" validate:"required"`
}
