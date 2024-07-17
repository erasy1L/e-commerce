package handler

import (
	"context"
	"fmt"

	productProto "github.com/erazr/ecommerce-microservices/internal/common/pb/product"
	"github.com/erazr/ecommerce-microservices/internal/product/domain/product"
)

type ProductGRPCHandler struct {
	repo product.Repository

	productProto.UnimplementedProductsServer
}

func NewProductGRPCHandler(repo product.Repository) *ProductGRPCHandler {
	return &ProductGRPCHandler{
		repo: repo,
	}
}

func (h *ProductGRPCHandler) UpdateProductStock(ctx context.Context, req *productProto.UpdateProductStockRequest) (*productProto.UpdateProductStockResponse, error) {
	var (
		success bool
		message string
	)

	for _, u := range req.Updates {
		p, err := h.repo.Get(ctx, u.ProductId)
		if err != nil {
			return nil, err
		}

		switch u.UpdateType {
		case productProto.UpdateType_INCREMENT:
			p.Amount += int(u.GetQuantity())

			err = h.repo.Update(ctx, p.ID, p)
			if err != nil {
				return nil, err
			}

			success = true
			message = fmt.Sprintf("stock for product with id %s incremented by %d", u.ProductId, u.GetQuantity())

		case productProto.UpdateType_DECREMENT:
			if p.Amount < int(u.GetQuantity()) {
				return &productProto.UpdateProductStockResponse{
					Success: false,
					Message: fmt.Sprintf("not enough stock for product with id %s", u.ProductId),
				}, nil
			}

			err = h.repo.Update(ctx, p.ID, p)
			if err != nil {
				return nil, err
			}

			p.Amount -= int(u.GetQuantity())
			success = true
			message = fmt.Sprintf("stock for product with id %s decremented by %d", u.ProductId, u.GetQuantity())
		}

	}

	return &productProto.UpdateProductStockResponse{
		Success: success,
		Message: message,
	}, nil
}

func (h *ProductGRPCHandler) ProductsAvailable(ctx context.Context, req *productProto.ProductsAvailableRequest) (*productProto.ProductsAvailableResponse, error) {
	products := make(map[string]int)

	availables := make([]*productProto.ProductAvailability, 0)

	for _, id := range req.ProductIds {
		products[id]++
	}

	for _, id := range req.ProductIds {
		p, err := h.repo.Get(ctx, id)
		if err != nil {
			return nil, err
		}

		availables = append(availables, &productProto.ProductAvailability{
			ProductId: id,
			Available: p.Amount >= products[id],
			Name:      p.Name,
			Stock:     int32(p.Amount),
		})
	}

	return &productProto.ProductsAvailableResponse{
		Availability: availables,
	}, nil
}

func (h *ProductGRPCHandler) GetProductPrices(ctx context.Context, req *productProto.GetProductPricesRequest) (*productProto.GetProductPricesResponse, error) {
	prices := make(map[string]float32)

	for _, id := range req.ProductIds {
		price, err := h.repo.GetPriceByID(ctx, id)
		if err != nil {
			return nil, err
		}

		prices[id] += float32(price)
	}

	return &productProto.GetProductPricesResponse{
		Prices: prices,
	}, nil
}
