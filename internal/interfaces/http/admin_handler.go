package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/the-go-dragons/final-project2/internal/domain"
	"github.com/the-go-dragons/final-project2/internal/usecase"
)

type AdminHandler struct {
	userUsecase *usecase.UserUsecase
	smsService  *usecase.SmsServiceImpl
}

func NewAdminHandler(userUsecase *usecase.UserUsecase, smsService *usecase.SmsServiceImpl) *AdminHandler {
	return &AdminHandler{
		userUsecase: userUsecase,
		smsService:  smsService,
	}
}

func (ah *AdminHandler) DisableUser(c echo.Context) error {
	admin := c.Get("user").(domain.User)

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Message: "Invalid user id"})
	}

	user, err := ah.userUsecase.GetUserById(uint(userId))
	if err != nil || user.ID <= 0 {
		return c.JSON(http.StatusBadRequest, Response{Message: "User not found"})
	}
	if !user.IsActive {
		return c.JSON(http.StatusBadRequest, Response{Message: "User already disabled"})
	}
	if user.ID == admin.ID {
		return c.JSON(http.StatusBadRequest, Response{Message: "Can not disable your self"})
	}
	if user.IsAdmin {
		return c.JSON(http.StatusBadRequest, Response{Message: "Can not disable other admins"})
	}

	user.IsActive = false
	ah.userUsecase.Update(user)

	return c.JSON(http.StatusOK, Response{Message: "User disabled successfully"})
}

type SMSHistoryResponse struct {
	Count        int                 `json:"count"`
	SMSHistories []domain.SMSHistory `json:"smsHistories"`
}

func (ah *AdminHandler) GetSMSHistoryByUserId(c echo.Context) error {
	_ = c.Get("user").(domain.User)

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Message: "Invalid user id"})
	}

	user, err := ah.userUsecase.GetUserById(uint(userId))
	if err != nil || user.ID <= 0 {
		return c.JSON(http.StatusBadRequest, Response{Message: "User not found"})
	}
	smsHistories, err := ah.smsService.GetSMSHistoryByUserId(user.ID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Message: "Could not get the sms histories: " + err.Error()})
	}

	return c.JSON(http.StatusOK, SMSHistoryResponse{Count: len(smsHistories), SMSHistories: smsHistories})
}
