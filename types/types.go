package types

import (
	"encoding/json"
	"time"
)

type ResponseUrl struct {
	ID        int    `json:"id,omitempty"`
	ShortCode string `json:"shortCode"`
	LongUrl   string `json:"longUrl"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type Task struct {
	TaskID    string          `json:"task_id"`
	Status    string          `json:"status"`
	Result    json.RawMessage `json:"result,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}
