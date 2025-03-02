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

// func (s *SqlHandler) Close() error {
// 	return s.Con.Close()
// }

// func (s *SqlHandler) Ping() error {
// 	return s.Con.Ping()
// }

func (s *SqlHandler) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.DB.QueryRow(query, args...)
}

func (s *SqlHandler) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.DB.Query(query, args...)
}

func (s *SqlHandler) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.DB.Exec(query, args...)
}

func (s *SqlHandler) Close() error {
	return s.DB.Close()
}

func (s *SqlHandler) Ping() error {
	return s.DB.Ping()
}
