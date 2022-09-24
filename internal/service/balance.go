package service

import (
	"context"
	"encoding/json"
	"strings"

	proto2 "github.com/alexveli/diploma/internal/proto"
	"github.com/alexveli/diploma/internal/repository"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type BalanceService struct {
	repo repository.Balance
}

func NewBalanceService(repo repository.Balance) *BalanceService {
	return &BalanceService{repo: repo}
}

func (b *BalanceService) GetAccountBalance(ctx context.Context, account *proto2.Account) (string, error) {

	balance, err := b.repo.GetAccountBalance(ctx, account.Userid)
	if err != nil {
		mylog.SugarLogger.Infof("cannot get account balance, %v", err)

		return "", err
	}
	buf, err := json.Marshal(balance)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot marshal balance, %v", err)

		return "", err
	}
	return string(buf), nil
}

func (b *BalanceService) Withdraw(ctx context.Context, withdraw *proto2.Withdraw) error {
	return b.repo.Withdraw(ctx, withdraw)
}

func (b *BalanceService) GetAccountWithdrawals(ctx context.Context, userid int64) (string, error) {
	withdrals, err := b.repo.GetAccountWithdrals(ctx, userid)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account balance, %v", err)

		return "", err
	}
	jsonWithdawals, err := json.Marshal(withdrals)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot marshal withdrawls, %v", err)

		return "", err
	}
	withdrawalsString := strings.ReplaceAll(string(jsonWithdawals), "\n", "")
	return withdrawalsString, nil
}
