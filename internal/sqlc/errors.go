// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlc

// Error defines a constant error
type Error string

// Error implements the Errors interface
func (e Error) Error() string { return string(e) }

const (
	ErrApplyMigration        = Error("apply migration")
	ErrCreateSchema          = Error("create schema")
	ErrFetchOtto             = Error("fetch otto")
	ErrForeignKeysDisabled   = Error("foreign keys disabled")
	ErrInvalidPath           = Error("invalid path")
	ErrInvalidSchemaMetadata = Error("invalid schema metadata")
	ErrPragmaReturnedNil     = Error("pragma returned nil")
	ErrUnappliedMigrations   = Error("unapplied migrations")
)
