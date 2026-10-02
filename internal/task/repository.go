package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const taskColumns = `id, user_id, title, description, done, created_at, updated_at`

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (*Task, error) {
	var t Task
	err := s.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Done, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID int64, title, description string) (*Task, error) {
	const q = `
		INSERT INTO tasks (user_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING ` + taskColumns

	t, err := scanTask(r.db.QueryRow(ctx, q, userID, title, description))
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}
	return t, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, id int64) (*Task, error) {
	const q = `
		SELECT ` + taskColumns + `
		FROM tasks
		WHERE id = $1 AND user_id = $2`

	t, err := scanTask(r.db.QueryRow(ctx, q, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

func (r *Repository) List(ctx context.Context, userID int64, f ListFilter) ([]Task, int, error) {
	const countQ = `
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1 AND ($2::boolean IS NULL OR done = $2)`

	var total int
	if err := r.db.QueryRow(ctx, countQ, userID, f.Done).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	const listQ = `
		SELECT ` + taskColumns + `
		FROM tasks
		WHERE user_id = $1 AND ($2::boolean IS NULL OR done = $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, listQ, userID, f.Done, f.limit(), f.offset())
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, total, nil
}

func (r *Repository) Update(ctx context.Context, userID, id int64, in UpdateInput) (*Task, error) {
	const q = `
		UPDATE tasks SET
			title       = COALESCE($3, title),
			description = COALESCE($4, description),
			done        = COALESCE($5, done),
			updated_at  = now()
		WHERE id = $1 AND user_id = $2
		RETURNING ` + taskColumns

	t, err := scanTask(r.db.QueryRow(ctx, q, id, userID, in.Title, in.Description, in.Done))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	const q = `DELETE FROM tasks WHERE id = $1 AND user_id = $2`

	tag, err := r.db.Exec(ctx, q, id, userID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
