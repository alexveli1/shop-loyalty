package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	domain2 "github.com/alexveli/diploma/internal/domain"
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
	statusInsert := "INSERT INTO orders (orderid, userid, status, uploaded_at) VALUES($1, $2, $3, $4)"
	statusUpdate := " ON CONFLICT (orderid) DO UPDATE SET status=EXCLUDED.status"
	_, err := a.conn.Exec(
		ctx,
		statusInsert+statusUpdate,
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

func (a *AccrualerRepo) IncreaseOrderAccrualAndBalanceCurrent(ctx context.Context, order *proto.Order) error {
	var txErr error
	orderInsert := "INSERT INTO orders (orderid, userid, status, accrualsum, processed_inaccrual_at) VALUES($1, $2, $3, $4, $5)"
	orderUpdate := " ON CONFLICT (orderid) DO UPDATE SET status=EXCLUDED.status, accrualsum=EXCLUDED.accrualsum, processed_inaccrual_at=EXCLUDED.processed_inaccrual_at"
	balanceInsert := "INSERT INTO balances (userid, current) VALUES ($1, $2)"
	balanceUpdate := " ON CONFLICT (userid) DO UPDATE SET userid=EXCLUDED.userid, current=balances.current+EXCLUDED.current"
	tx, err := a.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		mylog.SugarLogger.Errorf("cannot create transaction, %v", err)

		return err
	}
	defer func() {
		if r := recover(); r != nil {
			mylog.SugarLogger.Errorf("unexpected error, %v", recover())
		}
	}()
	defer func() {
		if txErr != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				mylog.SugarLogger.Errorf("cannot rollback transaction, %v", err)

				return
			}
		} else {
			err := tx.Commit(ctx)
			if err != nil {
				mylog.SugarLogger.Errorf("cannot commit transaction, %v", err)

				return
			}
		}
	}()
	_, err1 := tx.Exec(
		ctx,
		orderInsert+orderUpdate,
		order.Orderid,
		order.Userid,
		order.Status,
		order.Accrualsum,
		order.ProcessedByAccrualAt.AsTime(),
	)
	if err1 != nil {
		mylog.SugarLogger.Errorf("error when processing insert/update into orders, %v", err1)
	}
	_, err2 := tx.Exec(
		ctx,
		balanceInsert+balanceUpdate,
		order.Userid,
		order.Accrualsum,
	)
	if err1 != nil {
		mylog.SugarLogger.Errorf("error when processing insert/update into balances, %v", err1)
	}
	if err1 != nil || err2 != nil {
		txErr = domain2.ErrAccrualRepoCannotProcessTransaction
	}

	return txErr
}

func (a *AccrualerRepo) CheckOrderAlreadyUploaded(ctx context.Context, orderid int64) (int64, bool, error) {
	stmt := "SELECT userid FROM orders WHERE orderid=$1"
	var userid int64
	row := a.conn.QueryRow(ctx, stmt, orderid)
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
	sel := "SELECT orderid, userid FROM orders WHERE status='" + domain2.NEW + "' LIMIT 1"
	row := a.conn.QueryRow(ctx, sel)
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
