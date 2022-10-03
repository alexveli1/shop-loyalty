package service

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/alexveli/diploma/internal/domain"
	"github.com/alexveli/diploma/internal/proto"
	"github.com/alexveli/diploma/internal/repository"
	"github.com/alexveli/diploma/internal/transport/httpv1/client"
	mylog "github.com/alexveli/diploma/pkg/log"
)

type AccrualService struct {
	sendInterval time.Duration
	reqCh        chan string
	repo         repository.Accrualer
	client       client.HTTPClient
}

func NewAccrualer(interval time.Duration, repo repository.Accrualer, client client.HTTPClient) *AccrualService {
	return &AccrualService{
		sendInterval: interval,
		reqCh:        make(chan string),
		repo:         repo,
		client:       client,
	}
}

func (h *AccrualService) CheckRequestFormat(content string) (int64, bool) {
	num, err := strconv.ParseInt(content, 10, 64)
	if err != nil {
		mylog.SugarLogger.Errorf("cannot convert content to int64, %v", err)

		return 0, false
	}
	return num, true
}

func (h *AccrualService) CheckOrderAlreadyUploaded(ctx context.Context, orderid int64) (int64, bool, error) {
	return h.repo.CheckOrderAlreadyUploaded(ctx, orderid)
}
func (h *AccrualService) AddOrderToQueue(ctx context.Context, order *proto.Order) bool {
	ok := h.repo.InsertOrUpdateOrder(ctx, order)
	if !ok {
		mylog.SugarLogger.Errorf("cannot update order")

		return false
	}
	return true
}

func (h *AccrualService) GetFirstUnprocessedOrder(ctx context.Context) (*proto.Order, bool) {
	return h.repo.GetFirstUnprocessedOrder(ctx)
}

func (h *AccrualService) UpdateOrderAndBalance(ctx context.Context, accrualReply *proto.Order) {
	h.repo.IncreaseOrderAccrualAndBalanceCurrent(ctx, accrualReply)
}

func (h *AccrualService) SendToAccrual(ctx context.Context) {
	tick := time.NewTicker(h.sendInterval)
	for range tick.C {
		order, ok := h.GetFirstUnprocessedOrder(ctx)
		if ok {
			processedOrder, err := h.client.SendToAccrual(ctx, order.Orderid)
			if err == nil {
				h.UpdateOrderAndBalance(ctx, &proto.Order{
					Orderid:              order.Orderid,
					Userid:               order.Userid,
					Status:               processedOrder.Status,
					Accrualsum:           processedOrder.Accrual,
					ProcessedByAccrualAt: timestamppb.New(time.Now()),
				})
			} else {
				h.UpdateOrderAndBalance(ctx, &proto.Order{
					Orderid:              order.Orderid,
					Userid:               order.Userid,
					Status:               domain.NEW,
					ProcessedByAccrualAt: timestamppb.New(time.Now()),
				})
			}
		}
	}
}
