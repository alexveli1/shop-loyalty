package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	mylog "github.com/alexveli/diploma/pkg/log"
)

func NewPostgresDB(uri string) (*pgxpool.Pool, error) {
	db, err := pgxpool.Connect(context.Background(), uri)
	if err != nil {
		mylog.SugarLogger.Errorf("Cannot connect to postgresdb:%v", err)
		return nil, err
	}
	return db, nil
}
