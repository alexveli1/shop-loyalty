package domain

//sql requests for repos
const (
	//DBManager
	CreateTables = `create table if not exists orders
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
	DeleteTables = `TRUNCATE orders;
							TRUNCATE balances;
							TRUNCATE accounts;
							TRUNCATE withdrawals;`

	//Account
	AccountGetByNameOrID = "SELECT userid, username, passwordhash FROM accounts WHERE username = $1 OR userid = $2"
	InsertNewAccount     = "INSERT INTO accounts (username, passwordhash) VALUES($1,$2) ON CONFLICT DO NOTHING RETURNING userid"

	//Accrualer
	OrderInsertFromAccrual = "INSERT INTO orders (orderid, userid, status, accrualsum, processed_inaccrual_at) VALUES($1, $2, $3, $4, $5)"
	OrderUpdateFromAccrual = " ON CONFLICT (orderid) DO UPDATE SET status=EXCLUDED.status, accrualsum=EXCLUDED.accrualsum, processed_inaccrual_at=EXCLUDED.processed_inaccrual_at"
	OrderInsertStatus      = "INSERT INTO orders (orderid, userid, status, uploaded_at) VALUES($1, $2, $3, $4)"
	OrderUpdateStatus      = " ON CONFLICT (orderid) DO UPDATE SET status = EXCLUDED.status"
	OrderSelectByID        = "SELECT userid FROM orders WHERE orderid = $1"
	OrderSelectByStatus    = "SELECT orderid, userid FROM orders WHERE status = $1 LIMIT 1"
	BalanceInsert          = "INSERT INTO balances (userid, current) VALUES ($1, $2)"
	BalanceUpdate          = " ON CONFLICT (userid) DO UPDATE SET userid=EXCLUDED.userid, current=balances.current+EXCLUDED.current"

	//Balance
	WithdrawalGetByUserID  = "SELECT orderid, sum, processed_at FROM withdrawals WHERE userid = $1 ORDER BY processed_at DESC"
	BalanceGetByUserID     = "SELECT current, withdrawn FROM balances WHERE userid = $1"
	BalanceInsertWithdrawn = "INSERT INTO balances (userid, current, withdrawn) VALUES ($1, $2, $3)"
	BalanceUpdateWithdrawn = " ON CONFLICT (userid) DO UPDATE SET current=EXCLUDED.current, withdrawn=EXCLUDED.withdrawn"
	WithdrawalInsert       = "INSERT INTO withdrawals (orderid, userid, sum, processed_at) VALUES ($1, $2, $3, $4)"

	//Orders
	OrdersGetByUserID = `SELECT orderid, status, accrualsum, uploaded_at FROM orders WHERE userid = $1 ORDER BY uploaded_at`
)
