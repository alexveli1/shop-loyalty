package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
)

func (h *Handler) UserRegister(c *gin.Context) {
	var input proto.InputAccount
	if err := c.BindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	account := proto.Account{
		Username:     input.Login,
		PasswordHash: h.hasher.GetHash(input.Password),
	}
	err := h.services.Account.Register(
		c.Request.Context(),
		&account,
	)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			newResponse(c, http.StatusBadRequest, err.Error())

			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	token, err := h.tokenManager.GenerateToken(account.Userid)

	c.Header("Authorization", token)
	newResponse(c, http.StatusOK, "user "+account.Username+" successfully registered")
}

func (h *Handler) UserLogin(c *gin.Context) {
	var input proto.InputAccount
	if err := c.BindJSON(&input); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	account := proto.Account{
		Username:     input.Login,
		PasswordHash: h.hasher.GetHash(input.Password),
	}
	err := h.services.Account.Login(
		c.Request.Context(),
		&account,
	)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			newResponse(c, http.StatusBadRequest, err.Error())

			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	token, err := h.tokenManager.GenerateToken(account.Userid)

	c.Header("Authorization", token)
	newResponse(c, http.StatusOK, "user logged")
}
