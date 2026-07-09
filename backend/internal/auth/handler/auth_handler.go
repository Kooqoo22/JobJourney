package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/internal/auth/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/auth/mapper"
	"github.com/Kooqoo22/JobJourney/backend/internal/middleware"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type AuthHandler struct {
	usecase AuthUsecaseIface
}

func NewAuthHandler(u AuthUsecaseIface) *AuthHandler {
	return &AuthHandler{usecase: u}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	user, err := h.usecase.Register(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, utils.NewSuccess("registration successful, please check your email to verify your account", mapper.ToUserResponse(user)))
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	user, err := h.usecase.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccess("email verified successfully", mapper.ToUserResponse(user)))
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.usecase.ResendVerification(c.Request.Context(), req.Email); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("if the email is registered and unverified, a verification link has been sent"))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	result, err := h.usecase.Login(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccess("login successful", mapper.ToAuthResponse(result.User, result.AccessToken, result.RefreshToken, result.ExpiresIn)))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	result, err := h.usecase.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccess("token refreshed", mapper.ToTokenResponse(result.AccessToken, result.RefreshToken, result.ExpiresIn)))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.usecase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("logged out"))
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	if err := h.usecase.LogoutAll(c.Request.Context(), userID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("logged out from all devices"))
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.usecase.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("if the email is registered, a password reset link has been sent"))
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.usecase.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("password reset successful, please log in with your new password"))
}
