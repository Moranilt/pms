package pms

import (
	"io/fs"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

const testDirname = "./test_folder"

type FileTester struct {
	t *testing.T
}

type TestFile struct {
	valid   bool
	name    string
	content []byte
}

func (f *FileTester) MakeTestDir() {
	f.t.Helper()
	err := os.Mkdir(testDirname, 0777)
	if err != nil {
		f.t.Error(err)
	}
}

func (f *FileTester) RemoveAll() {
	os.RemoveAll(testDirname)
}

func (f *FileTester) CreateFiles(files []TestFile) {
	f.t.Helper()

	for _, file := range files {
		path := testDirname + "/" + file.name
		_, err := os.Create(path)
		if err != nil {
			f.t.Error(err)
		}
		os.WriteFile(path, file.content, fs.FileMode(os.O_APPEND))
	}
}

func (f *FileTester) CreateQueryMocks(files []TestFile, mock sqlmock.Sqlmock) {
	f.t.Helper()
	for _, file := range files {
		if !file.valid {
			continue
		}
		mock.ExpectExec(regexp.QuoteMeta(string(file.content))).WillReturnResult(sqlmock.NewResult(1, 1))
	}
}

func TestFileTester(t *testing.T) {
	t.Run("make directory", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		defer f.RemoveAll()

		_, err := os.Stat(testDirname)
		if os.IsNotExist(err) {
			t.Errorf("directory %q does not exist", testDirname)
		}

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("remove directory", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		f.RemoveAll()

		_, err := os.Stat(testDirname)
		if os.IsExist(err) {
			t.Errorf("directory %q should not be exist", testDirname)
		}
	})

	t.Run("create files", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		defer f.RemoveAll()

		files := []TestFile{
			{true, "test_file.txt", []byte("test content")},
			{false, "not_created.txt", []byte("not created content")},
		}
		f.CreateFiles(files)

		dfs, err := os.ReadDir(testDirname)
		if err != nil {
			t.Error(err)
		}

		for _, file := range files {
			var found bool
			for _, currentFile := range dfs {
				if file.name == currentFile.Name() {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("file not found %q", file.name)
			}
		}
	})

	t.Run("create query mocks", func(t *testing.T) {
		f := FileTester{t: t}
		f.MakeTestDir()
		defer f.RemoveAll()

		files := []TestFile{
			{true, "test_file.txt", []byte("test content")},
			{false, "not_created.txt", []byte("not created content")},
			{true, "test_file_2.txt", []byte("test file 2")},
		}
		f.CreateFiles(files)
		fakeMock := &fakeSqlmock{files: files, t: t}
		f.CreateQueryMocks(files, fakeMock)
	})
}

type fakeSqlmock struct {
	iterator int
	files    []TestFile
	t        *testing.T
}

func (f *fakeSqlmock) ExpectClose() *sqlmock.ExpectedClose                       { return nil }
func (f *fakeSqlmock) ExpectationsWereMet() error                                { return nil }
func (f *fakeSqlmock) ExpectPrepare(expectedSQL string) *sqlmock.ExpectedPrepare { return nil }
func (f *fakeSqlmock) ExpectQuery(expectedSQL string) *sqlmock.ExpectedQuery     { return nil }
func (f *fakeSqlmock) ExpectBegin() *sqlmock.ExpectedBegin                       { return nil }
func (f *fakeSqlmock) ExpectCommit() *sqlmock.ExpectedCommit                     { return nil }
func (f *fakeSqlmock) ExpectRollback() *sqlmock.ExpectedRollback                 { return nil }
func (f *fakeSqlmock) ExpectPing() *sqlmock.ExpectedPing                         { return nil }
func (f *fakeSqlmock) MatchExpectationsInOrder(d bool)                           {}
func (f *fakeSqlmock) NewRows(columns []string) *sqlmock.Rows                    { return nil }
func (f *fakeSqlmock) NewRowsWithColumnDefinition(columns ...*sqlmock.Column) *sqlmock.Rows {
	return nil
}
func (f *fakeSqlmock) NewColumn(name string) *sqlmock.Column { return nil }
func (f *fakeSqlmock) ExpectExec(expectedSQL string) *sqlmock.ExpectedExec {
	f.t.Helper()
	if f.iterator >= len(f.files) {
		f.t.Errorf("not expected call %q", expectedSQL)
		return &sqlmock.ExpectedExec{}
	}
	for !f.files[f.iterator].valid {
		f.iterator++
	}
	file := f.files[f.iterator]
	f.iterator++

	if string(file.content) != expectedSQL {
		f.t.Errorf("expected %q instead of %q", string(file.content), expectedSQL)
		return &sqlmock.ExpectedExec{}
	}

	return &sqlmock.ExpectedExec{}
}
