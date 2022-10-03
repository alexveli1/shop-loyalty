package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/alexveli/diploma/internal/domain"
)

type DBManager interface {
	CreateTables(ctx context.Context) error
}

type DBCreator struct {
	db *pgxpool.Pool
}

func NewDBCreator(db *pgxpool.Pool) *DBCreator {
	return &DBCreator{db: db}
}

func (d *DBCreator) CreateTables(ctx context.Context) error {
	_, err := d.db.Exec(ctx, domain.CreateTables)
	if err != nil {

		return err
	}

	return nil
}

func (d *DBCreator) DeleteTableCotents(ctx context.Context) {
	_, _ = d.db.Exec(ctx, domain.DeleteTables)
}
