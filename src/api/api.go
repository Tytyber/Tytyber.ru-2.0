package api

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"src/api/user_operation"
	"time"
)

var jwtSecret = []byte("super_secret_key")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  int    `json:"role"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// LoginHandler — основной HTTP-обработчик авторизации
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role, err := user_operation.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Создаём JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 часа
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Token: tokenString,
		Role:  role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterHandler — HTTP-обработчик регистрации нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := user_operation.RegisterUser(req.Username, req.Password, req.Email)
	if err != nil {
		// Обрабатываем ошибки понятным пользователю языком
		switch err {
		case user_operation.ErrUserExists:
			http.Error(w, "Пользователь с таким именем уже существует", http.StatusConflict)
		case user_operation.ErrInvalidInput:
			http.Error(w, "Некорректные данные для регистрации", http.StatusBadRequest)
		default:
			http.Error(w, "Ошибка сервера: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := RegisterResponse{
		Message: "Регистрация успешна! Добро пожаловать, " + req.Username + "!",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
