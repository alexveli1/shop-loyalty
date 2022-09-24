package domain

import "errors"

var (
	ErrUserNotFound                        = errors.New("user doesn't exists")
	ErrUserAlreadyExists                   = errors.New("user with such email already exists")
	ErrCannotStoreAccount                  = errors.New("cannot store user account")
	ErrPasswordIncorrect                   = errors.New("password incorrect")
	ErrOrderUploadedByCurrentUser          = errors.New("order already uploaded by current user")
	ErrOrderUploadedByAnotherUser          = errors.New("order already uploaded by another user")
	ErrSenderCannotSelectOrderToSend       = errors.New("sender client cannot select any order to send to accrual system")
	ErrSenderTooManyRequests               = errors.New("accrual system turned on throttling")
	ErrSenderCannotSendRequestToAccrual    = errors.New("failed to send request to accrual system after several retries")
	ErrSenderCannotUpdateOrderOrBalance    = errors.New("failed to update order or user balance with info from accrual system")
	ErrAccrualRepoCannotProcessTransaction = errors.New("some statements in transaction failed - rolling back")
	ErrAuthorizationInvalidToken           = errors.New("token invalid")
	ErrBalanceNoBalance                    = errors.New("no balance for user")
	ErrBalanceNotEnoughPoints              = errors.New("current points amount is less than withdraw amount")
	ErrBalanceNoWithdrawls                 = errors.New("no withdrawls for the user")
	ErrOrderAlreadyProcessed               = errors.New("order already processed")
	ErrOrderNoOrders                       = errors.New("no orders found for specified user")
)
