package http

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/the-go-dragons/final-project2/internal/domain"
	"github.com/the-go-dragons/final-project2/pkg/config"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func GenerateToken(user domain.User) (string, error) {
	expirationHoursCofig := config.Config.Jwt.Token.Expire.Hours
	JwtTokenSecretConfig := config.Config.Jwt.Token.Secret.Key

	duration := time.Duration(expirationHoursCofig) * time.Hour
	expirationTime := time.Now().Add(duration)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    expirationTime.Unix(),
	})

	secretKey := []byte(JwtTokenSecretConfig)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (uh userHandler) Login(c echo.Context) error {
	var request LoginRequest
	var user domain.User

	// Check the body data
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Message: "Invalid body request"})
	}
	if request.Username == "" || request.Password == "" {
		return c.JSON(http.StatusBadRequest, Response{Message: "Missing required fields"})
	}

	// Check for existence of user
	user, err = uh.userUsecase.GetUserByUsername(request.Username)
	if err != nil {
		return c.JSON(http.StatusNotFound, Response{Message: "No user found with this credentials"})
	}
	if !user.IsActive {
		return c.JSON(http.StatusNotFound, Response{Message: "User not found"})
	}

	// Check if password is correct
	equalErr := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(request.Password),
	)
	if equalErr == nil {
		// Generate the token
		token, err := GenerateToken(user)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{Message: "Server Error"})
		}

		// update IsLoginRequired field
		user.IsLoginRequired = false
		uh.userUsecase.Update(user)

		return c.JSON(http.StatusOK, LoginResponse{Message: "You logged in successfully", Token: token})
	}

	return c.JSON(http.StatusConflict, Response{Message: "No user found with this credentials"})
}
