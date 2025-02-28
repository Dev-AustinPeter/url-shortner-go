package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type SqlHandler struct {
	Con *sql.DB
}

func NewConnection(hostname string, port string, username string, password string, dbname string, driver string) *SqlHandler {

	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password='' sslmode=disable", hostname, port, username, dbname)
	fmt.Println(dataSourceName)
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil
	}
	return &SqlHandler{Con: db}
}

func (s *SqlHandler) Close() error {
	return s.Con.Close()
}

func (s *SqlHandler) Ping() error {
	return s.Con.Ping()
}
