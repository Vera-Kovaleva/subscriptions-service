package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Vera-Kovaleva/subscriptions-service/internal/domain"
	"github.com/Vera-Kovaleva/subscriptions-service/internal/infra/pointer"
	"github.com/Vera-Kovaleva/subscriptions-service/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionIntegration(t *testing.T) {
	rollback(t, func(ctx context.Context, connection domain.Connection) {
		repoSubscription := repository.NewSubscription()

		subsID1 := uuid.New()
		subsID2 := uuid.New()
		subsID3 := uuid.New()
		subsID4 := uuid.New()

		userID1 := uuid.New()
		userID2 := uuid.New()

		serviseName1 := domain.ServiceName("servise name 1")
		serviseName2 := domain.ServiceName("servise name 2")

		subscription1User1 := fixtureCreateSubscription(t, connection, subsID1, userID1, serviseName1)
		_ = fixtureCreateSubscription(t, connection, subsID2, userID1, serviseName2)

		subscription1User2 := fixtureCreateSubscription(t, connection, subsID3, userID2, serviseName1)

		subscriptionsFromDBUser1, err := repoSubscription.ReadAll(ctx, connection, userID1)
		require.NoError(t, err)

		subscriptionsFromDBUser2, err := repoSubscription.ReadAll(ctx, connection, userID2)

		require.NoError(t, err)
		require.Len(t, subscriptionsFromDBUser1, 2)
		require.Len(t, subscriptionsFromDBUser2, 1)

		require.Equal(t, subscription1User1, subscriptionsFromDBUser1[0])

		err = repoSubscription.Delete(ctx, connection, subsID1)
		require.NoError(t, err)

		subscriptionsFromDBUser1, err = repoSubscription.ReadAll(ctx, connection, userID1)
		require.NoError(t, err)
		require.Len(t, subscriptionsFromDBUser1, 1)

		now := time.Now()
		newEndDate := now.AddDate(0, 1, 0).UTC().Truncate(24 * time.Hour)
		subscription1User2.EndDate = pointer.Ref(newEndDate)
		err = repoSubscription.Update(ctx, connection, subscription1User2)
		require.NoError(t, err)

		subscriptionsFromDBUser2, err = repoSubscription.ReadAll(
			ctx,
			connection,
			subscription1User2.UserID,
		)
		require.NoError(t, err)
		require.Equal(t, pointer.Ref(newEndDate), subscriptionsFromDBUser2[0].EndDate)

		subscription2User2 := fixtureCreateSubscription(t, connection, subsID4, userID2, serviseName2)
		subscriptionsCosts, err := repoSubscription.AllMatchingSubscriptionsForPeriod(
			ctx,
			connection,
			userID2,
			serviseName2,
			now,
			pointer.Ref(newEndDate),
		)
		require.NoError(t, err)
		require.Equal(t, subscription2User2.Cost, subscriptionsCosts[0])
	})
}

func fixtureCreateSubscription(
	t *testing.T,
	connection domain.Connection,
	subscriptionID domain.SubscriptionID,
	userID domain.UserID,
	name domain.ServiceName,
) domain.Subscription {
	subscription := domain.Subscription{
		ID:        subscriptionID,
		UserID:    userID,
		Cost:      1,
		Name:      name,
		StartDate: time.Now().UTC().Truncate(24 * time.Hour),
	}
	require.NoError(t, repository.NewSubscription().Create(t.Context(), connection, subscription))

	return subscription
}
