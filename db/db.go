package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type SqlHandler struct {
	*sql.DB
}

type Database interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Close() error
	Ping() error
}

func NewConnection(hostname string, port string, username string, password string, dbname string, driver string) *SqlHandler {

	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password='' sslmode=disable", hostname, port, username, dbname)
	fmt.Println(dataSourceName)
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil
	}
	return &SqlHandler{DB: db}
}
