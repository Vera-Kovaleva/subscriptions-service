package domain

import (
	"context"
	"time"
)

type SubscriptionsRepository interface {
	Create(context.Context, Connection, Subscription) error
	Update(context.Context, Connection, Subscription) error
	Delete(context.Context, Connection, SubscriptionID) error
	ReadAll(context.Context, Connection, UserID) ([]Subscription, error)
	Read(context.Context, Connection, SubscriptionID) (Subscription, error)
	AllMatchingSubscriptionsForPeriod(
		context.Context,
		Connection,
		UserID,
		ServiceName,
		time.Time,
		*time.Time,
	) ([]int, error)
	GetLatestSubscriptionEndDate(context.Context, Connection, UserID, ServiceName) (*time.Time, error)
}
