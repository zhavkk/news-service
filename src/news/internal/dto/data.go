package dto

import (
	"time"
)

type CreateNewsRequest struct {
	Title     string               `json:"title" validate:"required,min=3,max=255"`
	Category  string               `json:"category" validate:"required,min=2,max=100"`
	Content   []CreateContentBlock `json:"content" validate:"required,dive"`
	StartTime time.Time            `json:"start_time" validate:"required"`
	EndTime   time.Time            `json:"end_time" validate:"required,gtfield=StartTime"`
}

type CreateContentBlock struct {
	Type     string `json:"type" validate:"required,oneof=text link"`
	Content  string `json:"content" validate:"required"`
	Position int    `json:"position" validate:"required,min=0"`
}

type UpdateNewsRequest struct {
	ID        string               `json:"id" validate:"required"`
	Title     string               `json:"title" validate:"omitempty,min=3,max=255"`
	Category  string               `json:"category" validate:"omitempty,min=2,max=100"`
	Content   []CreateContentBlock `json:"content" validate:"omitempty,dive"`
	StartTime *time.Time           `json:"start_time" validate:"omitempty"`
	EndTime   *time.Time           `json:"end_time" validate:"omitempty,gtfield=StartTime"`
}

type NewsListRequest struct {
	Page            int    `query:"page" validate:"min=1" default:"1"`
	Limit           int    `query:"limit" validate:"min=1,max=100" default:"10"`
	Search          string `query:"search"`
	Category        string `query:"category"`
	SortBy          string `query:"sort_by" default:"created_at" validate:"oneof=created_at start_time end_time title category"`
	SortDir         string `query:"sort_dir" default:"desc" validate:"oneof=asc desc"`
	CheckVisibility bool   `query:"check_visibility" default:"true"`
}

type NewsListResponse struct {
	Items      []NewsResponse `json:"items"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

type NewsResponse struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Category  string                 `json:"category"`
	Content   []ContentBlockResponse `json:"content"`
	CreatedAt time.Time              `json:"created_at"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
}

type ContentBlockResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	Position int    `json:"position"`
}

type UpdateNewsResponse struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
	Message   string    `json:"message"`
}

type GetNewsByIDRequest struct {
	ID              string `param:"id" validate:"required"`
	CheckVisibility bool   `query:"check_visibility" default:"true"`
}

type DeleteNewsRequest struct {
	ID string `param:"id" validate:"required"`
}

type DeleteNewsResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
