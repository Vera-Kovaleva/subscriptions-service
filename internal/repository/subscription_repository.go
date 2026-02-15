package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Vera-Kovaleva/subscriptions-service/internal/domain"
)

var (
	errSubscription         = errors.New("subscription repository error")
	ErrCreateSubscription   = errors.Join(errSubscription, errors.New("create failed"))
	ErrReadSubscription     = errors.Join(errSubscription, errors.New("read failed"))
	ErrReadAllSubscriptions = errors.Join(
		errSubscription,
		errors.New("read all failed"),
	)
	ErrDeleteSubscription                = errors.Join(errSubscription, errors.New("delete failed"))
	ErrUpdateSubscription                = errors.Join(errSubscription, errors.New("update failed"))
	ErrAllMatchingSubscriptionsForPeriod = errors.Join(
		errSubscription,
		errors.New("all matching subscriptions failed"),
	)
	ErrGetLatestDateSubscription = errors.Join(
		errSubscription,
		errors.New("get latest date failed"),
	)
)

var _ domain.SubscriptionsRepository = (*SubscriptionRepository)(nil)

type SubscriptionRepository struct{}

func NewSubscription() *SubscriptionRepository {
	return &SubscriptionRepository{}
}

func (s *SubscriptionRepository) Create(
	ctx context.Context,
	connection domain.Connection,
	subscription domain.Subscription,
) error {
	const query = `insert into subscriptions
	(id, service_name, month_cost, user_id, subs_start_date, subs_end_date)
	values
	($1, $2, $3, $4, $5, $6)`

	if _, err := connection.ExecContext(ctx, query, subscription.ID, subscription.Name, subscription.Cost, subscription.UserID, subscription.StartDate, subscription.EndDate); err != nil {
		return errors.Join(ErrCreateSubscription, err)
	}

	return nil
}

func (s *SubscriptionRepository) Delete(
	ctx context.Context,
	connection domain.Connection,
	subscriptionID domain.SubscriptionID,
) error {
	const query = `delete from subscriptions where id = $1`
	rowsAffected, err := connection.ExecContext(ctx, query, subscriptionID)
	if err != nil {
		return errors.Join(ErrDeleteSubscription, err)
	}
	if rowsAffected == 0 {
		return errors.Join(ErrDeleteSubscription, errors.New("subscription not found"))
	}
	return nil
}

func (s *SubscriptionRepository) Read(ctx context.Context,
	connection domain.Connection,
	subscriptionID domain.SubscriptionID,
) (domain.Subscription, error) {
	var subscription domain.Subscription
	const query = `select id, service_name, month_cost, user_id, subs_start_date, subs_end_date from subscriptions
	where id = $1`

	if err := connection.GetContext(ctx, &subscription, query, subscriptionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription, errors.Join(
				ErrReadSubscription,
				errors.New("subscription not found"),
			)
		}
		return subscription, errors.Join(ErrReadSubscription, err)
	}
	return subscription, nil
}

func (s *SubscriptionRepository) ReadAll(
	ctx context.Context,
	connection domain.Connection,
	userID domain.UserID,
	limit int,
	offset int,
) ([]domain.Subscription, error) {
	const query = `select id, service_name, month_cost, user_id, subs_start_date, subs_end_date from subscriptions where user_id=$1 order by subs_start_date desc limit $2 offset $3`
	var allUserSubscriptions []domain.Subscription
	if err := connection.SelectContext(ctx, &allUserSubscriptions, query, userID, limit, offset); err != nil {
		return allUserSubscriptions, errors.Join(ErrReadAllSubscriptions, err)
	}
	return allUserSubscriptions, nil
}

func (s *SubscriptionRepository) Update(ctx context.Context,
	connection domain.Connection,
	subscription domain.Subscription,
) error {
	const query = `update subscriptions set service_name = $2 , user_id = $3, month_cost = $4, subs_end_date=$5
	where id = $1`

	if _, err := connection.ExecContext(ctx, query, subscription.ID, subscription.Name, subscription.UserID, subscription.Cost, subscription.EndDate); err != nil {
		return errors.Join(ErrUpdateSubscription, err)
	}

	return nil
}

func (s *SubscriptionRepository) CalculateTotalCost(ctx context.Context,
	connection domain.Connection,
	subscriptionUserID domain.UserID,
	subscriptionName domain.ServiceName,
	start time.Time,
	end *time.Time,
) (int, error) {
	const query = `select 
    COALESCE(sum(
        month_cost * (
            -- Calculate number of months between start and end
            (extract(year from least(COALESCE(subs_end_date, $4), $4))::int - 
             extract(year from greatest(subs_start_date, $3))::int) * 12 +
            (extract(month from least(COALESCE(subs_end_date, $4), $4))::int - 
             extract(month from greatest(subs_start_date, $3))::int) + 1
        )
    ), 0) as total_cost
from subscriptions 
where user_id = $1 
  and ($2 = '' OR service_name = $2)
  and subs_start_date <= $4 
  and (subs_end_date IS NULL OR subs_end_date >= $3)`
	var totalCost int
	if err := connection.GetContext(ctx, &totalCost, query, subscriptionUserID, subscriptionName, start, end); err != nil {
		return totalCost, errors.Join(ErrAllMatchingSubscriptionsForPeriod, err)
	}
	return totalCost, nil
}

func (s *SubscriptionRepository) GetLatestSubscriptionEndDate(ctx context.Context,
	connection domain.Connection,
	userID domain.UserID,
	serviceName domain.ServiceName,
) (*time.Time, error) {
	const query = `select (subs_end_date) from subscriptions
	where user_id = $1 and service_name = $2 order by subs_start_date desc limit 1`
	var latestDate *time.Time
	if err := connection.GetContext(ctx, &latestDate, query, userID, serviceName); err != nil &&
		!errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.Join(ErrGetLatestDateSubscription, err)
	}
	return latestDate, nil
}
