package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
	"github.com/alexveli/diploma/pkg/check"
	mylog "github.com/alexveli/diploma/pkg/log"
)

func (h *Handler) ProcessOrder(c *gin.Context) {
	byteBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot read request body, %v", err)
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	body := strings.TrimSpace(string(byteBody))
	accountFromToken, err := h.GetAccountFromUsername(c)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)
		newResponse(c, http.StatusUnauthorized, err.Error())

		return
	}
	orderid, requestFormatIsValid := h.services.Accrual.CheckRequestFormat(body)
	if !requestFormatIsValid {
		mylog.SugarLogger.Infof("request format is incorrect")
		newResponse(c, http.StatusBadRequest, "request format is incorrect")

		return
	}
	orderNumberCompliantWithLuhn := check.CheckOrderNumber(orderid)
	if !orderNumberCompliantWithLuhn {
		mylog.SugarLogger.Errorf("order number %d has not passed Lunh test", orderid)
		newResponse(c, http.StatusUnprocessableEntity, "order number has not passed Lunh test")

		return
	}
	orderCreatorUserID, orderAlreadyUploaded, err := h.services.Accrual.CheckOrderAlreadyUploaded(c.Request.Context(), orderid)
	if err != nil {
		mylog.SugarLogger.Errorf("error while trying to check order existance, %v", err)
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	if orderAlreadyUploaded {
		if bySameUser(orderCreatorUserID, accountFromToken.Userid) {
			mylog.SugarLogger.Infof("order already uploaded by current user, %v", err)
			newResponse(c, http.StatusOK, "order already uploaded by current user")

			return
		}
		mylog.SugarLogger.Infof("order already uploaded by another user, %v", err)
		newResponse(c, http.StatusConflict, "order already uploaded by another user")

		return
	}
	orderAddedToQueueSuccessfully := h.services.Accrual.AddOrderToQueue(
		c.Request.Context(),
		&proto.Order{Orderid: orderid, Userid: accountFromToken.Userid, Status: domain.NEW, UploadedAt: timestamppb.Now()},
	)
	if !orderAddedToQueueSuccessfully {
		mylog.SugarLogger.Errorf("sending to accrual was unsuccessful")
		newResponse(c, http.StatusInternalServerError, "sending to accrual was unsuccessful")

		return
	}
	mylog.SugarLogger.Infof("sending to accrual system initiated for order %d", orderid)
	newResponse(c, http.StatusAccepted, "sending to accrual system initiated for order "+fmt.Sprint(orderid))
}

func (h *Handler) GetOrders(c *gin.Context) {
	accountFromToken, err := h.GetAccountFromUsername(c)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)
		newResponse(c, http.StatusUnauthorized, err.Error())

		return
	}
	byteOrders, err := h.services.Order.GetAccountOrders(c.Request.Context(), accountFromToken)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNoOrders) {
			mylog.SugarLogger.Errorf("no orders found for user %s", accountFromToken.Username)
			newResponse(c, http.StatusNoContent, domain.ErrOrderNoOrders.Error())

			return
		}
		mylog.SugarLogger.Errorf("cannot get account orders, %v", err)
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}
	newResponse(c, http.StatusOK, byteOrders)
}

func bySameUser(u1 int64, u2 int64) bool {
	return u1 == u2
}
