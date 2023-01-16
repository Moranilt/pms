package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	pms "github.com/Moranilt/pms/migrator"
	"github.com/jmoiron/sqlx"
)

type mockedMigrator struct {
	up      bool
	down    bool
	version bool
}

func NewMockedMigrator() (CreateMigrator, *mockedMigrator) {
	m := &mockedMigrator{}
	return func(db *sqlx.DB, path string) (pms.Migrator, error) {
		return m, nil
	}, m
}

func (m *mockedMigrator) Up() error {
	m.up = true
	return nil
}
func (m *mockedMigrator) Down() error {
	m.down = true
	return nil
}
func (m *mockedMigrator) Version(version int) error {
	m.version = true
	return nil
}

func (m *mockedMigrator) Test(t *testing.T, args ...string) {
	t.Helper()
	for _, name := range args {
		switch name {
		case "up":
			if !m.up {
				t.Error("expected to call Up function")
			}
		case "down":
			if !m.down {
				t.Error("expected to call Down function")
			}
		case "version":
			if !m.version {
				t.Error("expected to call Version function")
			}
		}
	}
}

type mockedPms struct {
	mock sqlmock.Sqlmock
	db   *sqlx.DB
}

func CreateMockedMigrator() (*mockedPms, error) {
	mockedDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	db := sqlx.NewDb(mockedDB, "sqlmock")
	return &mockedPms{mock: mock, db: db}, nil
}

func (m *mockedPms) MakeDefaultMock() {
	m.mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(pms.QUERY_CREATE_TABLE, pms.TABLE_NAME))).WillReturnResult(sqlmock.NewResult(1, 1))
	rows := sqlmock.NewRows([]string{"version"}).AddRow(0)
	m.mock.ExpectQuery(pms.SELECT_VERSION).WillReturnRows(rows)
}

func (m *mockedPms) MakeFakeConnection(conn string) (*sqlx.DB, error) {
	return m.db, nil
}

func TestCmdMigrator(t *testing.T) {
	os.Mkdir(DEFAULT_SOURCE, 0777)
	defer os.RemoveAll(DEFAULT_SOURCE)
	t.Run("db empty", func(t *testing.T) {
		mockePMS, err := CreateMockedMigrator()
		if err != nil {
			t.Error(err)
		}
		m := &CmdMigrator{}
		err = m.Run(mockePMS.MakeFakeConnection)

		if err.Error() != ERROR_DB_REQUIRED {
			t.Errorf("got %q, expected %q", err.Error(), ERROR_DB_REQUIRED)
		}
	})

	t.Run("only db is set", func(t *testing.T) {
		m := &CmdMigrator{
			db:      "test_db",
			source:  DEFAULT_SOURCE,
			version: DEFAULT_VERSION,
		}

		mockePMS, err := CreateMockedMigrator()
		if err != nil {
			t.Error(err)
		}
		mockePMS.MakeDefaultMock()

		err = m.Run(mockePMS.MakeFakeConnection)

		if err.Error() != ERROR_NOT_PROVIDED_ACTION {
			t.Errorf("got %q, expected %q", err.Error(), ERROR_NOT_PROVIDED_ACTION)
		}
	})

	t.Run("only db and 'up' flag", func(t *testing.T) {
		createMigrator, migrator := NewMockedMigrator()
		m := New(createMigrator)
		m.up = true
		m.db = "test_db"

		mockePMS, err := CreateMockedMigrator()
		if err != nil {
			t.Error(err)
		}
		mockePMS.MakeDefaultMock()

		err = m.Run(mockePMS.MakeFakeConnection)
		if err != nil {
			t.Error(err)
		}
		migrator.Test(t, "up")
	})
	t.Run("only db and 'down' flag", func(t *testing.T) {
		createMigrator, migrator := NewMockedMigrator()
		m := New(createMigrator)
		m.down = true
		m.db = "test_db"

		mockePMS, err := CreateMockedMigrator()
		if err != nil {
			t.Error(err)
		}
		mockePMS.MakeDefaultMock()

		err = m.Run(mockePMS.MakeFakeConnection)
		if err != nil {
			t.Error(err)
		}
		migrator.Test(t, "down")
	})
	t.Run("only db and 'version' flag", func(t *testing.T) {
		createMigrator, migrator := NewMockedMigrator()
		m := New(createMigrator)
		m.version = 1
		m.db = "test_db"

		mockePMS, err := CreateMockedMigrator()
		if err != nil {
			t.Error(err)
		}
		mockePMS.MakeDefaultMock()

		err = m.Run(mockePMS.MakeFakeConnection)
		if err != nil {
			t.Error(err)
		}
		migrator.Test(t, "version")
	})
}
