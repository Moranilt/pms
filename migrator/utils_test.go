package pms

import (
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func newSQlMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Error(err)
	}
	return sqlx.NewDb(mockDB, "sqlmock"), mock
}

func TestCreateTable(t *testing.T) {
	db, mock := newSQlMock(t)

	mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(QUERY_CREATE_TABLE, "test_table"))).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(0, 0))
	err := createTable(db, "test_table")
	if err != nil {
		t.Error(err)
	}
}
func TestGetVersionFromName(t *testing.T) {
	fileNames := []struct {
		version int
		name    string
	}{
		{1, "1_test.up.sql"},
		{100, "100_test.up.sql"},
		{1006, "100h6_test.up.sql"},
		{0, "dumb.up.sql"},
		{2, "2.up.sql"},
		{0, ""},
	}

	for _, data := range fileNames {
		t.Run(fmt.Sprintf("version %d with fileName %q", data.version, data.name), func(t *testing.T) {
			version := getVersionFromName(data.name)

			if version != data.version {
				t.Errorf("not expected version %d, expected %d", version, data.version)
			}
		})
	}
}

func TestGetFileContent(t *testing.T) {
	f := FileTester{t: t}
	f.MakeTestDir()
	defer f.RemoveAll()

	file, err := os.Create(testDirname + "/test.txt")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	fileContent := []byte("CREATE TABLE test(name VARCHAR)")
	os.WriteFile(testDirname+"/test.txt", fileContent, fs.FileMode(os.O_APPEND))

	bytes, err := getFileContent(testDirname, "test.txt")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(bytes, fileContent) {
		t.Errorf("not valid content %v, expected %v", bytes, fileContent)
	}
}

func TestReadDir(t *testing.T) {
	f := FileTester{t: t}
	f.MakeTestDir()
	defer f.RemoveAll()

	os.Create(testDirname + "/test.txt")

	files, err := readDir(testDirname)
	if err != nil {
		t.Error(err)
	}

	if len(files) == 0 {
		t.Error("length of files is equal to 0, expected 1")
	}
}

func TestGetMigrationVersion(t *testing.T) {
	db, mock := newSQlMock(t)

	expectedVersion := 2
	rows := mock.NewRows([]string{"version"}).AddRow(expectedVersion)
	mock.ExpectQuery(SELECT_VERSION).WillReturnRows(rows)

	version, err := getMigrationVersion(db)
	if err != nil {
		t.Error(err)
	}

	if version != expectedVersion {
		t.Errorf("expected version %d, got %d", expectedVersion, version)
	}
}
