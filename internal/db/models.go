// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type PointsEvent struct {
	ID         uuid.UUID             `json:"id"`
	UserID     uuid.UUID             `json:"user_id"`
	EventType  string                `json:"event_type"`
	RefID      sql.NullString        `json:"ref_id"`
	Points     int32                 `json:"points"`
	Meta       pqtype.NullRawMessage `json:"meta"`
	OccurredAt time.Time             `json:"occurred_at"`
}

type Redemption struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	RewardID    uuid.UUID `json:"reward_id"`
	PointsSpent int32     `json:"points_spent"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RewardsCatalog struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Description sql.NullString        `json:"description"`
	Cost        int32                 `json:"cost"`
	Segment     pqtype.NullRawMessage `json:"segment"`
	Active      bool                  `json:"active"`
	CreatedBy   uuid.NullUUID         `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
}

type Rule struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description sql.NullString  `json:"description"`
	Config      json.RawMessage `json:"config"`
	Active      bool            `json:"active"`
	CreatedBy   uuid.NullUUID   `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Segment struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description sql.NullString  `json:"description"`
	Criteria    json.RawMessage `json:"criteria"`
	Active      bool            `json:"active"`
	CreatedBy   uuid.NullUUID   `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}
