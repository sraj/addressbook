package application

type CreateRequest struct {
	Title   string `json:"title"   validate:"required,min=1,max=500"`
	Content string `json:"content" validate:"omitempty"`
}

type UpdateRequest struct {
	Title   string `json:"title"   validate:"required,min=1,max=500"`
	Content string `json:"content" validate:"omitempty"`
}
