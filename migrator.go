package pms

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	TABLE_NAME         = "migrations"
	QUERY_CREATE_TABLE = `CREATE TABLE %s (
		version VARCHAR(255) NOT NULL DEFAULT 0
	);
	INSERT INTO migrations (version) VALUES (0);`
	QUERY_UPDATE_VERSION = "UPDATE %s SET version=%d"

	ERROR_EQUAL_VERSION = "current version %d equals current"
	ERROR_UP_TO_DATE    = "migrations is up to date"
)

type Direction string

const (
	DIRECTION_UP   Direction = "up"
	DIRECTION_DOWN Direction = "down"
)

type DB interface {
	Begin() (*sql.Tx, error)
	Close() error
	Exec(query string, args ...any) (sql.Result, error)
	Ping() error
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type Migrator interface {
	Up() error
	Down() error
	Version(int) error
}
type Migration struct {
	db   DB
	path string
	l    Logger
}

// Create new instance of Migration structure
func New(db DB, path string) (Migrator, error) {
	_, err := readDir(path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if !tableExists(db, TABLE_NAME) {
		err = createTable(db, TABLE_NAME)

		if err != nil {
			return nil, err
		}
	}

	return &Migration{db: db, path: path, l: newEventLogger()}, nil
}

// Run all queries from files with `up` action.
func (m *Migration) Up() error {
	files, err := readDir(m.path)
	if err != nil {
		return err
	}
	filesToRead, err := getFilesWithDirection(files, DIRECTION_UP)
	if err != nil {
		return err
	}

	migrationVersion, err := getMigrationVersion(m.db)
	if err != nil {
		return err
	}

	q, err := newQuerier(m.db, m.path)
	if err != nil {
		return nil
	}

	var skipFile skipFileFunc = func(fileVersion int) bool {
		return fileVersion <= migrationVersion
	}

	err = q.RunFileQueries(-1, filesToRead, DIRECTION_UP, skipFile)
	if err != nil {
		return err
	}

	return nil
}

// Run all queries from files with `down` action.
func (m *Migration) Down() error {
	files, err := readDir(m.path)
	if err != nil {
		return err
	}
	filesToRead, err := getFilesWithDirection(files, DIRECTION_DOWN)
	if err != nil {
		return err
	}

	migrationVersion, err := getMigrationVersion(m.db)
	if err != nil {
		return err
	}

	q, err := newQuerier(m.db, m.path)
	if err != nil {
		return nil
	}
	var skipFile skipFileFunc = func(fileVersion int) bool {
		return fileVersion > migrationVersion
	}

	err = q.RunFileQueries(0, filesToRead, DIRECTION_DOWN, skipFile)
	if err != nil {
		return err
	}

	return nil
}

// Switch to the specified version.
//
// If specified version greater than current it'll run queries
// with `up` action.
//
// If specified version lower that current it'll run queries
// with `down` action.
//
// Otherwise it'll return an error.
func (m *Migration) Version(version int) error {
	files, err := readDir(m.path)
	if err != nil {
		return err
	}

	migrationVersion, err := getMigrationVersion(m.db)
	if err != nil {
		return err
	}
	var direction Direction

	if version > migrationVersion {
		direction = DIRECTION_UP
	} else if version == migrationVersion {
		return fmt.Errorf(ERROR_EQUAL_VERSION, version)
	} else {
		direction = DIRECTION_DOWN
	}
	filesToRead, err := getFilesWithDirection(files, direction)
	if err != nil {
		return err
	}

	latestFileVersion := getVersionFromName(filesToRead[len(filesToRead)-1].Name())
	if version > latestFileVersion && migrationVersion < latestFileVersion {
		version = latestFileVersion
		m.l.Warn(fmt.Sprintf("the selected version %d is greater than the latest version in files %d. Latest version will be set to %d.",
			version,
			latestFileVersion,
			latestFileVersion,
		))
	} else if migrationVersion == latestFileVersion && version > latestFileVersion {
		return fmt.Errorf(ERROR_UP_TO_DATE)
	}

	var skipFile skipFileFunc
	if direction == DIRECTION_UP {
		skipFile = func(fileVersion int) bool {
			return fileVersion <= migrationVersion || fileVersion > version
		}
	} else {
		skipFile = func(fileVersion int) bool {
			return fileVersion <= version || fileVersion > migrationVersion
		}
	}

	q, err := newQuerier(m.db, m.path)
	if err != nil {
		return nil
	}
	err = q.RunFileQueries(version, filesToRead, direction, skipFile)
	if err != nil {
		return err
	}

	return nil
}
