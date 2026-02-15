package database

import (
	"context"

	"github.com/Vera-Kovaleva/subscriptions-service/internal/domain"
)

var _ domain.ConnectionProvider = (*DummyProvider)(nil)

type DummyProvider struct {
	connection domain.Connection
}

func NewDummyProvider(connection domain.Connection) *DummyProvider {
	return &DummyProvider{
		connection: connection,
	}
}

func (p DummyProvider) Execute(
	ctx context.Context,
	receiver func(context.Context, domain.Connection) error,
) error {
	return receiver(ctx, p.connection)
}

func (p DummyProvider) ExecuteTx(
	ctx context.Context,
	receiver func(context.Context, domain.Connection) error,
) error {
	return receiver(ctx, p.connection)
}

func (p DummyProvider) Close() error {
	return nil
}
