package pms

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	TABLE_NAME         = "migrations"
	QUERY_CREATE_TABLE = `CREATE TABLE %s (
		version VARCHAR(255) NOT NULL DEFAULT 0
	);
	INSERT INTO migrations (version) VALUES (0);`
	QUERY_UPDATE_VERSION = "UPDATE %s SET version=%d"
	ERROR_EQUAL_VERSION  = "current version %d equals current"
)

type Direction string

const (
	DIRECTION_UP   Direction = "up"
	DIRECTION_DOWN Direction = "down"
)

type Migrator interface {
	Up() error
	Down() error
	Version(int) error
}
type Migration struct {
	db   *sqlx.DB
	path string
}

func New(db *sqlx.DB, path string) (Migrator, error) {
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

	return &Migration{db: db, path: path}, nil
}

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
		fmt.Printf("the selected version %d is greater than the latest version in files %d. Latest version will be set to %d.\n",
			version,
			latestFileVersion,
			latestFileVersion,
		)
	} else if migrationVersion == latestFileVersion && version > latestFileVersion {
		return fmt.Errorf("migrations is up to date")
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
