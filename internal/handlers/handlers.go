package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"user-crud-api-go/internal/config"
	"user-crud-api-go/internal/models"
)

type Handlers struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Config *config.Config
}

func New(db *pgxpool.Pool, r *redis.Client, cfg *config.Config) *Handlers {
	return &Handlers{DB: db, Redis: r, Config: cfg}
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func (h *Handlers) Health(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func (h *Handlers) Register(c *fiber.Ctx) error {
	var input models.UserCreate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var exists bool
	err := h.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", input.Email).Scan(&exists)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	if exists {
		return c.Status(409).JSON(fiber.Map{"error": "User already exists"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Password hash failed"})
	}

	var user models.User
	err = h.DB.QueryRow(context.Background(),
		`INSERT INTO users(name,email,hashed_password,age) VALUES($1,$2,$3,$4) RETURNING *`,
		input.Name, input.Email, string(hash), input.Age,
	).Scan(&user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	h.Redis.Del(context.Background(), "users:list:*")

	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token generation failed"})
	}

	refresh, _ := h.generateRefreshToken(user.ID, user.Email)

	return c.Status(201).JSON(models.TokenResponse{
		AccessToken:  token,
		RefreshToken: refresh,
		TokenType:    "bearer",
	})
}

func (h *Handlers) Login(c *fiber.Ctx) error {
	var input models.LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user models.User
	err := h.DB.QueryRow(context.Background(),
		"SELECT * FROM users WHERE email=$1", input.Email).Scan(&user)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, _ := h.generateToken(user.ID, user.Email)
	refresh, _ := h.generateRefreshToken(user.ID, user.Email)

	return c.JSON(models.TokenResponse{
		AccessToken:  token,
		RefreshToken: refresh,
		TokenType:    "bearer",
	})
}

func (h *Handlers) Refresh(c *fiber.Ctx) error {
	var input models.RefreshRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	claims, err := jwt.ParseWithClaims(input.RefreshToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWTSecret), nil
	})
	if err != nil || claims == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}

	refreshClaims := claims.Claims.(*Claims)
	token, _ := h.generateToken(refreshClaims.UserID, refreshClaims.Email)
	newRefresh, _ := h.generateRefreshToken(refreshClaims.UserID, refreshClaims.Email)

	return c.JSON(models.TokenResponse{
		AccessToken:  token,
		RefreshToken: newRefresh,
		TokenType:    "bearer",
	})
}

func (h *Handlers) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := h.DB.Query(context.Background(),
		"SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2", pageSize, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	users := []models.UserResponse{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.IsActive, &u.IsSuperuser, &u.HashedPassword, &u.CreatedAt, &u.UpdatedAt); err == nil {
			users = append(users, models.UserResponse{
				ID: u.ID, Name: u.Name, Email: u.Email, Age: u.Age,
				IsActive: u.IsActive, IsSuperuser: u.IsSuperuser, CreatedAt: u.CreatedAt,
			})
		}
	}

	return c.JSON(users)
}

func (h *Handlers) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var user models.User
	err = h.DB.QueryRow(context.Background(),
		"SELECT * FROM users WHERE id=$1", id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Age, &user.IsActive, &user.IsSuperuser, &user.HashedPassword, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(models.UserResponse{
		ID: user.ID, Name: user.Name, Email: user.Email, Age: user.Age,
		IsActive: user.IsActive, IsSuperuser: user.IsSuperuser, CreatedAt: user.CreatedAt,
	})
}

