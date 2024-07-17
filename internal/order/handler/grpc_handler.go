package handler

import (
	"context"
	"fmt"

	orderpb "github.com/erazr/ecommerce-microservices/internal/common/pb/order"
	"github.com/erazr/ecommerce-microservices/internal/order/domain/order"
)

type OrderGRPCHandler struct {
	repo order.Repository

	orderpb.UnimplementedOrdersServer
}

func NewOrderGRPCHandler(repo order.Repository) *OrderGRPCHandler {
	return &OrderGRPCHandler{repo: repo}
}

func (h *OrderGRPCHandler) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	o, err := h.repo.Get(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	o.Status = req.Status
	fmt.Println("Updating order status", req.Status, o.Status)

	if err := h.repo.Update(ctx, req.OrderId, o); err != nil {
		return nil, err
	}

	return &orderpb.UpdateOrderStatusResponse{
		Success: true,
	}, nil
}

func (h *OrderGRPCHandler) GetOrderProductIDs(ctx context.Context, req *orderpb.GetOrderProductIDsRequest) (*orderpb.GetOrderProductIDsResponse, error) {
	o, err := h.repo.Get(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &orderpb.GetOrderProductIDsResponse{ProductIds: o.ProductID}, nil
}
