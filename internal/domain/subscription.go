package domain

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Vera-Kovaleva/subscriptions-service/internal/infra/log"
)

var _ SubscriptionInterface = (*SubscriptionService)(nil)

var (
	errServiceSubscription         = errors.New("service error")
	ErrServiceReadByIDSubscription = errors.Join(
		errServiceSubscription,
		errors.New("read by id failed"),
	)
	ErrGetLatestSubscriptionEndDate = errors.Join(
		errServiceSubscription,
		errors.New("get end date failed"),
	)
	ErrServiceCreateSubscription = errors.Join(
		errServiceSubscription,
		errors.New("create failed"),
	)
	ErrServiceDeleteSubscription = errors.Join(
		errServiceSubscription,
		errors.New("delete failed"),
	)
	ErrServiceUpdateSubscription = errors.Join(
		errServiceSubscription,
		errors.New("update failed"),
	)
	ErrServiceReadAllByUserID = errors.Join(
		errServiceSubscription,
		errors.New("read all by user id failed"),
	)
	ErrServiceTotalSubscriptionsCost = errors.Join(
		errServiceSubscription,
		errors.New("total cost failed"),
	)
)

type SubscriptionService struct {
	provider         ConnectionProvider
	subscriptionRepo SubscriptionsRepository
}

func NewSubscriptionService(
	provider ConnectionProvider,
	subscriptionRepo SubscriptionsRepository,
) *SubscriptionService {
	return &SubscriptionService{
		provider:         provider,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, subscription Subscription) error {
	slog.DebugContext(ctx, "Service: creating subscription.", log.RequestID(ctx))
	err := s.provider.ExecuteTx(ctx, func(ctx context.Context, c Connection) error {
		latestEndDate, err := s.subscriptionRepo.GetLatestSubscriptionEndDate(
			ctx,
			c,
			subscription.UserID,
			subscription.Name,
		)
		if err != nil {
			return errors.Join(ErrGetLatestSubscriptionEndDate, err)
		}
		if latestEndDate != nil && (latestEndDate.After(subscription.StartDate)) {
			return errors.Join(
				ErrServiceCreateSubscription,
				errors.New("previous subscription has not ended"))
		}

		return s.subscriptionRepo.Create(ctx, c, subscription)
	})
	if err != nil {
		return errors.Join(ErrServiceCreateSubscription, err)
	}
	return nil
}

func (s *SubscriptionService) Delete(
	ctx context.Context,
	subscriptionID SubscriptionID,
) error {
	slog.DebugContext(ctx, "Service: deleting subscription.", log.RequestID(ctx))
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		return s.subscriptionRepo.Delete(ctx, c, subscriptionID)
	})
	if err != nil {
		return errors.Join(ErrServiceDeleteSubscription, err)
	}
	return nil
}

func (s *SubscriptionService) Update(ctx context.Context, subscription Subscription) error {
	slog.DebugContext(ctx, "Service: updating subscription.", log.RequestID(ctx))
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		return s.subscriptionRepo.Update(ctx, c, subscription)
	})
	if err != nil {
		return errors.Join(ErrServiceUpdateSubscription, err)
	}
	return nil
}

func (s *SubscriptionService) ReadByID(ctx context.Context,
	subscriptionID SubscriptionID,
) (Subscription, error) {
	slog.DebugContext(ctx, "Service: getting subscription.", log.RequestID(ctx))
	var subscription Subscription
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		var dbError error
		subscription, dbError = s.subscriptionRepo.Read(ctx, c, subscriptionID)
		return dbError
	})
	if err != nil {
		return subscription, errors.Join(ErrServiceReadByIDSubscription, err)
	}
	return subscription, nil
}

func (s *SubscriptionService) ReadAllByUserID(
	ctx context.Context,
	subscriptionUserID UserID,
	limit int, offset int,
) ([]Subscription, error) {
	slog.DebugContext(ctx, "Service: reading subscription by user ID.", log.RequestID(ctx))
	var subscriptions []Subscription
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		var dbErr error
		subscriptions, dbErr = s.subscriptionRepo.ReadAll(ctx, c, subscriptionUserID, limit, offset)
		return dbErr
	})
	if err != nil {
		return subscriptions, errors.Join(ErrServiceReadAllByUserID, err)
	}
	return subscriptions, nil
}

func (s *SubscriptionService) TotalSubscriptionsCost(
	ctx context.Context,
	subscriptionUserID UserID,
	subscriptionName ServiceName,
	start time.Time,
	end *time.Time,
) (int, error) {
	var totalCost int
	err := s.provider.Execute(ctx, func(ctx context.Context, c Connection) error {
		slog.DebugContext(ctx, "Service: calculating total cost.", log.RequestID(ctx))
		var dbErr error
		totalCost, dbErr = s.subscriptionRepo.CalculateTotalCost(
			ctx,
			c,
			subscriptionUserID,
			subscriptionName,
			start,
			end,
		)
		return dbErr
	})
	if err != nil {
		return totalCost, errors.Join(ErrServiceTotalSubscriptionsCost, err)
	}
	return totalCost, nil
}