func (h *Handlers) CreateUser(c *fiber.Ctx) error {
	var input models.UserCreate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	var user models.User
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO users(name,email,hashed_password,age) VALUES($1,$2,$3,$4) RETURNING *`,
		input.Name, input.Email, string(hash), input.Age,
	).Scan(&user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	h.Redis.Del(context.Background(), "users:list:*")

	return c.Status(201).JSON(models.UserResponse{
		ID: user.ID, Name: user.Name, Email: user.Email, Age: user.Age,
		IsActive: user.IsActive, IsSuperuser: user.IsSuperuser, CreatedAt: user.CreatedAt,
	})
}

func (h *Handlers) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var input models.UserUpdate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	query := "UPDATE users SET updated_at=NOW()"
	args := []interface{}{}
	argNum := 1

	if input.Name != nil {
		query += ", name=$" + string(rune('0'+argNum))
		args = append(args, *input.Name)
		argNum++
	}
	if input.Email != nil {
		query += ", email=$" + string(rune('0'+argNum))
		args = append(args, *input.Email)
		argNum++
	}
	if input.Age != nil {
		query += ", age=$" + string(rune('0'+argNum))
		args = append(args, *input.Age)
		argNum++
	}
	if input.IsActive != nil {
		query += ", is_active=$" + string(rune('0'+argNum))
		args = append(args, *input.IsActive)
		argNum++
	}

	query += " WHERE id=$" + string(rune('0'+argNum))
	args = append(args, id)

	var user models.User
	err = h.DB.QueryRow(context.Background(), query+" RETURNING *", args...).Scan(
		&user.ID, &user.Name, &user.Email, &user.Age, &user.IsActive, &user.IsSuperuser, &user.HashedPassword, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	h.Redis.Del(context.Background(), "users:"+id.String(), "users:list:*")

	return c.JSON(models.UserResponse{
		ID: user.ID, Name: user.Name, Email: user.Email, Age: user.Age,
		IsActive: user.IsActive, IsSuperuser: user.IsSuperuser, CreatedAt: user.CreatedAt,
	})
}

func (h *Handlers) DeleteUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	result, err := h.DB.Exec(context.Background(), "DELETE FROM users WHERE id=$1", id)
	if err != nil || result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	h.Redis.Del(context.Background(), "users:"+id.String(), "users:list:*")

	return c.SendStatus(204)
}

func (h *Handlers) ListTasks(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)
	offset := (page - 1) * pageSize

	rows, err := h.DB.Query(context.Background(),
		"SELECT * FROM tasks WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		userID, pageSize, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	tasks := []models.TaskResponse{}
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Priority, &t.Completed, &t.UserID, &t.CreatedAt, &t.UpdatedAt); err == nil {
			tasks = append(tasks, models.TaskResponse{
				ID: t.ID, Title: t.Title, Description: t.Description, Priority: t.Priority,
				Completed: t.Completed, UserID: t.UserID, CreatedAt: t.CreatedAt,
			})
		}
	}

	return c.JSON(tasks)
}

func (h *Handlers) GetTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	var task models.Task
	err = h.DB.QueryRow(context.Background(),
		"SELECT * FROM tasks WHERE id=$1", id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Priority, &task.Completed, &task.UserID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Task not found"})
	}

	return c.JSON(models.TaskResponse{
		ID: task.ID, Title: task.Title, Description: task.Description, Priority: task.Priority,
		Completed: task.Completed, UserID: task.UserID, CreatedAt: task.CreatedAt,
	})
}

func (h *Handlers) CreateTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var input models.TaskCreate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	priority := 1
	if input.Priority != nil {
		priority = *input.Priority
	}

	var task models.Task
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO tasks(title,description,priority,user_id) VALUES($1,$2,$3,$4) RETURNING *`,
		input.Title, input.Description, priority, userID,
	).Scan(&task)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create task"})
	}

	h.Redis.Del(context.Background(), "tasks:user:"+userID.String())

	return c.Status(201).JSON(models.TaskResponse{
		ID: task.ID, Title: task.Title, Description: task.Description, Priority: task.Priority,
		Completed: task.Completed, UserID: task.UserID, CreatedAt: task.CreatedAt,
	})
}

func (h *Handlers) UpdateTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	var input models.TaskUpdate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	query := "UPDATE tasks SET updated_at=NOW()"
	args := []interface{}{}
	argNum := 1

	if input.Title != nil {
		query += ", title=$" + string(rune('0'+argNum))
		args = append(args, *input.Title)
		argNum++
	}
	if input.Description != nil {
		query += ", description=$" + string(rune('0'+argNum))
		args = append(args, *input.Description)
		argNum++
	}
	if input.Priority != nil {
		query += ", priority=$" + string(rune('0'+argNum))
		args = append(args, *input.Priority)
		argNum++
	}
	if input.Completed != nil {
		query += ", completed=$" + string(rune('0'+argNum))
		args = append(args, *input.Completed)
		argNum++
	}

	query += " WHERE id=$" + string(rune('0'+argNum))
	args = append(args, taskID)
	argNum++
	query += " AND user_id=$" + string(rune('0'+argNum))
	args = append(args, userID)

	var task models.Task
	err = h.DB.QueryRow(context.Background(), query+" RETURNING *", args...).Scan(
		&task.ID, &task.Title, &task.Description, &task.Priority, &task.Completed, &task.UserID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Task not found"})
	}

	h.Redis.Del(context.Background(), "tasks:"+taskID.String(), "tasks:user:"+userID.String())

	return c.JSON(models.TaskResponse{
		ID: task.ID, Title: task.Title, Description: task.Description, Priority: task.Priority,
		Completed: task.Completed, UserID: task.UserID, CreatedAt: task.CreatedAt,
	})
}

func (h *Handlers) DeleteTask(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	result, err := h.DB.Exec(context.Background(),
		"DELETE FROM tasks WHERE id=$1 AND user_id=$2", taskID, userID)
	if err != nil || result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Task not found"})
	}

	h.Redis.Del(context.Background(), "tasks:"+taskID.String(), "tasks:user:"+userID.String())

	return c.SendStatus(204)
}

func (h *Handlers) generateToken(userID uuid.UUID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.Config.JWTExpiration * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.Config.JWTSecret))
}

func (h *Handlers) generateRefreshToken(userID uuid.UUID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.Config.RefreshExp * 24*time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.Config.JWTSecret))
}

func (h *Handlers) AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	tokenStr := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr = authHeader[7:]
	} else {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid authorization header"})
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("email", claims.Email)

	return c.Next()
}