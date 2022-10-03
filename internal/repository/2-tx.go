package repository

import (
	"context"

	"github.com/jackc/pgx/v4"

	mylog "github.com/alexveli/diploma/pkg/log"
)

func deferTx(ctx context.Context, tx pgx.Tx, err error) {
	if tx == nil {

		return
	}
	if err != nil {
		err := tx.Rollback(ctx)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot rollback transaction, %v", err)

			return
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot commit transaction, %v", err)

		return
	}
}
