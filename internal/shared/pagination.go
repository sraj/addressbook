package shared

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Size       int         `json:"size"`
	TotalPages int         `json:"total_pages"`
}

func NewPaginatedResponse(data interface{}, total int64, page, size int) PaginatedResponse {
	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}
	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}
}
