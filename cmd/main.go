package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strings"

	"github.com/Moranilt/pms"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	DEFAULT_SOURCE   = "migrations"
	DEFAULT_VERSION  = -1
	DEFAULT_PORT     = 5432
	DEFAULT_HOST     = "localhost"
	DEFAULT_DB       = ""
	DEFAULT_USER     = "root"
	DEFAULT_PASS     = ""
	DEFAULT_SSL_MODE = "disable"
	DEFAULT_DRIVER   = "mysql"
	DEFAULT_URL      = ""

	ERROR_DB_REQUIRED         = "error: 'url' or 'db' flag required"
	ERROR_NOT_PROVIDED_ACTION = "error: provide 'up', 'down' or 'version' flag"
)

type CreateMigrator = func(db pms.DB, path string) (pms.Migrator, error)
type CmdMigrator struct {
	source         string
	up             bool
	down           bool
	host           string
	port           int
	db             string
	user           string
	pass           string
	version        int
	sslMode        string
	driver         string
	url            string
	createMigrator CreateMigrator
}

func New(c CreateMigrator) *CmdMigrator {
	return &CmdMigrator{
		createMigrator: c,
		source:         DEFAULT_SOURCE,
		host:           DEFAULT_HOST,
		pass:           DEFAULT_PASS,
		port:           DEFAULT_PORT,
		db:             DEFAULT_DB,
		user:           DEFAULT_USER,
		version:        DEFAULT_VERSION,
		sslMode:        DEFAULT_SSL_MODE,
		driver:         DEFAULT_DRIVER,
	}
}

type FlagType[T int | string | bool] struct {
	Pointer      *T
	Name         string
	DefaultValue T
	Usage        string
}

func (c *CmdMigrator) StringFlags() []FlagType[string] {
	return []FlagType[string]{
		{&c.source, "source", DEFAULT_SOURCE, "Source of migration files. For example './migrations'"},
		{&c.host, "host", DEFAULT_HOST, "Database host"},
		{&c.db, "db", DEFAULT_DB, "Database name"},
		{&c.user, "user", DEFAULT_USER, "Database user"},
		{&c.pass, "pass", DEFAULT_PASS, "Database password"},
		{&c.sslMode, "sslMode", DEFAULT_SSL_MODE, "Set ssl mode"},
		{&c.driver, "driver", DEFAULT_DRIVER, "Set MySQL driver"},
		{&c.url, "url", DEFAULT_URL, "Connection URL"},
	}
}

func (c *CmdMigrator) BoolFlags() []FlagType[bool] {
	return []FlagType[bool]{
		{&c.up, "up", false, "Run all migrations from provided path"},
		{&c.down, "down", false, "Run all down migrations from provided path"},
	}
}

func (c *CmdMigrator) IntFlags() []FlagType[int] {
	return []FlagType[int]{
		{&c.port, "port", DEFAULT_PORT, "Database port"},
		{&c.version, "v", DEFAULT_VERSION, "Select version of migrations"},
	}
}

func (c *CmdMigrator) GetFlags() {
	for _, f := range c.StringFlags() {
		flag.StringVar(f.Pointer, f.Name, f.DefaultValue, f.Usage)
	}

	for _, f := range c.BoolFlags() {
		flag.BoolVar(f.Pointer, f.Name, f.DefaultValue, f.Usage)
	}

	for _, f := range c.IntFlags() {
		flag.IntVar(f.Pointer, f.Name, f.DefaultValue, f.Usage)
	}
	flag.Parse()
}

func (c *CmdMigrator) MakeConnectionString() string {
	if c.url != "" {
		return strings.ToValidUTF8(c.url, "")
	}
	return fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s port=%d sslmode=%s",
		c.host, c.db, c.user, c.pass, c.port, c.sslMode,
	)
}

func (c *CmdMigrator) Run(makeConnection func(driver string, conn string) (pms.DB, error)) error {
	c.db = strings.ToValidUTF8(strings.ReplaceAll(c.db, " ", ""), "")
	if c.url == "" && c.db == "" {
		return fmt.Errorf(ERROR_DB_REQUIRED)
	}

	if !c.up && !c.down && c.version == -1 {
		return fmt.Errorf(ERROR_NOT_PROVIDED_ACTION)
	}

	db, err := makeConnection(c.driver, c.MakeConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	m, err := c.createMigrator(db, c.source)
	if err != nil {
		return err
	}
	if c.up {
		m.Up()
		return nil
	}
	if c.down {
		m.Down()
		return nil
	}
	if c.version != -1 {
		m.Version(c.version)
		return nil
	}
	return nil
}

func makeConnection(diver string, conn string) (pms.DB, error) {
	db, err := sql.Open(diver, conn)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to db: %w", err)
	}

	return db, nil
}

func main() {
	cmd := New(pms.New)
	cmd.GetFlags()
	err := cmd.Run(makeConnection)
	if err != nil {
		fmt.Println(err)
	}
}
