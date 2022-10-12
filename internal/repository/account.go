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
	row := a.conn.QueryRow(ctx, domain.AccountGetByNameOrID, login.Username, login.Userid)
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
	executionResult := a.conn.QueryRow(ctx, domain.InsertNewAccount, account.Username, account.PasswordHash)
	err := executionResult.Scan(&userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			mylog.SugarLogger.Errorf("user already exists: %v", err)

			return domain.ErrUserAlreadyExists
		}
		mylog.SugarLogger.Errorf("cannot insert into accounts table: %v", err)

		return err
	}
	if !userid.Valid {
		mylog.SugarLogger.Errorf("error scanning userid - userid not valid")

		return domain.ErrUserIDInvalid
	}
	account.Userid = userid.Int64
	mylog.SugarLogger.Infof("user %s successfully registered with id %d", account.Username, account.Userid)
	err = nil
	return err
}
