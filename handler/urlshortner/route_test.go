package urlshortner_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Dev-AustinPeter/url-shortner-go/constants"
	"github.com/Dev-AustinPeter/url-shortner-go/db/repository"
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

func TestShorten_EmptyPayload(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("POST", "/shorten", nil)
	rec := httptest.NewRecorder()

	handler.Shorten(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":"missing request body"}`, rec.Body.String())

}
func TestShorten_EmptyUrl(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	reqBody := `{"longUrl": ""}`

	req, _ := http.NewRequest("POST", "/shorten", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.Shorten(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":"LongUrl is required"}`, rec.Body.String())

}
func TestShorten_CreateUrl_Error(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	reqBody := `{"longUrl": "http://google.com"}`

	req, _ := http.NewRequest("POST", "/shorten", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()

	emptyString := ""
	mockRepo.On("CreateUrl", "http://google.com").Return(&emptyString, errors.New("error creating url"))

	handler.Shorten(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":"error creating url"}`, rec.Body.String())

}

func TestShorten_CreateUrl_Success(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	reqBody := `{"longUrl": "http://google.com"}`

	req, _ := http.NewRequest("POST", "/shorten", strings.NewReader(reqBody))
	rec := httptest.NewRecorder()

	shortString := "abc123"
	mockRepo.On("CreateUrl", "http://google.com").Return(&shortString, nil)

	handler.Shorten(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{"shortCode":"abc123","longUrl":"http://google.com"}`, rec.Body.String())

}

func TestGetShorten_emptyShortCode(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/shorten", nil)
	rec := httptest.NewRecorder()

	handler.GetShorten(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":"ShortUrl is required"}`, rec.Body.String())
}

func TestGetShorten_Error(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/shorten/abc123", nil)
	req = mux.SetURLVars(req, map[string]string{"shortUrl": "abc123"})
	rec := httptest.NewRecorder()

	mockRepo.On("GetUrl", "abc123").Return(repository.Url{}, errors.New("error getting url"))

	handler.GetShorten(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"error":"ShortUrl not found"}`, rec.Body.String())
}

func TestGetShorten_Success(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/shorten/abc123", nil)
	req = mux.SetURLVars(req, map[string]string{"shortUrl": "abc123"})
	rec := httptest.NewRecorder()

	tN := time.Now()

	mockRepo.On("GetUrl", "abc123").Return(repository.Url{
		ShortCode: sql.NullString{String: "abc123", Valid: true},
		LongUrl:   sql.NullString{String: "http://google.com", Valid: true},
		CreatedAt: sql.NullTime{Time: tN, Valid: true},
	}, nil)

	expected := `{"shortCode":"abc123","longUrl":"http://google.com","createdAt":"` + tN.UTC().String() + `"}`

	handler.GetShorten(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, expected, rec.Body.String())
}

func TestCreateTaskId_Error(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/shorten", nil)
	rec := httptest.NewRecorder()

	mockRepo.On("CreateTaskId").Return(&types.Task{}, errors.New("error creating task id"))

	handler.CreateTaskId(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":"error creating task id"}`, rec.Body.String())

}

func TestCreateTaskId_Success(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	// Initialize the CacheManager
	mockCache := cachemanager.NewCacheManager(mockRedis, logger)
	// Initialize the MockUrlRepository
	mockRepo := new(mocks.MockUrlRepository)

	// Initialize the Handler
	handler := urlshortner.NewHandler(mockRepo, &logger, mockCache)

	req, _ := http.NewRequest("GET", "/shorten", nil)
	rec := httptest.NewRecorder()

	tN := time.Now()
	task := &types.Task{
		TaskID:    "123",
		Status:    "pending",
		CreatedAt: tN.UTC(),
	}

	jsonMsg := json.RawMessage(nil)
	mockRepo.On("UpdateTask", "123", "processing", jsonMsg).Return(nil)
	mockRepo.On("GetAllUrls").Return([]repository.Url{
		{
			ShortCode: sql.NullString{String: "abc123", Valid: true},
			LongUrl:   sql.NullString{String: "http://google.com", Valid: true},
		},
	}, nil)
	mockRepo.On("UpdateTask", "123", "completed", json.RawMessage(`[{"long":"http://google.com","short":"abc123"}]`)).Return(nil)

	mockRepo.On("CreateTaskId").Return(task, nil)

	handler.CreateTaskId(rec, req)
	time.Sleep(time.Second * 2)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{"created_at":"`+tN.UTC().Format(time.RFC3339Nano)+`","task_id":"123","status":"pending"}`, rec.Body.String())

}
