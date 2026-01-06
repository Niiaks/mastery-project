package handler

import (
	"mastery-project/internal/config"
	"net/http"
	"time"
)

type HealthHandler struct {
	Handler
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{Handler: NewHandler(cfg.ENV)}
}

func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC(),
		"env":    h.env,
	}
	//check db here
	h.JSON(w, http.StatusOK, response)
}
