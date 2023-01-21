[![Go Reference](https://pkg.go.dev/badge/github.com/Moranilt/pms.svg)](https://pkg.go.dev/github.com/Moranilt/pms)

# pms
Personal Migration System. CLI and GO package.

# Homebrew
[Pull Request](https://github.com/Homebrew/homebrew-core/pull/120892)

To make it easier to install with Homebrew, this package should be more notable to deploy it(>30 forks, >=30 watchers and >=75 stars). Please support this package!

# Supported drivers
- MySQL (4.1+)
- MariaDB
- Percona Server
- Google CloudSQL or Sphinx (2.2.3+)
- PostgreSQL

# How to use

## Install package
```bash
go get github.com/Moranilt/pms
```

## Make folder
First of all you should create a folder where to store migration-files:

```bash
mkdir migrations
```

## Make files
To create a migration file you should follow template `{version}_{any_name}.{action}.sql` where:
- `version` - version of migrations
- `any_name` - name which associated with queries inside(whatever you want)
- `action` - only `up` or `down`

### Example:
```
- 1_users.up.sql
- 1_users.down.sql
- 2_posts.up.sql
- 2_posts.down.sql
```

**1_users.up.sql**:
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL
);
```

**1_users.down.sql**:
```sql
DROP TABLE users;
```

**2_posts.up.sql**:
```sql
CREATE TABLE posts (
  id SERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**2_posts.down.sql**:
```sql
DROP TABLE posts;
```

## Run
After first run it'll create `migrations` table in your DB. **Do not delete or update it!**
### Inside of GO application

#### Up
To run all migrations you should use `Up` method which will run all files with `up` action in file.

If current version `2` and you have files with version up to `5`, `Up` method will run all files from `3` to `5` versions.

Example:

```go
package main

import (
	"log"

	"github.com/Moranilt/pms"
	"github.com/jmoiron/sqlx"
)

func main() {
	conn := "host=localhost dbname=postgres user=root password=1234 port=5432 sslmode=disable"
	db, err := sqlx.Connect("postgres", conn)
	if err != nil {
		log.Fatal("error while connecting to db", err)
	}

	migrator, err := pms.New(db, "./migrations")
	if err != nil {
		log.Fatal(err)
	}

	err = migrator.Up()
	if err != nil {
		log.Fatal("failed to run migrations: ", err)
	}
}
```

#### Down
To run all files with `down` action you should use `Down` method. 

It will run all files with `down` actions with descending order starts from the latest version of migration.

```go
package main

import (
	"log"

	"github.com/Moranilt/pms"
	"github.com/jmoiron/sqlx"
)

func main() {
	conn := "host=localhost dbname=postgres user=root password=1234 port=5432 sslmode=disable"
	db, err := sqlx.Connect("postgres", conn)
	if err != nil {
		log.Fatal("error while connecting to db", err)
	}

	migrator, err := pms.New(db, "./migrations")
	if err != nil {
		log.Fatal(err)
	}

	err = migrator.Down()
	if err != nil {
		log.Fatal("failed to run down migrations: ", err)
	}
}
```

#### Version
You can chose the version to jump to with `Version` method. Works like checkout to specific version.

If current version is `5` and you will run `migrator.Version(7)`, it will execute all migration files with `up` action from `6` to `7` versions.

If current version is `5` and you will run `migrator.Version(2)`, it will execute all migration files with `down` action from `5` to `3` versions.

```go
package main

import (
	"log"

	"github.com/Moranilt/pms"
	"github.com/jmoiron/sqlx"
)

func main() {
	conn := "host=localhost dbname=postgres user=root password=1234 port=5432 sslmode=disable"
	db, err := sqlx.Connect("postgres", conn)
	if err != nil {
		log.Fatal("error while connecting to db", err)
	}

	migrator, err := pms.New(db, "./migrations")
	if err != nil {
		log.Fatal(err)
	}

	err = migrator.Version(3)
	if err != nil {
		log.Fatal("failed to run version migrations: ", err)
	}

  	err = migrator.Version(5)
	if err != nil {
		log.Fatal("failed to run version migrations: ", err)
	}
}
```

### CMD
You can find binaries for your system in [releases](https://github.com/Moranilt/pms/releases).

Allowed flags:

**--help** - display all available flags \
**-db** string - Database name \
**-down** - Run all down migrations from provided path \
**-host** string - Database host (default "localhost") \
**-pass** string - Database password \
**-port** int - Database port (default 5432) \
**-source** string - Source of migration files. For example './migrations' (default "migrations") \
**-up** - Run all migrations from provided path \
**-user** string - Database user (default "root") \
**-v** int - Select version of migrations (default -1) \
**-sslMode** string - Set ssl mode (default "disable") \
**-driver** string - Set MySQL driver (default "mysql") \
**-url** string - Connection URL. `[driver]://[user]:[pass]@[host]:[port]/[db_name]?[flag_name]=[flag_value]`

Example `Up`:
```bash
pms -driver postgres -db postgres -host localhost -pass secret_pass -source migrations -user root -up
```

Example `Down`:
```bash
pms -db postgres -host localhost -pass secret_pass -source migrations -user root -down
```

Example `Version`:
```bash
pms -db postgres -host localhost -pass secret_pass -source migrations -user root -v 5
```

Example URL:
```bash
pms -driver postgres -url "postgres://root:123456@localhost:5432/authentication?sslmode=disable" -up
```
