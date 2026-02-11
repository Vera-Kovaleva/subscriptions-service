package database

import (
	"context"

	"github.com/Vera-Kovaleva/subscriptions-service/internal/domain"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ domain.ConnectionProvider = (*PostgresProvider)(nil)
	_ domain.Connection         = (*PostgresConnection)(nil)
	_ domain.Connection         = (*PostgresTransaction)(nil)
)

type (
	PostgresProvider struct {
		pool        *pgxpool.Pool
		connFactory func(*pgxpool.Conn) domain.Connection
		txFactory   func(pgx.Tx) domain.Connection
	}

	PostgresProviderOption func(*PostgresProvider)

	PostgresConnection struct {
		connection *pgxpool.Conn
	}

	PostgresTransaction struct {
		transaction pgx.Tx
	}
)

func NewPostgresProvider(pool *pgxpool.Pool, options ...PostgresProviderOption) *PostgresProvider {
	provider := &PostgresProvider{
		pool:        pool,
		connFactory: func(conn *pgxpool.Conn) domain.Connection { return NewPostgresConnection(conn) },
		txFactory:   func(tx pgx.Tx) domain.Connection { return NewPostgresTransaction(tx) },
	}

	for _, o := range options {
		o(provider)
	}

	return provider
}

func WithConnectionFactory(factory func(*pgxpool.Conn) domain.Connection) PostgresProviderOption {
	return func(p *PostgresProvider) {
		p.connFactory = factory
	}
}

func WithTransactionFactory(factory func(pgx.Tx) domain.Connection) PostgresProviderOption {
	return func(p *PostgresProvider) {
		p.txFactory = factory
	}
}

func (p *PostgresProvider) Execute(
	ctx context.Context,
	receiver func(context.Context, domain.Connection) error,
) error {
	return p.acquire(ctx, func(ctx context.Context, c *pgxpool.Conn) error {
		return receiver(ctx, p.connFactory(c))
	})
}

func (p *PostgresProvider) ExecuteTx(
	ctx context.Context,
	receiver func(context.Context, domain.Connection) error,
) error {
	return p.acquire(ctx, func(ctx context.Context, c *pgxpool.Conn) error {
		tx, err := c.Begin(ctx)
		if err != nil {
			return err
		}

		defer func(tx pgx.Tx) {
			if err := recover(); err != nil {
				_ = tx.Rollback(ctx)
			}
		}(tx)

		err = receiver(ctx, p.txFactory(tx))
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}

		return err
	})
}

func (p *PostgresProvider) acquire(
	ctx context.Context,
	f func(context.Context, *pgxpool.Conn) error,
) error {
	ctx = context.WithoutCancel(ctx)
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return f(ctx, conn)
}

func (p *PostgresProvider) Close() error {
	p.pool.Close()
	p.pool = nil
	p.connFactory = nil
	p.txFactory = nil

	return nil
}

func NewPostgresConnection(connection *pgxpool.Conn) *PostgresConnection {
	return &PostgresConnection{connection: connection}
}

func (p *PostgresConnection) ExecContext(
	ctx context.Context,
	query string,
	args ...any,
) (int64, error) {
	cmdTag, err := p.connection.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return cmdTag.RowsAffected(), nil
}

func (p *PostgresConnection) GetContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	return pgxscan.Get(ctx, p.connection, dest, query, args...)
}

func (p *PostgresConnection) SelectContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	return pgxscan.Select(ctx, p.connection, dest, query, args...)
}

func NewPostgresTransaction(transaction pgx.Tx) *PostgresTransaction {
	return &PostgresTransaction{transaction: transaction}
}

func (p *PostgresTransaction) ExecContext(
	ctx context.Context,
	query string,
	args ...any,
) (int64, error) {
	cmdTag, err := p.transaction.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return cmdTag.RowsAffected(), nil
}

func (p *PostgresTransaction) GetContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	return pgxscan.Get(ctx, p.transaction, dest, query, args...)
}

func (p *PostgresTransaction) SelectContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	return pgxscan.Select(ctx, p.transaction, dest, query, args...)
}
