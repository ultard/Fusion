package handlers

import (
	"errors"
	"fusion/app/database/models"
	"fusion/app/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type AuthRoute struct {
	config   utils.AppConfig
	jwt      utils.JWTService
	email    utils.EmailService
	db       *gorm.DB
	validate *validator.Validate
}

// RegisterAuthRoutes регистрирует маршруты для аутентификации
func RegisterAuthRoutes(app *fiber.App, db *gorm.DB, config utils.AppConfig, jwtService utils.JWTService, email utils.EmailService) {
	handler := &AuthRoute{
		config:   config,
		jwt:      jwtService,
		email:    email,
		db:       db,
		validate: validator.New(),
	}

	authGroup := app.Group("/auth")
	authGroup.Post("/register", handler.Register)
	authGroup.Post("/login", handler.Login)
	authGroup.Post("/logout", handler.Logout)
	authGroup.Post("/refresh", handler.Refresh)
	authGroup.Post("/reset-password", handler.ResetPassword)
	authGroup.Post("/change-password", handler.VerifyPasswordReset)
	authGroup.Post("/verify-email", handler.VerifyEmail)
}

// Register обрабатывает регистрацию нового пользователя
func (h AuthRoute) Register(c *fiber.Ctx) error {
	type RegisterInput struct {
		Username    string `json:"username" validate:"required,min=3,max=32"`
		Email       string `json:"email"    validate:"required,email"`
		Password    string `json:"password" validate:"required,min=8"`
		RedirectUrl string `json:"redirect_url" validate:"required"`
	}

	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.validate.Struct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input data")
	}

	var user models.User
	if err := h.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, "could not create user")
		}
	}

	if user.ID != uuid.Nil && user.IsEmailVerified {
		return fiber.NewError(fiber.StatusBadRequest, "user already registered")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not hash password")
	}

	newUser := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not start transaction")
	}

	if user.ID == uuid.Nil {
		if err := tx.Create(&newUser).Error; err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, "could not create user")
		}
		user = newUser
	} else {
		if err := tx.Model(&user).Updates(newUser).Error; err != nil {
			tx.Rollback()
			return fiber.NewError(fiber.StatusInternalServerError, "could not update user")
		}
	}

	token := uuid.New().String()
	verification := models.Verification{
		Type:      "EMAIL_VERIFY",
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(h.config.VerificationExpire),
	}

	if err := tx.Create(&verification).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "could not create verification token")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not commit transaction")
	}

	// TODO: перейти на rabbitmq
	type EmailData struct{ URL string }
	data := EmailData{
		URL: input.RedirectUrl + "?token=" + token,
	}

	if err := h.email.SendEmail(user.Email, "Verify email", "verify_email", data); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not send verification email")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user registered, please verify your email"})
}

// Login обрабатывает вход пользователя
func (h AuthRoute) Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.validate.Struct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input data")
	}

	var user models.User
	if err := h.db.Where("email = ? AND is_email_verified = ?", input.Email, true).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return fiber.NewError(fiber.StatusUnauthorized, "incorrect password")
	}

	jti := uuid.New()
	accessToken, err := h.jwt.GenerateAccessToken(h.config.SessionExpire, user.ID.String())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create access token")
	}

	refreshToken, err := h.jwt.GenerateRefreshToken(h.config.RefreshExpire, user.ID.String(), jti.String())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create refresh token")
	}

	user.Sessions = append(user.Sessions, models.Session{
		ID:        jti,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(h.config.RefreshExpire),
		Agent:     c.Get("User-Agent"),
	})

	if err := h.db.Save(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not save session")
	}

	return c.JSON(fiber.Map{"access_token": accessToken, "refresh_token": refreshToken})
}

// Logout обрабатывает выход пользователя
func (h AuthRoute) Logout(c *fiber.Ctx) error {
	type RefreshInput struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	var input RefreshInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.validate.Struct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input data")
	}

	token, err := h.jwt.ValidateToken(input.RefreshToken)
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	claims := token.Claims.(utils.JwtCustomClaim)
	if err := h.db.Where("JTI = ?", claims.ID).Delete(&models.Session{}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not logout user")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Refresh обрабатывает обновление access токена
func (h AuthRoute) Refresh(c *fiber.Ctx) error {
	type RefreshInput struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	var input RefreshInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.validate.Struct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input data")
	}

	token, err := h.jwt.ValidateToken(input.RefreshToken)
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	claims := token.Claims.(utils.JwtCustomClaim)
	sessionId := claims.ID

	var session models.Session
	if err := h.db.Preload("User").Where("id = ?", sessionId).First(&session).Error; err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	newAccessToken, err := h.jwt.GenerateAccessToken(h.config.SessionExpire, session.UserID.String())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not refresh access token")
	}

	return c.JSON(fiber.Map{"access_token": newAccessToken})
}

// ResetPassword обрабатывает запрос на сброс пароля
func (h AuthRoute) ResetPassword(c *fiber.Ctx) error {
	type ResetInput struct {
		Email       string `json:"email" validate:"required,email"`
		RedirectUrl string `json:"redirect_url" validate:"required"`
	}

	var input ResetInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.validate.Struct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input data")
	}

	var user models.User
	if err := h.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	token := uuid.New().String()
	verification := models.Verification{
		Type:      "EMAIL_VERIFY",
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(h.config.VerificationExpire),
	}

	if err := h.db.Create(&verification).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create verification")
	}

	type ResetData struct{ URL string }
	data := ResetData{
		URL: input.RedirectUrl + "?token=" + token,
	}

	if err := h.email.SendEmail(user.Email, "Verify email", "reset_password", data); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not send verification email")
	}

	return c.JSON(fiber.Map{"message": "password reset email sent"})
}

// VerifyPasswordReset обрабатывает проверку токена сброса пароля
func (h AuthRoute) VerifyPasswordReset(c *fiber.Ctx) error {
	type VerifyInput struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
	}

	var input VerifyInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.validate.Struct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input data")
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not start transaction")
	}

	var verification models.Verification
	if err := tx.
		Where("token = ?", input.Token).
		Where("type = ?", "PASSWORD_RESET").
		First(&verification).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusNotFound, "verification token not found")
	}

	if verification.ExpiresAt.Before(time.Now()) {
		tx.Rollback()
		return fiber.NewError(fiber.StatusUnauthorized, "verification token expired")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "could not hash password")
	}

	verification.User.Password = hashedPassword
	if err := tx.Save(&verification.User).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "could not reset password")
	}

	if err := tx.Delete(&verification).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete verification token")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not commit transaction")
	}

	return c.JSON(fiber.Map{"message": "password reseted successfully"})
}

// VerifyEmail обрабатывает проверку электронной почты
func (h AuthRoute) VerifyEmail(c *fiber.Ctx) error {
	type VerifyInput struct {
		Token string `json:"token" validate:"required"`
	}

	var input VerifyInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.validate.Struct(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input data")
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not start transaction")
	}

	var verification models.Verification
	if err := tx.
		Where("token = ? AND type = ? AND expires_at > ?", input.Token, "EMAIL_VERIFY", time.Now()).
		Preload("User").
		First(&verification).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusNotFound, "verification token not found or expired")
	}

	if err := tx.Model(&verification.User).Update("is_email_verified", true).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify email")
	}

	if err := tx.Delete(&verification).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete verification token")
	}

	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not commit transaction")
	}

	return c.JSON(fiber.Map{"message": "email verified"})
}
