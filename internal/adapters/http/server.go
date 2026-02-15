package http

import (
	"context"
	"log/slog"
	"strings"
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

func (s *Server) CreateSubscription(
	ctx context.Context,
	request CreateSubscriptionRequestObject,
) (CreateSubscriptionResponseObject, error) {
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

func (s *Server) DeleteSubscription(
	ctx context.Context,
	request DeleteSubscriptionRequestObject,
) (DeleteSubscriptionResponseObject, error) {
	err := s.subscriptions.Delete(ctx, uuid.UUID(request.Id))
	if err != nil {
		// Check if error message contains "not found"
		if strings.Contains(err.Error(), "not found") {
			return DeleteSubscription404JSONResponse{
				Message: "Subscription not found",
			}, nil
		}
		slog.Error("Failed to delete subscription", "error", err, "id", request.Id)
		return DeleteSubscription500JSONResponse{
			Message: "Failed to delete subscription",
		}, nil
	}
	return DeleteSubscription200JSONResponse{
		Message: "Subscription deleted successfully",
	}, nil
}

func (s *Server) GetSubscription(
	ctx context.Context,
	request GetSubscriptionRequestObject,
) (GetSubscriptionResponseObject, error) {
	sub, err := s.subscriptions.ReadByID(ctx, uuid.UUID(request.Id))
	if err != nil {
		return GetSubscription404JSONResponse{
			Message: "Subscription not found",
		}, nil
	}
	return GetSubscription200JSONResponse(toHTTPSubscription(sub)), nil
}

func (s *Server) ReadAllSubscriptions(
	ctx context.Context,
	request ReadAllSubscriptionsRequestObject,
) (ReadAllSubscriptionsResponseObject, error) {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic in ReadAllSubscriptions", "panic", r)
		}
	}()

	// Set defaults for optional parameters
	limit := 50
	offset := 0

	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	if request.Params.Offset != nil {
		offset = *request.Params.Offset
	}

	slog.Info("ReadAllSubscriptions called",
		"user_id", request.Params.UserId,
		"limit", limit,
		"offset", offset)

	list, err := s.subscriptions.ReadAllByUserID(
		ctx,
		uuid.UUID(request.Params.UserId),
		limit,
		offset,
	)
	if err != nil {
		slog.Error("Failed to read subscriptions", "error", err)
		return ReadAllSubscriptions500JSONResponse{
			Message: "Failed to retrieve subscriptions",
		}, nil
	}

	resp := make([]Subscription, 0, len(list))
	for _, sub := range list {
		resp = append(resp, toHTTPSubscription(sub))
	}

	slog.Info("ReadAllSubscriptions success", "count", len(resp))

	return ReadAllSubscriptions200JSONResponse(resp), nil
}

func (s *Server) CalculateTotalCost(
	ctx context.Context,
	request CalculateTotalCostRequestObject,
) (CalculateTotalCostResponseObject, error) {
	slog.Info("CalculateTotalCost called",
		"user_id", request.Params.UserId,
		"service_name", *request.Params.ServiceName,
		"start_date", request.Params.StartDate,
		"end_date", request.Params.EndDate)

	start, err := time.Parse("01-2006", request.Params.StartDate)
	if err != nil {
		slog.Error(
			"Invalid start_date format",
			"error",
			err,
			"start_date",
			request.Params.StartDate,
		)
		return CalculateTotalCost400JSONResponse{
			Message: "Invalid start_date format. Expected MM-YYYY (e.g., 07-2025)",
		}, nil
	}

	var end *time.Time
	if request.Params.EndDate != nil {
		t, err := time.Parse("01-2006", *request.Params.EndDate)
		if err != nil {
			slog.Error("Invalid end_date format", "error", err, "end_date", *request.Params.EndDate)
			return CalculateTotalCost400JSONResponse{
				Message: "Invalid end_date format. Expected MM-YYYY (e.g., 12-2025)",
			}, nil
		}
		end = &t
	}

	slog.Info("Parsed dates", "start", start, "end", end)

	totalCost, err := s.subscriptions.TotalSubscriptionsCost(
		ctx,
		request.Params.UserId,
		*request.Params.ServiceName,
		start,
		end,
	)
	if err != nil {
		slog.Error("Failed to calculate total cost", "error", err)
		return CalculateTotalCost500JSONResponse{
			Message: "Failed to calculate total cost",
		}, nil
	}

	slog.Info("CalculateTotalCost result", "total_cost", totalCost)

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
