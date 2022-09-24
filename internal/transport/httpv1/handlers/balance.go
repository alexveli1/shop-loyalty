package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
	"github.com/alexveli/diploma/pkg/check"
	mylog "github.com/alexveli/diploma/pkg/log"
)

func (h *Handler) GetBalance(c *gin.Context) {
	account, err := h.GetAccountFromUsername(c)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)
		newResponse(c, http.StatusUnauthorized, err.Error())

		return
	}
	balance, err := h.services.Balance.GetAccountBalance(c.Request.Context(), account)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	newResponse(c, http.StatusOK, balance)
}

func (h *Handler) Withdraw(c *gin.Context) {
	var withdraw proto.Withdraw
	if err := c.BindJSON(&withdraw); err != nil {
		mylog.SugarLogger.Errorf("invalid request format, %v", err)
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	account, err := h.GetAccountFromUsername(c)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)
		newResponse(c, http.StatusUnauthorized, err.Error())

		return
	}
	orderid, err := strconv.ParseInt(withdraw.Order, 10, 64)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot convert order number to int64, %v", err)
		newResponse(c, http.StatusUnprocessableEntity, err.Error())

		return
	}
	if !check.CheckOrderNumber(orderid) {
		mylog.SugarLogger.Errorf("order number format incorrect, %v", err)
		newResponse(c, http.StatusUnprocessableEntity, err.Error())

		return
	}
	withdraw.Userid = account.Userid
	err = h.services.Balance.Withdraw(c.Request.Context(), &withdraw)
	switch err {
	case domain.ErrBalanceNoBalance, domain.ErrBalanceNotEnoughPoints:
		mylog.SugarLogger.Errorf("not enough points, %v", err)
		newResponse(c, http.StatusPaymentRequired, err.Error())

		return
	case nil:
		mylog.SugarLogger.Infof("withdrawal successful")
		newResponse(c, http.StatusOK, "withdrawal was successful")

		return
	default:
		mylog.SugarLogger.Errorf("cannot execute withdrawal, %v", err)
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

}

func (h *Handler) GetWithdrawls(c *gin.Context) {
	account, err := h.GetAccountFromUsername(c)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)
		newResponse(c, http.StatusUnauthorized, err.Error())

		return
	}

	withdrawals, err := h.services.Balance.GetAccountWithdrawals(c.Request.Context(), account.Userid)
	switch err {
	case domain.ErrBalanceNoWithdrawls:
		mylog.SugarLogger.Errorf("no withdralws for the user, %v", err)
		newResponse(c, http.StatusNoContent, "no withdrawls for th user")

		return
	case nil:
		mylog.SugarLogger.Infoln("withdrawls got successfully")
		newResponse(c, http.StatusOK, withdrawals)

		return
	default:
		mylog.SugarLogger.Errorf("cannot get withdrawls, %v", err)
		newResponse(c, http.StatusInternalServerError, "internal server error")

		return
	}

}
