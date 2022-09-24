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
	accrualUrl           string
	retryInterval        time.Duration
	retryLimit           int
}

type HTTPClient interface {
	SendToAccrual(ctx context.Context, orderid int64) (*proto.AccrualReply, error)
	SetRetryInterval(accrualRetryInterval string) time.Duration
}

func NewAccrualHTTPClient(addr string, url string, retryInterval time.Duration, retryLimit int) *AccrualHTTPClient {
	return &AccrualHTTPClient{
		accrualSystemAddress: addr,
		accrualUrl:           url,
		retryInterval:        retryInterval,
		retryLimit:           retryLimit,
	}
}

func (c *AccrualHTTPClient) SendToAccrual(ctx context.Context, orderid int64) (*proto.AccrualReply, error) {
	defer func() {
		if r := recover(); r != nil {
			mylog.SugarLogger.Errorf("unexpected error caused panic, %v", r)
		}
	}()
	mylog.SugarLogger.Info("trying to send request to accrual system: http://" + c.accrualSystemAddress + c.accrualUrl + fmt.Sprint(orderid))
	r, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://"+c.accrualSystemAddress+c.accrualUrl+fmt.Sprint(orderid),
		nil,
	)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot initiate request, %v ", err)

		return &proto.AccrualReply{}, err
	}

	client := &http.Client{}
	rand.Seed(time.Now().UnixNano())
	retryInterval := c.retryInterval
	for i := 0; i < c.retryLimit; i++ {
		resp, err := client.Do(r)
		if err != nil {
			mylog.SugarLogger.Errorf("cannot send request to accrual, %v", err)
			time.Sleep(retryInterval)

			continue
		}
		mylog.SugarLogger.Infof("response header from accrual received. Header: %v, Status: %d", resp.Header, resp.StatusCode)
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			retryInterval = c.SetRetryInterval(resp.Header.Get("Retry-After"))

			return &proto.AccrualReply{}, domain.ErrSenderTooManyRequests
		case http.StatusOK:
			body, err := io.ReadAll(resp.Body)
			mylog.SugarLogger.Infof("response body from accrual received:, %s", string(body))
			if err != nil {
				mylog.SugarLogger.Errorf("Cannot io.ReadAll resp.Body, %v", err)

				return &proto.AccrualReply{}, err
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					mylog.SugarLogger.Errorf("Cannot close response body, %v", err)
				}
			}(resp.Body)
			var accrualReply proto.AccrualReply
			err = json.Unmarshal(body, &accrualReply)

			mylog.SugarLogger.Infof("accrual system returned order, %v", accrualReply)

			if err != nil {
				mylog.SugarLogger.Errorf("cannot unmarshal body from accrual system, %v", err)

				return &proto.AccrualReply{}, err
			}
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
