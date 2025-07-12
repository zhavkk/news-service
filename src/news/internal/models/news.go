package models

import "time"

type News struct {
	ID        int64          `json:"id"`
	Title     string         `json:"title"`
	Category  string         `json:"category"`
	Content   []ContentBlock `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
}

type ContentBlock struct {
	ID        int64     `json:"id"`
	NewsID    int64     `json:"news_id"`
	Type      BlockType `json:"type"`
	Content   string    `json:"content"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type BlockType string

const (
	TextBlock BlockType = "text"
	LinkBlock BlockType = "link"
)

func (n *News) IsVisible() bool {
	now := time.Now()
	startOk := n.StartTime.IsZero() || now.After(n.StartTime) || now.Equal(n.StartTime)
	endOk := n.EndTime.IsZero() || now.Before(n.EndTime) || now.Equal(n.EndTime)
	return startOk && endOk
}
