package utilid

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UtilID struct {
	Data  uuid.UUID
	Valid bool
}

func New() *UtilID {
	id := uuid.New()
	return &UtilID{Valid: true, Data: id}
}

func FromString(value string) *UtilID {
	if value == "" {
		return &UtilID{}
	}

	id, err := uuid.Parse(value)

	if err != nil {
		return &UtilID{}
	}

	return &UtilID{Data: id, Valid: true}
}

func FromPg(value pgtype.UUID) *UtilID {
	if !value.Valid || value.String() == "" {
		return &UtilID{}
	}

	id, err := uuid.Parse(value.String())

	if err != nil {
		return &UtilID{}
	}

	return &UtilID{Data: id, Valid: true}
}

func (d UtilID) AsGoogleUUID() uuid.UUID {
	return d.Data
}

func (d UtilID) AsPgUUID() pgtype.UUID {
	return pgtype.UUID{Valid: d.Valid, Bytes: [16]byte(d.Data)}
}

func (d UtilID) AsString() string {
	return d.Data.String()
}
