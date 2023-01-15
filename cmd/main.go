package main

import (
	"flag"
	"fmt"
	"strings"

	pms "github.com/Moranilt/pms/migrator"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type CmdMigrator struct {
	source  string
	up      bool
	down    bool
	host    string
	port    int
	db      string
	user    string
	pass    string
	version int
}

func (c *CmdMigrator) GetFlags() {
	flag.StringVar(&c.source, "source", "./migrations", "Source of migration files. For example './migrations'")
	flag.BoolVar(&c.up, "up", false, "Run all migrations from provided path")
	flag.BoolVar(&c.down, "down", false, "Run all down migrations from provided path")
	flag.StringVar(&c.host, "host", "localhost", "Database host")
	flag.IntVar(&c.port, "port", 5432, "Database port")
	flag.StringVar(&c.db, "db", "", "Database name")
	flag.StringVar(&c.user, "user", "root", "Database user")
	flag.StringVar(&c.pass, "pass", "", "Database password")
	flag.IntVar(&c.version, "v", -1, "Select version of migrations")
	flag.Parse()
}

func (c *CmdMigrator) Run() {
	c.db = strings.ToValidUTF8(strings.ReplaceAll(c.db, " ", ""), "")
	if c.db == "" || len(c.db) == 0 {
		fmt.Println("Error: 'db' flag required")
		return
	}

	if !c.up && !c.down && c.version == -1 {
		fmt.Println("Error: provide 'up', 'down' or 'version' flag")
		return
	}

	conn := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s port=%d sslmode=disable",
		c.host, c.db, c.user, c.pass, c.port,
	)
	db, err := sqlx.Connect("postgres", conn)
	if err != nil {
		fmt.Printf("Error while connecting to db: %v\n", err)
		return
	}
	defer db.Close()

	m, err := pms.New(db, c.source)
	if err != nil {
		fmt.Println(err)
		return
	}
	if c.up {
		err := m.Up()
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if c.down {
		err := m.Down()
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	if c.version != -1 {
		err := m.Version(c.version)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}

func main() {
	cmd := &CmdMigrator{}
	cmd.GetFlags()
	cmd.Run()
}
