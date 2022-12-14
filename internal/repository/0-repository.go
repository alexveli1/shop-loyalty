package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	proto2 "github.com/alexveli/diploma/internal/proto"
)

type Account interface {
	GetAccount(ctx context.Context, login *proto2.Account) (*proto2.Account, error)
	StoreAccount(ctx context.Context, account *proto2.Account) error
}

type Order interface {
	GetAccountOrders(ctx context.Context, account *proto2.Account) ([]proto2.GetOrdersItem, error)
}

type Balance interface {
	GetAccountBalance(ctx context.Context, userid int64) (*proto2.Balance, error)
	GetAccountWithdrals(ctx context.Context, userid int64) ([]proto2.Withdraw, error)
	Withdraw(ctx context.Context, withdraw *proto2.Withdraw) error
}

type Accrualer interface {
	InsertOrUpdateOrder(ctx context.Context, order *proto2.Order) bool
	IncreaseOrderAccrualAndBalanceCurrent(ctx context.Context, order *proto2.Order)
	CheckOrderAlreadyUploaded(ctx context.Context, orderid int64) (int64, bool, error)
	GetFirstUnprocessedOrder(ctx context.Context) (*proto2.Order, bool)
}

type Repositories struct {
	Account   Account
	Balance   Balance
	Order     Order
	Accrualer Accrualer
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		Account:   NewAccountRepo(db),
		Balance:   NewBalanceRepo(db),
		Order:     NewOrderRepo(db),
		Accrualer: NewAccrualerRepo(db),
	}
}
