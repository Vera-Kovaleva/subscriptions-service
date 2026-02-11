package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type (
	SubscriptionID = uuid.UUID
	UserID         = uuid.UUID
	ServiceName    = string

	Subscription struct {
		ID        SubscriptionID `db:"id"`
		Name      ServiceName    `db:"service_name"`
		Cost      int            `db:"month_cost"`
		UserID    UserID         `db:"user_id"`
		StartDate time.Time      `db:"subs_start_date"`
		EndDate   *time.Time     `db:"subs_end_date"`
	}

	Connection interface {
		GetContext(context.Context, any, string, ...any) error
		SelectContext(context.Context, any, string, ...any) error
		ExecContext(context.Context, string, ...any) (int64, error)
	}
	ConnectionProvider interface {
		Execute(context.Context, func(context.Context, Connection) error) error
		ExecuteTx(context.Context, func(context.Context, Connection) error) error
		io.Closer
	}

	SubscriptionInterface interface {
		Create(context.Context, Subscription) error
		ReadByID(context.Context, SubscriptionID) (Subscription, error)
		Update(context.Context, Subscription) error
		Delete(context.Context, SubscriptionID) error
		ReadAllByUserID(context.Context, UserID) ([]Subscription, error)
		TotalSubscriptionsCost(
			context.Context,
			UserID,
			*ServiceName,
			time.Time,
			*time.Time,
		) (int, error)
	}
)
