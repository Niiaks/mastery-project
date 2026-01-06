package handler

import (
	"mastery-project/internal/config"
	"mastery-project/internal/service"
)

type Handlers struct {
	Health *HealthHandler
	Auth   *AuthHandler
	Item   *ItemHandler
}

func NewHandlers(cfg *config.Config, service *service.Services) *Handlers {
	return &Handlers{
		Health: NewHealthHandler(cfg),
		Auth:   NewAuthHandler(cfg, service.Auth),
		Item:   NewItemHandler(cfg, service.Item),
	}
}
