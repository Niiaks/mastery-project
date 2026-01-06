package handler

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	env string
}

// NewHandler creates a new base handler
func NewHandler(env string) Handler {
	return Handler{env: env}
}

func (h Handler) JSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return
	}
}
