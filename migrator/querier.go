package pms

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"

	"github.com/jmoiron/sqlx"
)

type skipFileFunc = func(fileVersion int) bool

type querier struct {
	tx   *sql.Tx
	path string
}

// path - folder path
func newQuerier(db *sqlx.DB, path string) (*querier, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &querier{tx: tx, path: path}, nil
}

func (q *querier) Add(path string, fileName string) error {
	content, err := getFileContent(path, fileName)
	if err != nil {
		q.tx.Rollback()
		return err
	}
	_, err = q.tx.Exec(string(content))
	if err != nil {
		q.tx.Rollback()
		return fmt.Errorf(
			"cannot execute file %q with content content: %q. Error: %w",
			fileName,
			string(content),
			err,
		)
	}

	return nil
}

func (q *querier) Exec(query string, args ...any) (sql.Result, error) {
	if len(args) != 0 {
		return q.tx.Exec(query, args)
	}
	return q.tx.Exec(query)
}

func (q *querier) Rollback() {
	q.tx.Rollback()
}

func (q *querier) Commit() error {
	return q.tx.Commit()
}

func (q *querier) RunFileQueries(version int, filesToRead []fs.DirEntry, direction Direction, skipFile skipFileFunc) error {
	switch direction {
	case DIRECTION_UP:
		for _, file := range filesToRead {
			filenameChunks := strings.Split(file.Name(), ".")
			name := filenameChunks[0]
			fileVersion := getVersionFromName(name)
			if skipFile(fileVersion) {
				continue
			}
			if fileVersion > version {
				version = fileVersion
			}
			err := q.Add(q.path, file.Name())
			if err != nil {
				q.Rollback()
				return err
			}
		}
	case DIRECTION_DOWN:
		for i := len(filesToRead) - 1; i >= 0; i-- {
			file := filesToRead[i]
			filenameChunks := strings.Split(file.Name(), ".")
			name := filenameChunks[0]
			fileVersion := getVersionFromName(name)
			if skipFile(fileVersion) {
				continue
			}
			err := q.Add(q.path, file.Name())
			if err != nil {
				q.Rollback()
				return err
			}
		}
	default:
		return fmt.Errorf("unhandled direction %q", direction)
	}

	_, err := q.Exec(fmt.Sprintf(QUERY_UPDATE_VERSION, TABLE_NAME, version))
	if err != nil {
		q.Rollback()
		return err
	}
	err = q.Commit()
	if err != nil {
		return err
	}
	return nil
}
