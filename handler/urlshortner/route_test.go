package urlshortner_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dev-AustinPeter/url-shortner-go/constants"
	"github.com/Dev-AustinPeter/url-shortner-go/handler/urlshortner"
	"github.com/Dev-AustinPeter/url-shortner-go/services/cachemanager"
	mocks "github.com/Dev-AustinPeter/url-shortner-go/tests/mock"
	"github.com/Dev-AustinPeter/url-shortner-go/types"
	"github.com/go-redis/redismock/v9"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGetTaskBaseOnTaskId_CacheHit(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/task/123", nil)
	req = mux.SetURLVars(req, map[string]string{"taskId": "123"})
	rec := httptest.NewRecorder()

	task := types.Task{
		TaskID:    "123",
		Status:    "Test Task",
		Result:    json.RawMessage{},
		CreatedAt: time.Now(),
	}

	// Simulating cache hit
	taskJson, _ := json.Marshal(task)
	// Mock Redis
	mockRedis.On("Get", req.Context(), "123").Return(string(taskJson), nil)

	handler.GetTaskBaseOnTaskId(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockRedis.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetTask")
}

func TestGetTaskBaseOnTaskId_CacheHitUsingClientMock(t *testing.T) {
	logger := zerolog.Nop()

	// Initialize the CacheManager
	redisClientMock, mockRedis := redismock.NewClientMock()
	mockCache := cachemanager.NewCacheManager(redisClientMock, logger)
	mockRepo := new(mocks.MockUrlRepository)

	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	task := types.Task{
		TaskID:    "123",
		Status:    "Test Task",
		Result:    json.RawMessage{},
		CreatedAt: time.Now(),
	}

	// Simulating cache hit
	taskJson, _ := json.Marshal(task)

	// Mock Redis
	mockRedis.ExpectGet("123").SetVal(string(taskJson))

	req, _ := http.NewRequest("GET", "/task/123", nil)
	req = mux.SetURLVars(req, map[string]string{"taskId": "123"})
	rec := httptest.NewRecorder()

	handler.GetTaskBaseOnTaskId(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertNotCalled(t, "GetTask")
}

func TestGetTaskBaseOnTaskId_CacheMiss_DBHit(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/task/123", nil)
	req = mux.SetURLVars(req, map[string]string{"taskId": "123"})
	rec := httptest.NewRecorder()

	task := types.Task{
		TaskID:    "123",
		Status:    "Test Task",
		Result:    json.RawMessage{},
		CreatedAt: time.Now(),
	}

	// Simulating cache miss
	mockRedis.On("Get", req.Context(), "123").Return("", redis.Nil)
	taskJson, _ := json.Marshal(task)
	mockRedis.On("Set", req.Context(), "123", string(taskJson), time.Minute*constants.CACHE_TTL_PERMANENT).Return(nil)

	// Mock Repository
	mockRepo.On("GetTask", "123").Return(task, nil)

	handler.GetTaskBaseOnTaskId(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetTaskBaseOnTaskId_CacheMiss_DBMiss(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/task/123", nil)
	req = mux.SetURLVars(req, map[string]string{"taskId": "123"})
	rec := httptest.NewRecorder()

	// Simulating cache miss
	mockRedis.On("Get", req.Context(), "123").Return("", redis.Nil)
	mockRepo.On("GetTask", "123").Return(types.Task{}, errors.New("task not found"))

	handler.GetTaskBaseOnTaskId(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockRedis.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
