package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshRepository struct {
	db *pgxpool.Pool
}

func NewRefreshRepository(db *pgxpool.Pool) *RefreshRepository {
	return &RefreshRepository{db: db}
}

func (r *RefreshRepository) Create(ctx context.Context, userID int64, hash string, expiresAt time.Time) error {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`

	if _, err := r.db.Exec(ctx, q, userID, hash, expiresAt); err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *RefreshRepository) Rotate(ctx context.Context, oldHash, newHash string, expiresAt time.Time) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const deleteQ = `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > now()
		RETURNING user_id`

	var userID int64
	err = tx.QueryRow(ctx, deleteQ, oldHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrInvalidToken
	}
	if err != nil {
		return 0, fmt.Errorf("delete refresh token: %w", err)
	}

	const insertQ = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`

	if _, err := tx.Exec(ctx, insertQ, userID, newHash, expiresAt); err != nil {
		return 0, fmt.Errorf("insert rotated token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return userID, nil
}

func (r *RefreshRepository) Delete(ctx context.Context, hash string) error {
	const q = `DELETE FROM refresh_tokens WHERE token_hash = $1`

	if _, err := r.db.Exec(ctx, q, hash); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (r *RefreshRepository) DeleteAllForUser(ctx context.Context, userID int64) error {
	const q = `DELETE FROM refresh_tokens WHERE user_id = $1`

	if _, err := r.db.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("delete user refresh tokens: %w", err)
	}
	return nil
}
