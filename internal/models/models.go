package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Age             *int      `json:"age,omitempty"`
	IsActive        bool      `json:"is_active"`
	IsSuperuser     bool      `json:"is_superuser"`
	HashedPassword  string    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Task struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Priority    int       `json:"priority"`
	Completed   bool      `json:"completed"`
	UserID      uuid.UUID `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserCreate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Age      *int   `json:"age,omitempty"`
}

type UserUpdate struct {
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Age      *int    `json:"age,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Age         *int      `json:"age,omitempty"`
	IsActive    bool      `json:"is_active"`
	IsSuperuser bool     `json:"is_superuser"`
	CreatedAt  time.Time `json:"created_at"`
}

type TaskCreate struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Priority   *int    `json:"priority,omitempty"`
}

type TaskUpdate struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Priority    *int    `json:"priority,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

type TaskResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Priority    int       `json:"priority"`
	Completed   bool      `json:"completed"`
	UserID      uuid.UUID `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}