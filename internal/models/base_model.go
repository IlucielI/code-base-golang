package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for GORM entities.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// JSONMap is a custom GORM data type for handling arbitrary JSONB map data.
type JSONMap map[string]interface{}

// Value serializes the JSONMap into JSON driver.Value for SQL storage.
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONMap: %w", err)
	}
	return string(bytes), nil
}

// Scan deserializes a database driver value into JSONMap.
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for JSONMap: %T", value)
	}

	if len(bytes) == 0 {
		*m = make(JSONMap)
		return nil
	}

	var res map[string]interface{}
	if err := json.Unmarshal(bytes, &res); err != nil {
		return fmt.Errorf("failed to unmarshal JSONMap: %w", err)
	}

	*m = res
	return nil
}
