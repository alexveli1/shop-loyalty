package domain

import "errors"

var (
	ErrUserNotFound                     = errors.New("user doesn't exists")
	ErrUserAlreadyExists                = errors.New("user with such email already exists")
	ErrPasswordIncorrect                = errors.New("password incorrect")
	ErrSenderCannotSendRequestToAccrual = errors.New("failed to send request to accrual system after several retries")
	ErrAuthorizationInvalidToken        = errors.New("token invalid")
	ErrBalanceNoBalance                 = errors.New("no balance for user")
	ErrBalanceNotEnoughPoints           = errors.New("current points amount is less than withdraw amount")
	ErrBalanceNoWithdrawls              = errors.New("no withdrawls for the user")
	ErrOrderNoOrders                    = errors.New("no orders found for specified user")
	ErrUserIDInvalid                    = errors.New("after scanning statement result userid.Valid == false")
)
