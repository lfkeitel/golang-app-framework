// +build dbsqlite dball

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/lfkeitel/golang-app-framework/src/utils"
	"github.com/lfkeitel/verbose"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func init() {
	RegisterDatabaseAccessor("sqlite", newSQLiteDBInit())
}

type sqliteDB struct {
	createFuncs  map[string]func(*utils.DatabaseAccessor) error
	migrateFuncs []func(*utils.DatabaseAccessor) error
}

func newSQLiteDBInit() *sqliteDB {
	s := &sqliteDB{}

	s.createFuncs = map[string]func(*utils.DatabaseAccessor) error{
		"settings": s.createSettingTable,
	}

	s.migrateFuncs = []func(*utils.DatabaseAccessor) error{}

	return s
}

func (s *sqliteDB) connect(d *utils.DatabaseAccessor, c *utils.Config) error {
	var err error
	if err = os.MkdirAll(path.Dir(c.Database.Address), os.ModePerm); err != nil {
		return fmt.Errorf("Failed to create directories: %v", err)
	}
	d.DB, err = sql.Open("sqlite3", c.Database.Address)
	if err != nil {
		return err
	}

	err = d.DB.Ping()
	if err != nil {
		return err
	}

	_, err = d.Exec("PRAGMA foreign_keys = ON")
	return err
}

func (s *sqliteDB) createTables(d *utils.DatabaseAccessor) error {
	rows, err := d.DB.Query(`SELECT name FROM sqlite_master WHERE type='table'`)
	if err != nil {
		return err
	}
	defer rows.Close()
	tables := make(map[string]bool)
	for _, table := range utils.DatabaseTableNames {
		tables[table] = false
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables[tableName] = true
	}

	for table, create := range s.createFuncs {
		if !tables[table] {
			if err := create(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *sqliteDB) migrateTables(d *utils.DatabaseAccessor) error {
	var currDBVer int
	verRow := d.DB.QueryRow(`SELECT "value" FROM "settings" WHERE "id" = 'db_version'`)
	if verRow == nil {
		return errors.New("Failed to get database version")
	}
	verRow.Scan(&currDBVer)

	utils.SystemLogger.WithFields(verbose.Fields{
		"current-version": currDBVer,
		"active-version":  dbVersion,
	}).Debug("Database Versions")

	// No migration needed
	if currDBVer == dbVersion {
		return nil
	}

	neededMigrations := s.migrateFuncs[currDBVer:dbVersion]
	for _, migrate := range neededMigrations {
		if migrate == nil {
			continue
		}
		if err := migrate(d); err != nil {
			return err
		}
	}

	_, err := d.DB.Exec(`UPDATE "settings" SET "value" = ? WHERE "id" = 'db_version'`, dbVersion)
	return err
}

func (s *sqliteDB) init(d *utils.DatabaseAccessor, c *utils.Config) error {
	if err := s.connect(d, c); err != nil {
		return err
	}

	d.Driver = "sqlite"

	if err := s.createTables(d); err != nil {
		return err
	}

	return s.migrateTables(d)
}

func (s *sqliteDB) createSettingTable(d *utils.DatabaseAccessor) error {
	sql := `CREATE TABLE "settings" (
	    "id" TEXT PRIMARY KEY NOT NULL,
	    "value" TEXT DEFAULT ''
	)`

	if _, err := d.DB.Exec(sql); err != nil {
		return err
	}

	_, err := d.DB.Exec(`INSERT INTO "settings" ("id", "value") VALUES ('db_version', ?)`, dbVersion)
	return err
}
