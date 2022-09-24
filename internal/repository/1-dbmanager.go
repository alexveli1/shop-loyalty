package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
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

const creatorSQL = `create table if not exists orders
(
    id                     serial
        constraint orders_pk
            primary key,
    orderid                bigint  not null,
    userid                 integer not null,
    ordersum               double precision,
    accrualsum             double precision,
    withdrawsum            double precision,
    uploaded_at            timestamp,
    processed_inaccrual_at timestamp,
    status                 text
);

create unique index if not exists orders_orderid_uindex
    on orders (orderid);

create table if not exists balances
(
    id        serial
        constraint balances_pk
            primary key,
    userid    integer,
    current   double precision,
    withdrawn double precision
);

create unique index if not exists balances_userid_uindex
    on balances (userid);

create table if not exists accounts
(
    id           serial
        constraint accounts_pk
            primary key,
    userid       serial,
    username     text not null,
    passwordhash text not null
);

create unique index if not exists accounts_userid_uindex
    on accounts (userid);

create unique index if not exists accounts_username_uindex
    on accounts (username);

create table if not exists withdrawals
(
    withdrawal_id serial
        constraint withdrawals_pk
            primary key,
    orderid       bigint,
    sum           double precision,
    userid        integer,
    processed_at  timestamp
);`

func (d *DBCreator) CreateTables(ctx context.Context) error {
	_, err := d.db.Exec(ctx, creatorSQL)
	if err != nil {

		return err
	}

	return nil
}
