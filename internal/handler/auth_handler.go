package handler

import (
	"errors"
	"net/http"

	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/kp"
	"realtime-chat-system/pkg/logger"
	"realtime-chat-system/pkg/response"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Identifier string `json:"identifier"` // username or email
	Password   string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token string `json:"token"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

// Register handles user registration
func (h *AuthHandler) Register(ctx *kp.Ctx) {
	ctx.L("register")
	var customError *kp.Error
	// Only accept POST method
	// if r.Method != http.MethodPost {
	// 	response.BadRequest(w, "Method not allowed")
	// 	return
	// }

	// Parse request body
	var req RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid_request",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}

	// Register user
	user, _, err := h.authService.Register(ctx, req.Username, req.Email, req.Password, ctx.TraceID())
	if err != nil {
		// response.Error(w, err)
		if !errors.As(err, &customError) {
			customError = &kp.Error{
				StatusCode: http.StatusInternalServerError,
				Message:    "internal_server",
				Err:        err,
			}
		}
		ctx.JSONError(customError)
		return
	}

	// Return user data and token
	userResp := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// response.Created(w, userResp, "User registered successfully")
	ctx.JSON(http.StatusCreated, response.SuccessResponse{Data: userResp, Success: true})
}

// Login handles user login
func (h *AuthHandler) Login(ctx *kp.Ctx) {
	maskingBody := []logger.MaskRule{{
		Field: "body.identifier",
		Type:  logger.MaskFirst,
	}, {
		Field: "body.password",
		Type:  logger.MaskPassword,
	}}
	ssid := ctx.TraceID()
	ctx.L("login", maskingBody...)
	var customError *kp.Error

	// Only accept POST method
	// if ctx.Request.Method != http.MethodPost {
	// 	response.BadRequest(ctx.Writer, "Method not allowed")
	// 	return
	// }

	// Parse request body
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid_request",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}

	// Login user
	token, err := h.authService.Login(ctx, req.Identifier, req.Password, ssid)
	if err != nil {
		// response.Error(w, err)
		if !errors.As(err, &customError) {
			customError = &kp.Error{
				StatusCode: http.StatusInternalServerError,
				Message:    "internal_server",
				Err:        err,
			}
		}
		ctx.JSONError(customError)
		return
	}

	// Return token only
	authResp := AuthResponse{
		Token: token,
	}

	// response.Success(w, authResp, "Login successful")
	ctx.JSON(http.StatusOK, response.SuccessResponse{Data: authResp, Success: true})
}

// RegisterRoutes registers auth routes
func (h *AuthHandler) RegisterRoutes(app kp.IMicroservice) {
	app.GET("/api/auth/register", h.Register)
	app.POST("/api/auth/login", h.Login)
}
