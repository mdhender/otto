// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlc

import (
	"context"
	"database/sql"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"sort"
	"strings"
)

const (
	expectedSchemaVersion = "0.0.1"
)

var (
	//go:embed schema.sql
	schemaDDL string

	//go:embed migrations/*.sql
	migrationFS embed.FS
)

type DB struct {
	Path    string
	DB      *sql.DB
	Ctx     context.Context
	Queries *Queries
}

// CreateDatabase ensures that the database exists.
// If the database doesn't exist, it is created.
// It returns an error if the database cannot be created.
func CreateDatabase(path string) error {
	if sb, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return ErrInvalidPath
		}
	} else if sb.IsDir() {
		return ErrInvalidPath
	} else if !sb.Mode().IsRegular() {
		return ErrInvalidPath
	}

	// create the database.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// we need to create the schema if there is no migrations table.
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name ='migrations'")
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	if rows.Next() == false {
		// no tables, so create the schema
		if _, err = db.Exec(schemaDDL); err != nil {
			return errors.Join(ErrCreateSchema, err)
		}
	}

	return db.Close()
}

// MigrateSchema applies missing migrations to the database.
// It returns an error if any of the migration scripts fail.
func MigrateSchema(path string) error {
	db, err := OpenDatabase(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	//q, ctx := New(db), context.Background()

	// gather the list of migrations to apply
	var allMigrations []string
	if files, err := migrationFS.ReadDir("migrations"); err != nil {
		return err
	} else {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			allMigrations = append(allMigrations, file.Name())
		}
	}
	sort.Strings(allMigrations)
	log.Printf("all migrations: %v\n", allMigrations)

	// gather the list of migrations that have already been applied
	appliedMigrations := map[string]bool{}
	if rows, err := db.Query("SELECT id FROM migrations"); err != nil {
		return err
	} else {
		defer func() {
			_ = rows.Close()
		}()

		// loop through the rows, collecting the names into the existingMigrations slice.
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			appliedMigrations[id] = true
		}
	}
	log.Printf("applied migrations: %v\n", appliedMigrations)

	// loop through the migrations, applying them in order.
	for _, migrationSql := range allMigrations {
		id := strings.TrimSuffix(migrationSql, ".sql")
		if appliedMigrations[id] {
			continue
		}
		log.Printf("applying migration: %s\n", id)
		dml, err := migrationFS.ReadFile("migrations/" + migrationSql)
		if err != nil {
			log.Fatalf("error reading migration: %v\n", err)
		}
		// we need to apply the migration in a transaction.
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("error starting transaction: %v\n", err)
		}
		_, err = tx.Exec(string(dml))
		if err != nil {
			_ = tx.Rollback()
			return errors.Join(fmt.Errorf("%s: %w", migrationSql, ErrApplyMigration), err)
		}
		err = tx.Commit()
		if err != nil {
			log.Fatalf("error committing transaction: %v\n", err)
		}
		appliedMigrations[id] = true
	}

	// as a sanity check, log any un-applied migrations and return an error if any were found.
	found := false
	for _, migrationSql := range allMigrations {
		id := strings.TrimSuffix(migrationSql, ".sql")
		log.Printf("sanity checking migration: %s\n", id)
		if !appliedMigrations[id] {
			log.Printf("unapplied migration: %s\n", migrationSql)
			found = true
		}
	}
	if found {
		return ErrUnappliedMigrations
	}

	//metadataRows, err := q.FetchSchemaMetadata(ctx)
	//if err != nil {
	//	return err
	//} else if len(metadataRows) != 1 {
	//	return ErrInvalidSchemaMetadata
	//} else if metadataRows[0].Version != expectedSchemaVersion {
	//	return ErrInvalidSchemaMetadata
	//}
	//
	//// if otto does not have a magic key or password, then create one.
	//otto, err := q.FetchUser(ctx, "otto")
	//if err != nil {
	//	return errors.Join(ErrFetchOtto, err)
	//}
	//if otto.Magic == "" {
	//	otto.Magic = uuid.New().String()
	//	if err = q.UpdateOttoMagic(ctx, otto.Magic); err != nil {
	//		return err
	//	}
	//	if err = q.UpSetUserMagic(ctx, "otto", "otto"); err != nil {
	//		return errors.Join(ErrSetOttoMagic, err)
	//	}
	//}

	return nil
}

// OpenDatabase returns a new database connection if the database exists.
// It returns an error if the database cannot be opened.
func OpenDatabase(path string) (*sql.DB, error) {
	if sb, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrInvalidPath
		}
	} else if sb.IsDir() {
		return nil, ErrInvalidPath
	} else if !sb.Mode().IsRegular() {
		return nil, ErrInvalidPath
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// confirm that the database has foreign keys enabled
	var rslt sql.Result
	checkPragma := "PRAGMA" + " foreign_keys = ON"
	if rslt, err = db.Exec(checkPragma); err != nil {
		return nil, errors.Join(ErrForeignKeysDisabled, err)
	} else if rslt == nil {
		return nil, ErrPragmaReturnedNil
	}

	return db, nil
}

func (db *DB) CloseDatabase() {
	if db != nil && db.DB != nil {
		_ = db.DB.Close()
		db.DB = nil
	}
}
