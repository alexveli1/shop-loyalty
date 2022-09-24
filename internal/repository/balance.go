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

type BalanceRepo struct {
	conn *pgxpool.Pool
}

func NewBalanceRepo(db *pgxpool.Pool) *BalanceRepo {
	return &BalanceRepo{
		conn: db,
	}
}
func (a *BalanceRepo) GetAccountBalance(ctx context.Context, userid int64) (*proto.Balance, error) {
	selectBalance := "SELECT current, withdrawn FROM balances WHERE userid=$1"
	row := a.conn.QueryRow(ctx, selectBalance, userid)
	var balance proto.Balance
	var current, withdrawn sql.NullFloat64
	err := row.Scan(&current, &withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Infof("no balance record for the user, %v", err)

			return &proto.Balance{}, nil
		}
		mylog.SugarLogger.Errorf("cannot scan balance into variable, %v", err)

		return &proto.Balance{}, err
	}
	if current.Valid {
		balance.Current = current.Float64
	}
	if withdrawn.Valid {
		balance.Withdrawn = withdrawn.Float64
	}
	return &balance, nil
}

func (a *BalanceRepo) Withdraw(ctx context.Context, withdraw *proto.Withdraw) error {
	selectBalance := "SELECT current, withdrawn FROM balances WHERE userid = $1"
	balanceInsert := "INSERT INTO balances (userid, current, withdrawn) VALUES ($1, $2, $3)"
	balanceUpdate := " ON CONFLICT (userid) DO UPDATE SET current=EXCLUDED.current, withdrawn=EXCLUDED.withdrawn"
	insertWithdrawal := "INSERT INTO withdrawals (orderid, userid, sum, processed_at) VALUES ($1,$2,$3,$4)"
	tx, err := a.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		mylog.SugarLogger.Errorf("cannot create transaction, %v", err)

		return err
	}
	defer func() {
		if r := recover(); r != nil {
			mylog.SugarLogger.Errorf("unexpected panic, %v", r)
		}
	}()
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				mylog.SugarLogger.Errorf("cannot rollback transaction, %v", err)

				return
			}
		}
		err := tx.Commit(ctx)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot commit transaction, %v", err)

			return
		}
	}()
	if err != nil {
		mylog.SugarLogger.Errorf("cannot conver order string to int64, %v", err)

		return err
	}
	var balance proto.Balance
	var current, withdrawn sql.NullFloat64
	selectedBalance := tx.QueryRow(
		ctx,
		selectBalance,
		withdraw.Userid,
	)
	err = selectedBalance.Scan(&current, &withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Errorf("no records for user in balance, %v", err)

			return domain.ErrBalanceNoBalance
		}
		mylog.SugarLogger.Errorf("cannot scan balance, %v", err)

		return err
	}
	if current.Valid {
		balance.Current = current.Float64
	}
	if withdrawn.Valid {
		balance.Withdrawn = withdrawn.Float64
	}

	if withdraw.Sum > balance.Current || !current.Valid {
		mylog.SugarLogger.Errorf("cannot withdraw - current points amount is less than withdrawal sum, %v", err)

		return domain.ErrBalanceNotEnoughPoints
	}
	newCurrent := balance.Current - withdraw.Sum
	newWithdrawn := balance.Withdrawn + withdraw.Sum
	_, err = tx.Exec(
		ctx,
		balanceInsert+balanceUpdate,
		withdraw.Userid,
		newCurrent,
		newWithdrawn,
	)
	if err != nil {
		mylog.SugarLogger.Errorf("error when processing insert/update into balances, %v", err)

		return err
	}

	_, err = tx.Exec(
		ctx,
		insertWithdrawal,
		withdraw.Order,
		withdraw.Userid,
		withdraw.Sum,
		time.Now(),
	)
	if err != nil {
		mylog.SugarLogger.Errorf("error when processing insert into withdrawals, %v", err)

		return err
	}
	return nil
}

func (a *BalanceRepo) GetAccountWithdrals(ctx context.Context, userid int64) ([]proto.Withdraw, error) {
	selectWithdrawls := "SELECT orderid, sum, processed_at FROM withdrawals WHERE userid = $1 ORDER BY processed_at DESC"
	rows, err := a.conn.Query(ctx, selectWithdrawls, userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Errorf("no withdrawls for the user, %v", err)

			return []proto.Withdraw{}, domain.ErrBalanceNoWithdrawls
		}
		mylog.SugarLogger.Errorf("cannot query orders for withdrawals, %v", err)

		return []proto.Withdraw{}, err
	}
	var withdrawal proto.Withdraw
	var withdrawSum sql.NullFloat64
	var orderID sql.NullInt64
	var processedAt sql.NullTime
	var withdrawals []proto.Withdraw
	for rows.Next() {
		err := rows.Scan(&orderID, &withdrawSum, &processedAt)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot scan rows into withdrawal, %v", err)

			return []proto.Withdraw{}, err
		}
		if orderID.Valid {
			withdrawal.Order = fmt.Sprint(orderID.Int64)
		}
		if withdrawSum.Valid {
			withdrawal.Sum = withdrawSum.Float64
		}
		if processedAt.Valid {
			withdrawal.ProcessedAt = processedAt.Time.Local().Format(time.RFC3339)
		}
		withdrawals = append(withdrawals, proto.Withdraw{
			Order:       withdrawal.Order,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		})
	}

	return withdrawals, nil
}

func (a *BalanceRepo) DeleteAllBalances(ctx context.Context) {
	_, _ = a.conn.Exec(ctx, "TRUNCATE balances")
}
