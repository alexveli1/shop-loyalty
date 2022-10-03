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
	"github.com/alexveli/diploma/internal/proto"
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

func (a *OrderRepo) GetAccountOrders(ctx context.Context, account *proto.Account) ([]proto.GetOrdersItem, error) {
	var uploadedAt sql.NullTime
	var ordAccrual sql.NullFloat64
	var ordNum sql.NullInt64
	orders := make([]proto.GetOrdersItem, 0)
	rows, err := a.conn.Query(ctx, domain.OrdersGetByUserID, account.Userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Infof("no rows for the user %s found", account.Username)

			return []proto.GetOrdersItem{}, domain.ErrOrderNoOrders
		}
		mylog.SugarLogger.Errorf("cannot get orders, %v", err)

		return []proto.GetOrdersItem{}, err
	}
	for rows.Next() {
		ord := proto.GetOrdersItem{}
		err := rows.Scan(
			&ordNum,
			&ord.Status,
			&ordAccrual,
			&uploadedAt,
		)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot scan row, %v", err)

			return []proto.GetOrdersItem{}, err
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
		orders = append(orders, proto.GetOrdersItem{
			Number:     ord.Number,
			Status:     ord.Status,
			Accrual:    ord.Accrual,
			UploadedAt: ord.UploadedAt,
		})
	}
	return orders, nil
}
