package main

import (
	"gorm.io/driver/sqlite" // Sqlite driver based on GGO
	// "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
)

func openDB(connURL string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(connURL), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt: true,
	})
}
