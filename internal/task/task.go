package task

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListFilter struct {
	Done     *bool
	Page     int
	PageSize int
}

func (f ListFilter) limit() int  { return f.PageSize }
func (f ListFilter) offset() int { return (f.Page - 1) * f.PageSize }

type Metadata struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
