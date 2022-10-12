package service

import (
	"context"
	"time"

	proto2 "github.com/alexveli/diploma/internal/proto"
	"github.com/alexveli/diploma/internal/repository"
	"github.com/alexveli/diploma/internal/transport/httpv1/client"
)

type UserRegisterInput struct {
	Name     string
	Password string
}

type UserLoginInput struct {
	Name     string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Account interface {
	Register(ctx context.Context, input *proto2.Account) error
	Login(ctx context.Context, input *proto2.Account) error
	GetAccount(ctx context.Context, userid int64) (*proto2.Account, error)
}

type Order interface {
	GetAccountOrders(ctx context.Context, account *proto2.Account) ([]byte, error)
}

type Balance interface {
	GetAccountBalance(ctx context.Context, account *proto2.Account) (string, error)
	Withdraw(ctx context.Context, withdraw *proto2.Withdraw) error
	GetAccountWithdrawals(ctx context.Context, userid int64) (string, error)
}

type Accrual interface {
	CheckRequestFormat(content string) (int64, bool)
	CheckOrderAlreadyUploaded(ctx context.Context, orderid int64) (int64, bool, error)
	AddOrderToQueue(ctx context.Context, order *proto2.Order) bool
	GetFirstUnprocessedOrder(ctx context.Context) (*proto2.Order, bool)
	UpdateOrderAndBalance(ctx context.Context, accrualReply *proto2.Order)
	SendToAccrual(ctx context.Context)
}

type Services struct {
	Account Account
	Balance Balance
	Order   Order
	Accrual Accrual
}

func NewServices(repositories *repository.Repositories, interval time.Duration, client client.HTTPClient) *Services {
	return &Services{
		Account: NewAccountService(repositories.Account),
		Balance: NewBalanceService(repositories.Balance),
		Order:   NewOrderService(repositories.Order),
		Accrual: NewAccrualer(interval, repositories.Accrualer, client),
	}
}
