package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"SubscriptionAggregator/pkg/model"
)

type Postgres struct {
	DB *sql.DB
}

func New(ctx context.Context, connString string) (*Postgres, error) {
	const op = "repository.postgresql.New"

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := RunMigrations(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: migrations failed: %w", op, err)
	}

	return &Postgres{DB: db}, nil
}

func (p *Postgres) Close() error {
	return p.DB.Close()
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter model.SubscriptionFilter) ([]*model.Subscription, error)
	GetTotalCost(ctx context.Context, filter model.SubscriptionFilter) (int, error)
}

type postgresSubscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) SubscriptionRepository {
	return &postgresSubscriptionRepo{db: db}
}

func (r *postgresSubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	const op = "repository.postgresql.Create"

	query := `
		INSERT INTO subscriptions 
			(id, service_name, price, user_id, start_date, end_date) 
		VALUES 
			($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *postgresSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	const op = "repository.postgresql.GetByID"

	query := `
		SELECT 
			id, service_name, price, user_id, start_date, end_date 
		FROM 
			subscriptions 
		WHERE 
			id = $1`

	var sub model.Subscription
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &sub, nil
}

func (r *postgresSubscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
	const op = "repository.postgresql.Update"

	query := `
		UPDATE subscriptions 
		SET 
			service_name = $2, 
			price = $3, 
			user_id = $4, 
			start_date = $5, 
			end_date = $6 
		WHERE 
			id = $1`

	result, err := r.db.ExecContext(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to check rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("not found")
	}

	return nil
}

func (r *postgresSubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgresql.Delete"

	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete subscription: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to check rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s, subscription not found", op)
	}

	return nil
}

func (r *postgresSubscriptionRepo) List(ctx context.Context, filter model.SubscriptionFilter) ([]*model.Subscription, error) {
	const op = "repository.postgresql.List"

	query := `
		SELECT 
			id, service_name, price, user_id, start_date, end_date 
		FROM 
			subscriptions 
		WHERE 
			($1::uuid IS NULL OR user_id = $1) AND
			($2::text IS NULL OR service_name = $2) AND
			($3::timestamp IS NULL OR start_date >= $3) AND
			($4::timestamp IS NULL OR (end_date IS NULL OR end_date <= $4))`

	rows, err := r.db.QueryContext(ctx, query,
		filter.UserID,
		filter.ServiceName,
		filter.FromDate,
		filter.ToDate,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		var sub model.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan subscription: %w", op, err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return subscriptions, nil
}

func (r *postgresSubscriptionRepo) GetTotalCost(ctx context.Context, filter model.SubscriptionFilter) (int, error) {
	const op = "repository.postgresql.GetTotalCost"
	//COALESCE to prevent null error
	query := `
		SELECT 
			COALESCE(SUM(price), 0) 
		FROM 
			subscriptions 
		WHERE 
			($1::uuid IS NULL OR user_id = $1) AND
			($2::text IS NULL OR service_name = $2) AND
			($3::timestamp IS NULL OR start_date >= $3) AND
			($4::timestamp IS NULL OR (end_date IS NULL OR end_date <= $4))`

	var total int
	err := r.db.QueryRowContext(ctx, query,
		filter.UserID,
		filter.ServiceName,
		filter.FromDate,
		filter.ToDate,
	).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return total, nil
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	const op = "repository.postgresql.RunMigrations"

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: failed to get working directory: %w", op, err)
	}

	migrationPath := filepath.Join(wd, "migrations", "001_init.sql")

	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("%s, failed to read migration file at %s: %w", op, migrationPath, err)
	}

	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		return fmt.Errorf("%s: failed to execute migration: %w", op, err)
	}

	return nil
}
