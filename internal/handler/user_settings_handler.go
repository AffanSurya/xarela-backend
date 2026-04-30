package handler

import (
	"net/http"

	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserSettingsHandler struct {
	service service.UserSettingsService
}

func NewUserSettingsHandler(service service.UserSettingsService) UserSettingsHandler {
	return UserSettingsHandler{service: service}
}

func (h UserSettingsHandler) Register(group *echo.Group) {
	group.GET("/settings", h.Get)
	group.PUT("/settings", h.Update)
}

type updateUserSettingsRequest struct {
	BaseCurrency *string `json:"base_currency"`
	Timezone     *string `json:"timezone"`
	FullName     *string `json:"full_name"`
}

func (h UserSettingsHandler) Get(c echo.Context) error {
	user, ok := currentAuthUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized", "Missing or invalid access token", nil))
	}

	settings, err := h.service.Get(c.Request().Context(), uuid.MustParse(user.ID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "Internal server error", nil))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": settings})
}

func (h UserSettingsHandler) Update(c echo.Context) error {
	user, ok := currentAuthUser(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, errorResponse("unauthorized", "Missing or invalid access token", nil))
	}

	var request updateUserSettingsRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("validation_error", "Invalid input", nil))
	}

	settings, err := h.service.Update(c.Request().Context(), uuid.MustParse(user.ID), service.UpdateUserSettingsRequest{
		BaseCurrency: request.BaseCurrency,
		Timezone:     request.Timezone,
		FullName:     request.FullName,
	})
	if err != nil {
		return userSettingsErrorResponse(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": settings})
}

func currentAuthUser(c echo.Context) (service.AuthUser, bool) {
	value := c.Get("auth_user")
	if value == nil {
		return service.AuthUser{}, false
	}

	user, ok := value.(service.AuthUser)
	if !ok || user.ID == "" {
		return service.AuthUser{}, false
	}

	return user, true
}

func userSettingsErrorResponse(c echo.Context, err error) error {
	if validationErr, ok := err.(*service.ValidationError); ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": validationErr})
	}

	return c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "Internal server error", nil))
}
