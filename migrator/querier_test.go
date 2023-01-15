package pms

import (
	"testing"
)

func TestQuerier(t *testing.T) {
	t.Run("NewQuerier", func(t *testing.T) {
		db, mock := newSQlMock(t)
		defer db.Close()
		mock.ExpectBegin()

		_, err := newQuerier(db, "")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Commit", func(t *testing.T) {
		db, mock := newSQlMock(t)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectCommit()
		q, err := newQuerier(db, "")
		if err != nil {
			t.Error(err)
		}
		q.Commit()
	})

	t.Run("Rollback", func(t *testing.T) {
		db, mock := newSQlMock(t)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()
		q, err := newQuerier(db, "")
		if err != nil {
			t.Error(err)
		}
		q.Rollback()
	})

	t.Run("Add", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		defer f.RemoveAll()

		db, mock := newSQlMock(t)
		defer db.Close()

		mock.ExpectBegin()

		files := []TestFile{
			{true, "1_users.up.sql", []byte("INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com')")},
			{true, "1_users.down.sql", []byte("DROP TABLE users")},
			{true, "2_users.up.sql", []byte(`INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com'); INSERT INTO users (name, email) VALUES ('Bobby', 'bob@mail.com');`)},
			{true, "2_users.down.sql", []byte("DELETE FROM users WHERE name='Bobby' AND email='bob@mail.com'")},
		}
		f.CreateFiles(files)
		f.CreateQueryMocks(files, mock)

		q, err := newQuerier(db, "")
		if err != nil {
			t.Error(err)
		}
		for _, file := range files {
			q.Add(testDirname, file.name)
		}
	})
}
