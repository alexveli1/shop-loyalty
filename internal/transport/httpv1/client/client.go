package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type AccrualHTTPClient struct {
	accrualSystemAddress string
	accrualGetRoot       string
	retryInterval        time.Duration
	retryLimit           int
	client               *http.Client
}

type HTTPClient interface {
	SendToAccrual(ctx context.Context, orderid int64) (*proto.AccrualReply, error)
	SetRetryInterval(accrualRetryInterval string) time.Duration
}

func NewAccrualHTTPClient(addr string, url string, retryInterval time.Duration, retryLimit int) *AccrualHTTPClient {
	return &AccrualHTTPClient{
		accrualSystemAddress: addr,
		accrualGetRoot:       url,
		retryInterval:        retryInterval,
		retryLimit:           retryLimit,
		client:               &http.Client{},
	}
}

func (c *AccrualHTTPClient) SendToAccrual(ctx context.Context, orderid int64) (*proto.AccrualReply, error) {
	defer func() {
		if r := recover(); r != nil {
			mylog.SugarLogger.Errorf("unexpected error caused panic, %v", r)
		}
	}()
	r, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.accrualSystemAddress+c.accrualGetRoot+fmt.Sprint(orderid),
		nil,
	)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot initiate request, %v ", err)

		return &proto.AccrualReply{}, err
	}

	rand.Seed(time.Now().UnixNano())
	retryInterval := c.retryInterval
	for i := 0; i < c.retryLimit; i++ {
		resp, err := c.client.Do(r)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot send request to accrual, %v", err)
			time.Sleep(retryInterval)

			continue
		}
		mylog.SugarLogger.Infof("response header from accrual received. Header: %v, Status: %d", resp.Header, resp.StatusCode)
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			retryInterval = c.SetRetryInterval(resp.Header.Get("Retry-After"))
		case http.StatusOK:
			body, err := io.ReadAll(resp.Body)
			mylog.SugarLogger.Infof("response body from accrual received:, %s", string(body))
			if err != nil {
				mylog.SugarLogger.Errorf("Cannot io.ReadAll resp.Body, %v", err)
				err1 := err
				err := resp.Body.Close()
				if err != nil {
					mylog.SugarLogger.Errorf("cannot close request body, %v", err)

					return &proto.AccrualReply{}, err
				}

				return &proto.AccrualReply{}, err1
			}
			var accrualReply proto.AccrualReply
			err = json.Unmarshal(body, &accrualReply)
			if err != nil {
				mylog.SugarLogger.Errorf("cannot unmarshal body from accrual system, %v", err)

				return &proto.AccrualReply{}, err
			}
			mylog.SugarLogger.Infof("accrual system returned order, %v", proto.AccrualReply{
				Order:   accrualReply.Order,
				Status:  accrualReply.Status,
				Accrual: accrualReply.Accrual,
			},
			)

			return &accrualReply, nil
		}

	}
	return &proto.AccrualReply{}, domain.ErrSenderCannotSendRequestToAccrual
}

func (c *AccrualHTTPClient) SetRetryInterval(accrualRetryInterval string) time.Duration {
	retNum, err := strconv.ParseInt(accrualRetryInterval, 10, 64)
	if err != nil {
		mylog.SugarLogger.Infof("cannot convert retry interval to int64, %v", err)

		return c.retryInterval
	}

	return time.Duration(retNum) * time.Second
}
