package http

import (
	"net/http"
	"strconv"
	"strings"

	"backend-challenge/internal/adapters/http/dto"
	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users  ports.UserService
	tokens ports.TokenProvider
}

func NewUserHandler(users ports.UserService, tokens ports.TokenProvider) *UserHandler {
	return &UserHandler{users: users, tokens: tokens}
}

// Register handles user registration
//
//	@Summary		Register a new user
//	@Description	Create a new user account with name, email, and password.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		dto.RegisterRequest	true	"User Registration Details"
//	@Success		201		{object}	domain.User
//	@Failure		400		{object}	map[string]string	"Invalid input or email already exists"
//	@Router			/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var payload dto.RegisterRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		writeValidationError(c, err)
		return
	}
	created, err := h.users.Create(c.Request.Context(), domain.CreateUser{
		Name:     strings.TrimSpace(payload.Name),
		Email:    strings.TrimSpace(payload.Email),
		Password: payload.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, created)
}

// Login handles user authentication
//
//	@Summary		Login and get a JWT token
//	@Description	Authenticate with email and password to receive a JWT token for subsequent requests.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		dto.LoginRequest	true	"Login Credentials"
//	@Success		200		{object}	map[string]string	"token: JWT_TOKEN"
//	@Failure		401		{object}	map[string]string	"Invalid credentials"
//	@Router			/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var payload dto.LoginRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		writeValidationError(c, err)
		return
	}
	u, err := h.users.Authenticate(c.Request.Context(), payload.Email, payload.Password)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := h.tokens.Create(u.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not create token")
		return
	}
	writeJSON(c, http.StatusOK, map[string]string{"token": token})
}

// CreateUser creates a new user (Admin/Authorized)
//
//	@Summary		Create a new user
//	@Description	Endpoint for authorized users to create a new user.
//	@Tags			users
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		dto.CreateUserRequest	true	"New User Details"
//	@Success		201		{object}	domain.User
//	@Failure		400		{object}	map[string]string	"Invalid input"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Router			/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var payload dto.CreateUserRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		writeValidationError(c, err)
		return
	}
	created, err := h.users.Create(c.Request.Context(), domain.CreateUser{
		Name:     strings.TrimSpace(payload.Name),
		Email:    strings.TrimSpace(payload.Email),
		Password: payload.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, created)
}

// GetUser retrieves a user by ID
//
//	@Summary		Get user by ID
//	@Description	Fetch details of a specific user.
//	@Tags			users
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	domain.User
//	@Failure		404	{object}	map[string]string	"User not found"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Router			/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	usr, err := h.users.Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, usr)
}

// ListUsers lists all users
//
//	@Summary		List all users
//	@Description	Retrieve a list of all registered users.
//	@Tags			users
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Success		200	{object}	dto.UserListResponse
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Router			/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid page")
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid limit")
		return
	}
	users, total, err := h.users.List(c.Request.Context(), page, limit)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to list users")
		return
	}
	writeJSON(c, http.StatusOK, dto.BaseResponse[dto.UserListResponse]{
		Data: dto.UserListResponse{
			Users: users,
			Total: total,
		},
	})
}

// UpdateUser updates an existing user
//
//	@Summary		Update user details
//	@Description	Modify name or email of a user.
//	@Tags			users
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			payload	body		dto.UpdateUserRequest	true	"Update Details"
//	@Success		200		{object}	domain.User
//	@Failure		400		{object}	map[string]string	"Invalid input"
//	@Failure		404		{object}	map[string]string	"User not found"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Router			/users/{id} [patch]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var payload dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		writeValidationError(c, err)
		return
	}
	updated, err := h.users.Update(c.Request.Context(), id, domain.UpdateUser{
		Name:  payload.Name,
		Email: payload.Email,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, updated)
}

// DeleteUser deletes a user by ID
//
//	@Summary		Delete a user
//	@Description	Remove a user account permanently.
//	@Tags			users
//	@Security		ApiKeyAuth
//	@Produce		json
//	@Param			id	path		string				true	"User ID"
//	@Success		200	{object}	map[string]string	"status: deleted"
//	@Failure		404	{object}	map[string]string	"User not found"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Router			/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.users.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]string{"status": "deleted"})
}
