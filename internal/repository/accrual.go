package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type AccrualerRepo struct {
	conn *pgxpool.Pool
}

func NewAccrualerRepo(db *pgxpool.Pool) *AccrualerRepo {
	return &AccrualerRepo{
		conn: db,
	}
}
func (a *AccrualerRepo) InsertOrUpdateOrder(ctx context.Context, order *proto.Order) bool {
	_, err := a.conn.Exec(
		ctx,
		domain.OrderInsertStatus+domain.OrderUpdateStatus,
		order.Orderid,
		order.Userid,
		order.Status,
		order.UploadedAt.AsTime(),
	)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot update order %d, error:%v", order.Orderid, err)
		return false
	}
	mylog.SugarLogger.Infof("order %d updated successfully", order.Orderid)
	return true
}

func (a *AccrualerRepo) IncreaseOrderAccrualAndBalanceCurrent(ctx context.Context, order *proto.Order) {
	tx, err := a.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		tx = nil
		mylog.SugarLogger.Errorf("cannot create transaction, %v", err)

		return
	}
	defer func() {
		if r := recover(); r != nil {
			mylog.SugarLogger.Errorf("unexpected error, %v", recover())
		}
	}()

	defer defer_tx(ctx, tx, err)

	_, err = tx.Exec(
		ctx,
		domain.OrderInsertFromAccrual+domain.OrderUpdateFromAccrual,
		order.Orderid,
		order.Userid,
		order.Status,
		order.Accrualsum,
		order.ProcessedByAccrualAt.AsTime(),
	)
	if err != nil {
		mylog.SugarLogger.Errorf("error when processing insert/update into orders, %v", err)

		return
	}
	_, err = tx.Exec(
		ctx,
		domain.BalanceInsert+domain.BalanceUpdate,
		order.Userid,
		order.Accrualsum,
	)
	if err != nil {
		mylog.SugarLogger.Errorf("error when processing insert/update into balances, %v", err)

		return
	}
}

func (a *AccrualerRepo) CheckOrderAlreadyUploaded(ctx context.Context, orderid int64) (int64, bool, error) {
	var userid int64
	row := a.conn.QueryRow(ctx, domain.OrderSelectByID, orderid)
	err := row.Scan(&userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Infof("no users ever uploaded this order %d", orderid)

			return 0, false, nil
		}
		mylog.SugarLogger.Errorf("unable to get userid from orders, %v", err)

		return 0, false, err
	}

	return userid, true, nil
}

func (a *AccrualerRepo) GetFirstUnprocessedOrder(ctx context.Context) (*proto.Order, bool) {
	row := a.conn.QueryRow(ctx, domain.OrderSelectByStatus, domain.NEW)
	var order proto.Order
	err := row.Scan(&order.Orderid, &order.Userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Infof("no rows to send to accrual system, %v", err)

			return &proto.Order{}, false
		}
		mylog.SugarLogger.Errorf("error selecting single orderid to send to accrual system, %v", err)

		return &proto.Order{}, false
	}
	return &order, true
}
