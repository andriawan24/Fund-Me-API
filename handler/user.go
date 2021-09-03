package handler

import (
	"fund-me/helper"
	"fund-me/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userService user.Service
}

func NewUserHandler(userService user.Service) *userHandler {
	return &userHandler{userService}
}

func (h *userHandler) RegisterUser(c *gin.Context) {
	// Grab user input and Map user input to struct RegisterUserInput
	var input user.RegisterUserInput

	err := c.ShouldBindJSON(&input)
	if err != nil {
		errors := helper.FormatValidationError(err)
		errMessage := gin.H{"errors": errors}

		response := helper.APIResponse(
			"Register failed",
			http.StatusUnprocessableEntity,
			"error",
			errMessage,
		)
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	// Struct that has been made parsed into service
	newUser, err := h.userService.RegisterUser(input)

	if err != nil {
		response := helper.APIResponse(
			"Register failed",
			http.StatusBadRequest,
			"error",
			nil,
		)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// token, err := h.jwtService.Token

	formatter := user.FormatUser(newUser, "tokentokentokentokentokentoken")

	response := helper.APIResponse(
		"Account has been registered",
		http.StatusOK,
		"success",
		formatter,
	)

	c.JSON(http.StatusOK, response)
}
