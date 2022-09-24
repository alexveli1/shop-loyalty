package handlers

import (
	"github.com/gin-gonic/gin"

	proto2 "github.com/alexveli/diploma/internal/proto"
	mylog "github.com/alexveli/diploma/pkg/log"
)

func (h *Handler) GetAccountFromUsername(c *gin.Context) (*proto2.Account, error) {
	userid, err := h.tokenManager.ExtractUserIdFromToken(c)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get userid, %v", err)
		return &proto2.Account{}, err
	}
	account, err := h.services.Account.GetAccount(c.Request.Context(), userid)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)
		return &proto2.Account{}, err
	}
	return account, nil
}
