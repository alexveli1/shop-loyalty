package config

import (
	"time"

	"github.com/caarlos0/env/v6"

	mylog "github.com/alexveli/diploma/pkg/log"
)

type (
	Config struct {
		Postgres PostgresConfig
		Accrual  AccrualConfig
		Server   HTTPServerConfig
		Client   HTTPClientConfig
		Hash     HashConfig
		Auth     AuthConfig
	}
	PostgresConfig struct {
		DatabaseURI string        `env:"DATABASE_URI,required" envDefault:"postgres://user:1234567890qwerty@localhost:5432/gophermart"`
		Timeout     time.Duration `env:"DATABASE_TIMEOUT" envDefault:"30s"`
	}
	HTTPClientConfig struct {
		AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS,required" envDefault:"127.0.0.1:8081"`
		AccrualSystemGetRoot string        `env:"ACCRUAL_URL,required" envDefault:"/api/orders/"`
		RetryInterval        time.Duration `env:"RETRY_INTERVAL,required" envDefault:"1s"`
		RetryLimit           int           `env:"RETRY_LIMIT,required" envDefault:"10"`
	}
	AccrualConfig struct {
		SendInterval time.Duration `env:"SEND_INTERVAL" envDefault:"1s"`
	}
	HTTPServerConfig struct {
		RunAddress string `env:"RUN_ADDRESS,required" envDefault:"127.0.0.1:8080"`
	}
	HashConfig struct {
		Key string `env:"KEY" envDefault:"j3n4b%21&#"`
	}
	AuthConfig struct {
		JWT                    JWTConfig
		PasswordSalt           string `env:"SALT,unset" envDefault:"kjH^#(876320"`
		VerificationCodeLength int    `env:"VERIFICATION_CODE_LEN" envDefault:"8"`
	}
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
