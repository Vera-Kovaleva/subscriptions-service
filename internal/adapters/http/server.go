package http

import (
	"context"
	"time"

	"github.com/Vera-Kovaleva/subscriptions-service/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	subscriptions domain.SubscriptionInterface
}

func NewServer(
	subscriptions domain.SubscriptionInterface,
) *Server {
	return &Server{
		subscriptions: subscriptions,
	}
}

func (s *Server) CreateSubscription(ctx context.Context, request CreateSubscriptionRequestObject) (CreateSubscriptionResponseObject, error) {
	subscription, err := toDomainSubscription(request.Body)
	if err != nil {
		return CreateSubscription400JSONResponse{
			Message: "Invalid request data: " + err.Error(),
		}, nil
	}

	subscription.ID = uuid.New()

	err = s.subscriptions.Create(ctx, subscription)
	if err != nil {
		return CreateSubscription500JSONResponse{
			Message: "Failed to create subscription",
		}, nil
	}

	return CreateSubscription201JSONResponse(toHTTPSubscription(subscription)), nil
}

func (s *Server) DeleteSubscription(ctx context.Context, request DeleteSubscriptionRequestObject) (DeleteSubscriptionResponseObject, error) {
	err := s.subscriptions.Delete(ctx, uuid.UUID(request.Id))
	if err != nil {
		return DeleteSubscription500JSONResponse{
			Message: "Failed to delete subscription",
		}, nil
	}
	return DeleteSubscription200JSONResponse{
		Message: "Subscription deleted successfully",
	}, nil
}

func (s *Server) GetSubscription(ctx context.Context, request GetSubscriptionRequestObject) (GetSubscriptionResponseObject, error) {
	sub, err := s.subscriptions.ReadByID(ctx, uuid.UUID(request.Id))
	if err != nil {
		return GetSubscription404JSONResponse{
			Message: "Subscription not found",
		}, nil
	}
	return GetSubscription200JSONResponse(toHTTPSubscription(sub)), nil
}

func (s *Server) ReadAllSubscriptions(ctx context.Context, request ReadAllSubscriptionsRequestObject) (ReadAllSubscriptionsResponseObject, error) {
	list, err := s.subscriptions.ReadAllByUserID(
		ctx,
		uuid.UUID(request.Params.UserId),
	)
	if err != nil {
		return ReadAllSubscriptions500JSONResponse{
			Message: "Failed to retrieve subscriptions",
		}, nil
	}

	resp := make([]Subscription, 0, len(list))
	for _, sub := range list {
		resp = append(resp, toHTTPSubscription(sub))
	}

	return ReadAllSubscriptions200JSONResponse(resp), nil
}

func (s *Server) CalculateTotalCost(ctx context.Context, request CalculateTotalCostRequestObject) (CalculateTotalCostResponseObject, error) {
	// Parse MM-YYYY format
	start, err := time.Parse("01-2006", request.Params.StartDate)
	if err != nil {
		return CalculateTotalCost400JSONResponse{
			Message: "Invalid start_date format. Expected MM-YYYY (e.g., 07-2025)",
		}, nil
	}

	var end *time.Time
	if request.Params.EndDate != nil {
		t, err := time.Parse("01-2006", *request.Params.EndDate)
		if err != nil {
			return CalculateTotalCost400JSONResponse{
				Message: "Invalid end_date format. Expected MM-YYYY (e.g., 12-2025)",
			}, nil
		}
		end = &t
	}

	totalCost, err := s.subscriptions.TotalSubscriptionsCost(
		ctx,
		request.Params.UserId,
		*request.Params.ServiceName,
		start,
		end,
	)

	if err != nil {
		return CalculateTotalCost500JSONResponse{
			Message: "Failed to calculate total cost",
		}, nil
	}

	return CalculateTotalCost200JSONResponse{
		TotalCost: totalCost,
	}, nil
}

func toDomainSubscription(req *CreateSubscriptionJSONRequestBody) (domain.Subscription, error) {
	start, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	var end *time.Time
	if req.EndDate != nil {
		t, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			return domain.Subscription{}, err
		}
		end = &t
	}

	return domain.Subscription{
		Name:      req.ServiceName,
		Cost:      req.Price,
		UserID:    uuid.UUID(req.UserId),
		StartDate: start,
		EndDate:   end,
	}, nil
}

func toHTTPSubscription(s domain.Subscription) Subscription {
	start := s.StartDate.Format("01-2006")

	var end *string
	if s.EndDate != nil {
		str := s.EndDate.Format("01-2006")
		end = &str
	}

	return Subscription{
		Id:          openapi_types.UUID(s.ID),
		ServiceName: s.Name,
		Price:       s.Cost,
		UserId:      openapi_types.UUID(s.UserID),
		StartDate:   start,
		EndDate:     end,
	}
}
