package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/alexveli/diploma/internal/domain"
	proto2 "github.com/alexveli/diploma/internal/proto"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type OrderRepo struct {
	conn *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{
		conn: db,
	}
}
func (a *OrderRepo) GetOrder(orderid int64) (*proto2.Order, error) {
	return &proto2.Order{}, nil
}
func (a *OrderRepo) StoreOrder(order *proto2.Order) error {
	return nil
}

func (a *OrderRepo) GetAccountOrders(ctx context.Context, account *proto2.Account) ([]proto2.GetOrdersItem, error) {
	var uploadedAt sql.NullTime
	var ordAccrual sql.NullFloat64
	var ordNum sql.NullInt64
	selectLine1 := "SELECT orderid, status, "
	selectLIne2 := "accrualsum, "
	selectLine3 := "uploaded_at FROM orders WHERE userid=$1 ORDER BY uploaded_at"
	orders := make([]proto2.GetOrdersItem, 0)
	rows, err := a.conn.Query(ctx, selectLine1+selectLIne2+selectLine3, account.Userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Infof("no rows for the user %s found", account.Username)

			return []proto2.GetOrdersItem{}, domain.ErrOrderNoOrders
		}
		mylog.SugarLogger.Errorf("cannot get orders, %v", err)

		return []proto2.GetOrdersItem{}, err
	}
	for rows.Next() {
		ord := proto2.GetOrdersItem{}
		err := rows.Scan(
			&ordNum,
			&ord.Status,
			&ordAccrual,
			&uploadedAt,
		)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot scan row, %v", err)

			return []proto2.GetOrdersItem{}, err
		}
		if ordNum.Valid {
			ord.Number = fmt.Sprint(ordNum.Int64)
		}
		if ordAccrual.Valid {
			ord.Accrual = ordAccrual.Float64
		}
		if uploadedAt.Valid {
			ord.UploadedAt = fmt.Sprint(uploadedAt.Time.Local().Format(time.RFC3339))
		}
		orders = append(orders, ord)
	}
	return orders, nil
}

func (a *OrderRepo) DeleteAllOrders(ctx context.Context) {
	_, _ = a.conn.Exec(ctx, "TRUNCATE orders")
}
