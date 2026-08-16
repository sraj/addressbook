package shared

import (
	"encoding/json"
	"net/http"

	"github.com/mobentum/kern"
	"github.com/mobentum/kern/middleware"
)

type ErrorBody struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

type ErrorWithFields struct {
	Error     string      `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
	Fields    interface{} `json:"fields,omitempty"`
}

func ValidationError(c *kern.Context, err error) {
	_ = c.JSON(http.StatusUnprocessableEntity, ErrorWithFields{
		Error:     "validation failed",
		RequestID: middleware.RequestIDFromContext(c.Context()),
		Fields:    err,
	})
}

func WriteJSONError(w http.ResponseWriter, r *http.Request, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{
		Error:     message,
		RequestID: middleware.RequestIDFromContext(r.Context()),
	})
}

func SendError(c *kern.Context, status int, message string, errs ...error) {
	if len(errs) > 0 && errs[0] != nil {
		if l := kern.LoggerFromContext(c.Context()); l != nil {
			l.Error(message, "error", errs[0])
		}
	}
	_ = c.JSON(status, ErrorBody{
		Error:     message,
		RequestID: middleware.RequestIDFromContext(c.Context()),
	})
}
