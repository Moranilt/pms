package pms

import (
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewMigrator(t *testing.T) {
	f := FileTester{t: t}
	f.MakeTestDir()
	defer f.RemoveAll()
	db, mock := newSQlMock(t)
	defer db.Close()

	mock.ExpectPing()
	mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(QUERY_CREATE_TABLE, TABLE_NAME))).WillReturnResult(sqlmock.NewResult(0, 0))

	m, err := New(db, testDirname)
	if err != nil {
		t.Error(err)
	}
	if m == nil {
		t.Error("migrator is nil")
	}
}

func TestMigrationUp(t *testing.T) {
	f := FileTester{t: t}
	f.MakeTestDir()
	defer f.RemoveAll()

	db, mock := newSQlMock(t)
	defer db.Close()

	mock.ExpectPing()
	rows := mock.NewRows([]string{"version"}).AddRow(0)
	mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(QUERY_CREATE_TABLE, TABLE_NAME))).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(SELECT_VERSION).WillReturnRows(rows)

	files := []TestFile{
		{true, "1_users.up.sql", []byte("INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com')")},
		{false, "1_users.down.sql", []byte("DELETE FROM users WHERE name='Bobby' AND email='bob@mail.com'")},
		{true, "2_users.up.sql", []byte(`INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com'); INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com');`)},
	}

	mock.ExpectBegin()
	f.CreateFiles(files)
	f.CreateQueryMocks(files, mock)
	mock.ExpectExec(fmt.Sprintf(QUERY_UPDATE_VERSION, TABLE_NAME, 2)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	m, err := New(db, testDirname)
	if err != nil {
		t.Error(err)
	}
	if m == nil {
		t.Error("migrator is nil")
	}
	err = m.Up()
	if err != nil {
		t.Error(err)
	}
}

func TestMigrationDown(t *testing.T) {
	f := FileTester{t: t}
	f.MakeTestDir()
	defer f.RemoveAll()

	db, mock := newSQlMock(t)
	defer db.Close()

	mock.ExpectPing()
	rows := mock.NewRows([]string{"version"}).AddRow(2)
	mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(QUERY_CREATE_TABLE, TABLE_NAME))).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(SELECT_VERSION).WillReturnRows(rows)

	files := []TestFile{
		{false, "1_users.up.sql", []byte("INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com')")},
		{true, "1_users.down.sql", []byte("DROP TABLE users")},
		{false, "2_users.up.sql", []byte(`INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com'); INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com');`)},
		{true, "2_users.down.sql", []byte("DELETE FROM users WHERE name='Bobby' AND email='bob@mail.com'")},
		{false, "3_posts.down.sql", []byte("CREATE TABLE posts(id SERIAL);")},
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].name > files[j].name
	})

	mock.ExpectBegin()
	f.CreateFiles(files)
	f.CreateQueryMocks(files, mock)
	mock.ExpectExec(fmt.Sprintf(QUERY_UPDATE_VERSION, TABLE_NAME, 0)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	m, err := New(db, testDirname)
	if err != nil {
		t.Error(err)
	}
	if m == nil {
		t.Error("migrator is nil")
	}
	err = m.Down()
	if err != nil {
		t.Error(err)
	}
}

func TestMigratorVersion(t *testing.T) {
	t.Run("increase version", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		defer f.RemoveAll()

		db, mock := newSQlMock(t)
		defer db.Close()

		mock.ExpectPing()
		rows := mock.NewRows([]string{"version"}).AddRow(1)
		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(QUERY_CREATE_TABLE, TABLE_NAME))).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery(SELECT_VERSION).WillReturnRows(rows)

		files := []TestFile{
			{false, "1_users.up.sql", []byte("INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com')")},
			{false, "1_users.down.sql", []byte("DROP TABLE users")},
			{true, "2_users.up.sql", []byte(`INSERT INTO users (name, email) VALUES ('Melony', 'melony@mail.com');`)},
			{false, "2_users.down.sql", []byte("DELETE FROM users WHERE name='Bobby' AND email='bob@mail.com'")},
			{true, "3_posts.up.sql", []byte("CREATE TABLE posts(id SERIAL);")},
			{true, "4_comments.up.sql", []byte("CREATE TABLE comments(id SERIAL);")},
			{false, "5_comments.up.sql", []byte("CREATE TABLE comments(id SERIAL);")},
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].name < files[j].name
		})

		mock.ExpectBegin()
		f.CreateFiles(files)
		f.CreateQueryMocks(files, mock)
		mock.ExpectExec(fmt.Sprintf(QUERY_UPDATE_VERSION, TABLE_NAME, 4)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		m, err := New(db, testDirname)
		if err != nil {
			t.Error(err)
		}
		if m == nil {
			t.Error("migrator is nil")
		}
		err = m.Version(4)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("decrease version", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		defer f.RemoveAll()

		db, mock := newSQlMock(t)
		defer db.Close()

		mock.ExpectPing()
		rows := mock.NewRows([]string{"version"}).AddRow(2)
		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(QUERY_CREATE_TABLE, TABLE_NAME))).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery(SELECT_VERSION).WillReturnRows(rows)

		files := []TestFile{
			{false, "1_users.up.sql", []byte("INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com')")},
			{false, "1_users.down.sql", []byte("DROP TABLE users")},
			{false, "2_users.up.sql", []byte(`INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com'); INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com');`)},
			{true, "2_users.down.sql", []byte("DELETE FROM users WHERE name='Bobby' AND email='bob@mail.com'")},
			{false, "3_posts.up.sql", []byte("CREATE TABLE posts(id SERIAL);")},
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].name < files[j].name
		})

		mock.ExpectBegin()
		f.CreateFiles(files)
		f.CreateQueryMocks(files, mock)
		mock.ExpectExec(fmt.Sprintf(QUERY_UPDATE_VERSION, TABLE_NAME, 1)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		m, err := New(db, testDirname)
		if err != nil {
			t.Error(err)
		}
		if m == nil {
			t.Error("migrator is nil")
		}
		err = m.Version(1)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("current version", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		defer f.RemoveAll()

		db, mock := newSQlMock(t)
		defer db.Close()

		mock.ExpectPing()
		rows := mock.NewRows([]string{"version"}).AddRow(2)
		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf(QUERY_CREATE_TABLE, TABLE_NAME))).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery(SELECT_VERSION).WillReturnRows(rows)

		files := []TestFile{
			{false, "1_users.up.sql", []byte("INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com')")},
			{false, "1_users.down.sql", []byte("DROP TABLE users")},
			{false, "2_users.up.sql", []byte(`INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com'); INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com');`)},
			{true, "2_users.down.sql", []byte("DELETE FROM users WHERE name='Bobby' AND email='bob@mail.com'")},
			{false, "3_posts.up.sql", []byte("CREATE TABLE posts(id SERIAL);")},
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].name < files[j].name
		})

		f.CreateFiles(files)

		m, err := New(db, testDirname)
		if err != nil {
			t.Error(err)
		}
		if m == nil {
			t.Error("migrator is nil")
		}
		err = m.Version(2)
		if err.Error() != fmt.Sprintf(ERROR_EQUAL_VERSION, 2) {
			t.Errorf("not valid error message %q, expected %q", err, fmt.Sprintf(ERROR_EQUAL_VERSION, 2))
		}
	})
}
