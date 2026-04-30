package handler

import (
	"net/http"

	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) AuthHandler {
	return AuthHandler{service: service}
}

func (h AuthHandler) Register(group *echo.Group) {
	group.POST("/register", h.RegisterUser)
	group.POST("/verify-email", h.VerifyEmail)
	group.POST("/login", h.Login)
	group.POST("/refresh", h.Refresh)
	group.POST("/logout", h.Logout)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type verifyEmailRequest struct {
	VerificationToken string `json:"verification_token"`
}

func (h AuthHandler) RegisterUser(c echo.Context) error {
	var request registerRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("validation_error", "Invalid input", nil))
	}

	result, err := h.service.Register(c.Request().Context(), service.RegisterRequest{
		Email:    request.Email,
		Password: request.Password,
		FullName: request.FullName,
	})
	if err != nil {
		return authErrorResponse(c, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"data": map[string]any{
			"user":          result.User,
			"access_token":  result.Token.AccessToken,
			"refresh_token": result.Token.RefreshToken,
			"expires_in":    result.Token.ExpiresIn,
			"token_type":    result.Token.TokenType,
		},
	})
}

func (h AuthHandler) VerifyEmail(c echo.Context) error {
	var request verifyEmailRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("validation_error", "Invalid input", nil))
	}

	result, err := h.service.VerifyEmail(c.Request().Context(), service.VerifyEmailRequest{VerificationToken: request.VerificationToken})
	if err != nil {
		return authErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (h AuthHandler) Login(c echo.Context) error {
	var request loginRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("validation_error", "Invalid input", nil))
	}

	result, err := h.service.Login(c.Request().Context(), service.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return authErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"access_token":  result.Token.AccessToken,
			"refresh_token": result.Token.RefreshToken,
			"expires_in":    result.Token.ExpiresIn,
			"token_type":    result.Token.TokenType,
		},
	})
}

func (h AuthHandler) Refresh(c echo.Context) error {
	var request refreshRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("validation_error", "Invalid input", nil))
	}

	result, err := h.service.Refresh(c.Request().Context(), service.RefreshRequest{RefreshToken: request.RefreshToken})
	if err != nil {
		return authErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"access_token":  result.Token.AccessToken,
			"refresh_token": result.Token.RefreshToken,
			"expires_in":    result.Token.ExpiresIn,
			"token_type":    result.Token.TokenType,
		},
	})
}

func (h AuthHandler) Logout(c echo.Context) error {
	var request refreshRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("validation_error", "Invalid input", nil))
	}

	result, err := h.service.Logout(c.Request().Context(), service.RefreshRequest{RefreshToken: request.RefreshToken})
	if err != nil {
		return authErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func authErrorResponse(c echo.Context, err error) error {
	if validationErr, ok := err.(*service.ValidationError); ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": validationErr})
	}

	switch err {
	case service.ErrEmailAlreadyExists:
		return c.JSON(http.StatusConflict, errorResponse("email_already_exists", "Email already registered", map[string]string{"email": "duplicate"}))
	case service.ErrInvalidCredentials:
		return c.JSON(http.StatusUnauthorized, errorResponse("invalid_credentials", "Email or password is incorrect", nil))
	case service.ErrRefreshTokenInvalid:
		return c.JSON(http.StatusUnauthorized, errorResponse("refresh_token_invalid", "Refresh token is invalid or expired", nil))
	case service.ErrRefreshTokenReuse:
		return c.JSON(http.StatusUnauthorized, errorResponse("refresh_token_reuse", "Refresh token reuse detected", nil))
	case service.ErrEmailVerificationInvalid:
		return c.JSON(http.StatusBadRequest, errorResponse("email_verification_invalid", "Verification token is invalid or expired", nil))
	case service.ErrEmailVerificationExpired:
		return c.JSON(http.StatusBadRequest, errorResponse("email_verification_expired", "Verification token has expired", nil))
	case service.ErrUnauthorized:
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized", "Refresh token is missing or invalid", nil))
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "Internal server error", nil))
	}
}

func errorResponse(code, message string, fields map[string]string) map[string]any {
	payload := map[string]any{
		"code":    code,
		"message": message,
	}
	if fields != nil {
		payload["fields"] = fields
	}
	return map[string]any{"error": payload}
}
