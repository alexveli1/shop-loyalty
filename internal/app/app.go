package app

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexveli/diploma/internal/config"
	repository2 "github.com/alexveli/diploma/internal/repository"
	"github.com/alexveli/diploma/internal/service"
	"github.com/alexveli/diploma/internal/transport/httpv1/client"
	"github.com/alexveli/diploma/internal/transport/httpv1/handlers"
	"github.com/alexveli/diploma/pkg/auth"
	"github.com/alexveli/diploma/pkg/hash"
	mylog "github.com/alexveli/diploma/pkg/log"
	"github.com/alexveli/diploma/pkg/postgres"
)

func Run() {
	mylog.SugarLogger = mylog.InitLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Config{}
	_, err := config.NewConfig(&cfg)
	if err != nil {
		mylog.SugarLogger.Fatalf("Cannot get config: %v", err)
	}
	mylog.SugarLogger.Infof("Config set: %v", cfg)

	setFlags(&cfg)
	mylog.SugarLogger.Infof("Flags set: %v", cfg)

	db, err := postgres.NewPostgresDB(cfg.Postgres.DatabaseURI)
	if err != nil {
		mylog.SugarLogger.Fatalf("Cannot connect to db: %v", err)
	}
	mylog.SugarLogger.Infof("DB set: %v", db)

	repos := repository2.NewRepositories(db)
	mylog.SugarLogger.Infof("Repos set: %v", repos)

	//initiate database
	dbmanager := repository2.NewDBCreator(db)
	err = dbmanager.CreateTables(ctx)
	if err != nil {
		mylog.SugarLogger.Fatalf("cannot generate tables in database, %v", err)
	}
	//delete table contents in case tables already exist
	dbmanager.DeleteTableCotents(ctx)

	tokenManager, err := auth.NewManager(cfg.JWT)
	if err != nil {
		mylog.SugarLogger.Fatalf("Cannot initiate tockenManager: %v", err)
	}
	mylog.SugarLogger.Infof("TokenManager set: %v", tokenManager)

	hasher := hash.NewHasher(cfg.Server.HashKey)
	mylog.SugarLogger.Infof("Hasher set: %v", hasher)

	accrualHTTPClient := client.NewAccrualHTTPClient(cfg.Client.AccrualSystemAddress, cfg.Client.AccrualSystemGetRoot, cfg.Client.RetryInterval, cfg.Client.RetryLimit)
	mylog.SugarLogger.Infof("Client set: %v", accrualHTTPClient)

	services := service.NewServices(repos, cfg.Client.SendInterval, accrualHTTPClient)
	mylog.SugarLogger.Infof("Services set: %v", services)

	go services.Accrual.SendToAccrual(ctx)

	handler := handlers.NewHandler(services, tokenManager, *hasher)
	mylog.SugarLogger.Infof("Handlers set: %v", handler)

	//srv := server.NewServer(&cfg, handler.Init(&cfg))

	srv := &http.Server{
		Addr:    cfg.Server.RunAddress,
		Handler: handler.Init(&cfg),
	}

	quit := make(chan os.Signal, 1)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			mylog.SugarLogger.Errorf("error occurred while running http server: %s\n", err.Error())
			quit <- syscall.SIGTERM
		}
	}()
	// Graceful Shutdown
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	<-quit
	mylog.SugarLogger.Info("terminating")

	ctx, shutdown := context.WithTimeout(ctx, cfg.Server.TerminateTimeout)
	defer shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		mylog.SugarLogger.Errorf("failed to stop server: %v", err)
	}

}

func setFlags(cfg *config.Config) {
	flag.StringVar(&cfg.Client.AccrualSystemAddress, "r", cfg.Client.AccrualSystemAddress, "address for starting accrual system instance")
	flag.StringVar(&cfg.Server.RunAddress, "a", cfg.Server.RunAddress, "address for starting gophermart")
	flag.StringVar(&cfg.Postgres.DatabaseURI, "d", cfg.Postgres.DatabaseURI, "database connection string")
	flag.Parse()
}
