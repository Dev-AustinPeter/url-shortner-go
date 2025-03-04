package mocks

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
)

type MockDB struct {
	mock sqlmock.Sqlmock
	db   *sql.DB
}

func (m *MockDB) Prepare(query string) (*sql.Stmt, error) {
	return m.db.Prepare(query)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(query, args...)
}

func (m *MockDB) Close() error {
	return m.db.Close()
}

func (m *MockDB) Ping() error {
	return nil
}

func NewMockDB() (*MockDB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	return &MockDB{
		mock: mock,
		db:   db,
	}, mock
}
