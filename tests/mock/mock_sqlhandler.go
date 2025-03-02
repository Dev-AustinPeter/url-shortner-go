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

/* type MockUrlRepository struct {
	*MockSqlHandler
}

func (m *MockUrlRepository) CreateUrl(url string) (*string, error) {
	args := m.Called(url)
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockUrlRepository) GetUrl(shortCode string) (repository.Url, error) {
	args := m.Called(shortCode)
	return args.Get(0).(repository.Url), args.Error(1)
}

func (m *MockUrlRepository) CreateTaskId() (*types.Task, error) {
	args := m.Called()
	return args.Get(0).(*types.Task), args.Error(1)
}

func (m *MockUrlRepository) GetTask(taskId string) (types.Task, error) {
	args := m.Called(taskId)
	return args.Get(0).(types.Task), args.Error(1)
}

func (m *MockUrlRepository) UpdateTask(taskId string, status string, result json.RawMessage) error {
	args := m.Called(taskId, status, result)
	return args.Error(0)
}

func (m *MockUrlRepository) GetAllUrls() ([]repository.Url, error) {
	args := m.Called()
	return args.Get(0).([]repository.Url), args.Error(1)
}
*/
