package pms

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	SELECT_VERSION = "SELECT version FROM migrations"
)

func getFilesWithDirection(files []fs.DirEntry, inc Direction) ([]fs.DirEntry, error) {
	var filesToRead []fs.DirEntry
	sort.Slice(filesToRead, func(i, j int) bool {
		return filesToRead[i].Name() > filesToRead[j].Name()
	})

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filenameChunks := strings.Split(file.Name(), ".")
		if len(filenameChunks) < 3 {
			return nil, fmt.Errorf("file %q should have an extension", file.Name())
		}

		if filenameChunks[1] == string(inc) {
			filesToRead = append(filesToRead, file)
		}
	}

	return filesToRead, nil
}

func getFileContent(path string, fileName string) ([]byte, error) {
	file, err := fs.ReadFile(os.DirFS(path), fileName)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func getVersionFromName(name string) int {
	var version int
	for _, s := range name {
		if s == '_' {
			break
		}
		if s >= '0' && s <= '9' {
			version = version*10 + int(s-'0')
		}
	}

	return version
}

func readDir(path string) ([]fs.DirEntry, error) {
	if files, err := os.ReadDir(path); err != nil {
		return nil, err
	} else {
		return files, err
	}
}

func createTable(db *sqlx.DB, tableName string) error {
	_, err := db.Exec(fmt.Sprintf(QUERY_CREATE_TABLE, tableName))

	if err != nil {
		return fmt.Errorf("cannot create table %q: %w", tableName, err)
	}

	return nil
}

func getMigrationVersion(db *sqlx.DB) (int, error) {
	var migrationVersion int
	err := db.Get(&migrationVersion, SELECT_VERSION)
	if err != nil {
		return 0, err
	}

	return migrationVersion, nil
}

func tableExists(db *sqlx.DB, tableName string) bool {
	_, tableCheck := db.Query("SELECT * FROM " + tableName + ";")
	return tableCheck == nil
}