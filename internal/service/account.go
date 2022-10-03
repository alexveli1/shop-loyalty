package service

import (
	"context"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
	"github.com/alexveli/diploma/internal/repository"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type AccountService struct {
	repo repository.Account
}

func NewAccountService(repo repository.Account) *AccountService {
	return &AccountService{repo: repo}
}

func (a *AccountService) Register(ctx context.Context, account *proto.Account) error {
	err := a.repo.StoreAccount(ctx, account)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot store user account: %v", err)

		return err
	}

	return nil
}

func (a *AccountService) Login(ctx context.Context, input *proto.Account) error {

	login, err := a.repo.GetAccount(ctx, input)
	if err != nil {

		return domain.ErrUserNotFound
	}
	err = VerifyPassword(input.PasswordHash, login.PasswordHash)
	if err != nil {

		return domain.ErrPasswordIncorrect
	}
	input.Userid = login.Userid
	return nil
}

func (a *AccountService) GetAccount(ctx context.Context, userid int64) (*proto.Account, error) {
	lookup := proto.Account{
		Userid:       userid,
		PasswordHash: "",
	}
	account, err := a.repo.GetAccount(ctx, &lookup)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot get account, %v", err)
		return &proto.Account{}, err
	}
	return account, nil

}

func VerifyPassword(pwd1 string, pwd2 string) error {
	if pwd1 != pwd2 {

		return domain.ErrPasswordIncorrect
	}

	return nil
}
