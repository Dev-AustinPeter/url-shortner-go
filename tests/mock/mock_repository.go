package mocks

import (
	"encoding/json"

	"github.com/Dev-AustinPeter/url-shortner-go/db/repository"
	"github.com/Dev-AustinPeter/url-shortner-go/types"
	"github.com/stretchr/testify/mock"
)

type MockUrlRepository struct {
	mock.Mock
}

var _ repository.UrlRepository = (*MockUrlRepository)(nil)

func (m *MockUrlRepository) GetTask(taskId string) (types.Task, error) {
	args := m.Called(taskId)
	return args.Get(0).(types.Task), args.Error(1)
}

func (m *MockUrlRepository) UpdateTask(taskId string, status string, result json.RawMessage) error {
	args := m.Called(taskId, status, result)
	return args.Error(0)
}

func (m *MockUrlRepository) CreateTaskId() (*types.Task, error) {
	args := m.Called()
	return args.Get(0).(*types.Task), args.Error(1)
}

func (m *MockUrlRepository) GetUrl(shortCode string) (repository.Url, error) {
	args := m.Called(shortCode)
	return args.Get(0).(repository.Url), args.Error(1)
}

func (m *MockUrlRepository) CreateUrl(url string) (*string, error) {
	args := m.Called(url)
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockUrlRepository) GetAllUrls() ([]repository.Url, error) {
	args := m.Called()
	return args.Get(0).([]repository.Url), args.Error(1)
}

func (m *MockUrlRepository) GetLongUrl(longUrl string) (repository.Url, error) {
	args := m.Called(longUrl)
	return args.Get(0).(repository.Url), args.Error(1)
}
