package config

import (
	"time"

	"github.com/caarlos0/env/v6"

	mylog "github.com/alexveli/diploma/pkg/log"
)

type (
	Config struct {
		Postgres PostgresConfig
		Server   HTTPServerConfig
		Client   HTTPClientConfig
		JWT      JWTConfig
	}
	//PostgresConfig - connection string for opening connection
	PostgresConfig struct {
		DatabaseURI string `env:"DATABASE_URI" envDefault:"postgres://user:1234567890qwerty@localhost:5432/gophermart"`
	}

	//HTTPClientConfig - configuring connection of
	//http client to accrual system
	//for sending orders for getting accrual points
	HTTPClientConfig struct {
		AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
		AccrualSystemGetRoot string        `env:"ACCRUAL_URL,required" envDefault:"/api/orders/"`
		RetryInterval        time.Duration `env:"RETRY_INTERVAL,required" envDefault:"1s"` //RetryInterval - how often trying to connect to accrual system in case of network issues
		RetryLimit           int           `env:"RETRY_LIMIT,required" envDefault:"10"`    //RetryLimit - how many retries for network or unavailability issues
		SendInterval         time.Duration `env:"SEND_INTERVAL" envDefault:"1s"`           //SendInterval - how often gophermart tries to send order to accrual system
	}
	//HTTPServerConfig - config for starting http server
	HTTPServerConfig struct {
		RunAddress       string        `env:"RUN_ADDRESS"`                         //RunAddress - address to bind server
		HashKey          string        `env:"HASH_KEY" envDefault:"j3n4b%21&#"`    //HashKey - used for password hashing
		TerminateTimeout time.Duration `env:"TERMINATION_TIMEOUT" envDefault:"1s"` //TerminateTimeout - used for graceful shutdown of the server
	}
	//JWTConfig - configuration for tokens generation
	JWTConfig struct {
		AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
		RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"24h"`
		SigningKey      string        `env:"SIGNING_KEY" envDefault:"Ed1039%^&*3JS"`
	}
)

func NewConfig(cfg *Config) (*Config, error) {
	mylog.SugarLogger.Infoln("Init Config")
	if err := env.Parse(cfg); err != nil {
		mylog.SugarLogger.Errorf("%+v", err)
		return nil, err
	}
	mylog.SugarLogger.Infof("%+v", cfg)
	return cfg, nil
}
