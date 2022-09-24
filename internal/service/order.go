package service

import (
	"context"
	"encoding/json"

	proto2 "github.com/alexveli/diploma/internal/proto"
	"github.com/alexveli/diploma/internal/repository"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type OrderService struct {
	repo repository.Order
}

func NewOrderService(repo repository.Order) *OrderService {
	return &OrderService{repo: repo}
}

func (o *OrderService) BindOrderToAccount(ctx context.Context, account *proto2.Account, orderid int64) error {
	order := proto2.Order{
		Userid:  account.Userid,
		Orderid: orderid,
	}
	err := o.repo.StoreOrder(&order)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot bind order to username, %v", err)
		return err
	}
	return nil
}

func (o *OrderService) GetAccountOrders(ctx context.Context, account *proto2.Account) ([]byte, error) {
	orders, err := o.repo.GetAccountOrders(ctx, account)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get orders, %v", err)
		return []byte{}, err
	}
	jsonOrders, err := json.Marshal(orders)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot marshal orders, %v", err)
		return []byte{}, err
	}
	return jsonOrders, nil
}
