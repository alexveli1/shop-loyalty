package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type AccountRepo struct {
	conn *pgxpool.Pool
}

func NewAccountRepo(db *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{
		conn: db,
	}
}

func (a *AccountRepo) GetAccount(ctx context.Context, login *proto.Account) (*proto.Account, error) {
	var account proto.Account
	row := a.conn.QueryRow(ctx, "SELECT userid, username, passwordhash FROM accounts WHERE username = $1 OR userid = $2", login.Username, login.Userid)
	err := row.Scan(&account.Userid, &account.Username, &account.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Errorf("no rows %v", err)

			return &proto.Account{}, domain.ErrUserNotFound
		}
		mylog.SugarLogger.Errorf("error when scanning values %v", err)

		return &proto.Account{}, err
	}

	return &account, nil
}

func (a *AccountRepo) StoreAccount(ctx context.Context, account *proto.Account) error {
	var userid sql.NullInt64
	selectAccount := "SELECT userid FROM accounts WHERE username=$1"
	tx, err := a.conn.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		if err != nil {
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
	if err != nil {
		mylog.SugarLogger.Errorf("cannot initiate transaction, %v", err)

		return err
	}
	row := tx.QueryRow(ctx, selectAccount, account.Username)
	err = row.Scan(&userid)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {

			mylog.SugarLogger.Errorf("cannot get account, %v", err)

			return err
		}
	}
	executionResult := a.conn.QueryRow(ctx, "INSERT INTO accounts (username, passwordhash) VALUES($1,$2) RETURNING userid", account.Username, account.PasswordHash)
	err = executionResult.Scan(&userid)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot store account: %v", err)

		return err
	}
	if userid.Valid {
		account.Userid = userid.Int64
	}
	mylog.SugarLogger.Infof("user %s successfully registered with id %d", account.Username, account.Userid)
	err = nil
	return err
}

func (a *AccountRepo) DeleteAllAccounts(ctx context.Context) {
	_, _ = a.conn.Exec(ctx, "TRUNCATE accounts")
}
