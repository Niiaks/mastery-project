package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"mastery-project/internal/model"
	"mastery-project/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewAuthService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (auth *AuthService) Login(ctx context.Context, request model.LoginRequest) (*model.UserResponse, string, error) {
	user, err := auth.userRepo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	value := GenerateSessionID()
	//create session
	session := &model.Session{
		UserID:    user.ID,
		SessionID: value,
	}
	sessErr := auth.sessionRepo.CreateSession(ctx, session)
	if sessErr != nil {
		return nil, "", sessErr
	}

	response := &model.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}

	return response, value, nil

}

func (auth *AuthService) Register(ctx context.Context, request model.CreateUserRequest) (*model.UserResponse, error) {
	exists, err := auth.userRepo.EmailExists(ctx, request.Email)

	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Email:    request.Email,
		Password: string(hash),
		Name:     request.Name,
	}

	createErr := auth.userRepo.CreateUser(ctx, user)
	if createErr != nil {
		return nil, createErr
	}
	response := &model.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}
	return response, nil
}

func GenerateSessionID() string {
	key := rand.Text()
	return base64.URLEncoding.EncodeToString([]byte(key))
}
