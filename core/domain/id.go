package domain

import "github.com/google/uuid"

// ID is a domain identifier backed by a UUID v7.
type ID struct{ uuid uuid.UUID }

// NewID generates a new time-ordered unique identifier.
func NewID() ID {
	return ID{uuid: uuid.Must(uuid.NewV7())}
}

// ParseID parses a string into an ID, returning a CodeValidation domain error on failure.
func ParseID(s string) (ID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID{}, NewError(CodeValidation, "invalid ID")
	}
	return ID{uuid: u}, nil
}

// IDFrom wraps an existing uuid.UUID as an ID.
func IDFrom(u uuid.UUID) ID {
	return ID{uuid: u}
}

func (id ID) String() string  { return id.uuid.String() }
func (id ID) UUID() uuid.UUID { return id.uuid }
func (id ID) IsZero() bool    { return id.uuid == uuid.Nil }
