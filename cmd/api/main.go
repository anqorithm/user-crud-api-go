package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"user-crud-api-go/internal/config"
	"user-crud-api-go/internal/handlers"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	defer db.Close()

	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			age INTEGER,
			is_active BOOLEAN NOT NULL DEFAULT true,
			is_superuser BOOLEAN NOT NULL DEFAULT false,
			hashed_password VARCHAR(255) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create users table")
	}

	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			priority INTEGER NOT NULL DEFAULT 1,
			completed BOOLEAN NOT NULL DEFAULT false,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create tasks table")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	h := handlers.New(db, rdb, cfg)

	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/health", h.Health)

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)

	users := api.Group("/users")
	users.Get("/", h.ListUsers)
	users.Get("/:id", h.GetUser)
	users.Post("/", h.CreateUser)
	users.Patch("/:id", h.UpdateUser)
	users.Delete("/:id", h.DeleteUser)

	tasks := api.Group("/tasks", h.AuthMiddleware)
	tasks.Get("/", h.ListTasks)
	tasks.Get("/:id", h.GetTask)
	tasks.Post("/", h.CreateTask)
	tasks.Patch("/:id", h.UpdateTask)
	tasks.Delete("/:id", h.DeleteTask)

	log.Println("Server starting on :8000")
	log.Fatal(app.Listen(":8000"))
}