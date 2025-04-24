package entity

import (
	"encoding/json"
	"time"
)

// Entity represents a generic entity with ID and metadata
type Entity struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Deployment  string                 `json:"deployment"`
	Timestamp   time.Time              `json:"timestamp"`
	Cursor      string                 `json:"cursor,omitempty"`
	Data        map[string]interface{} `json:"data"`
	MetaData    map[string]interface{} `json:"meta_data,omitempty"`
}

// GraphResponse represents the raw response from TheGraph API
type GraphResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []GraphError           `json:"errors,omitempty"`
}

// GraphError represents errors returned by TheGraph API
type GraphError struct {
	Message   string                 `json:"message"`
	Locations []GraphErrorLocation   `json:"locations,omitempty"`
	Path      []string               `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphErrorLocation represents the location of an error in a GraphQL query
type GraphErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// MarshalForEvent serializes the entity for use in a message bus
func (e *Entity) MarshalForEvent() ([]byte, error) {
	return MarshalJSON(e)
}

// UnmarshalFromEvent deserializes the entity from a message bus payload
func UnmarshalFromEvent(data []byte) (*Entity, error) {
	var entity Entity
	err := UnmarshalJSON(data, &entity)
	return &entity, err
}

// MarshalJSON marshals the given interface into JSON
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON unmarshals JSON data into the given interface
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
} 