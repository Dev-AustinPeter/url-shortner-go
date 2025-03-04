package cachemanager

import (
	"context"
	"errors"
	"testing"
	"time"

	mocks "github.com/Dev-AustinPeter/url-shortner-go/tests/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestCacheManager_Set_Success(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	cm := NewCacheManager(mockRedis, logger)
	ctx := context.Background()

	mockRedis.On("Set", ctx, "testKey", "testValue", 5*time.Minute).Return(nil)

	err := cm.Set(ctx, "testKey", "testValue", 5)

	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestCacheManager_Set_Error(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	cm := NewCacheManager(mockRedis, logger)
	ctx := context.Background()

	mockRedis.On("Set", ctx, "testKey", "testValue", 5*time.Minute).Return(errors.New("redis error"))

	err := cm.Set(ctx, "testKey", "testValue", 5)

	assert.Error(t, err)
	mockRedis.AssertExpectations(t)
}

func TestCacheManager_Get_Success(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	cm := NewCacheManager(mockRedis, logger)
	ctx := context.Background()

	mockRedis.On("Get", ctx, "testKey").Return("testValue", nil)

	val, err := cm.Get(ctx, "testKey")

	assert.NoError(t, err)
	assert.Equal(t, "testValue", val)
	mockRedis.AssertExpectations(t)
}

func TestCacheManager_Get_Error(t *testing.T) {
	mockRedis := new(mocks.MockRedisClient)
	logger := zerolog.Nop()

	cm := NewCacheManager(mockRedis, logger)
	ctx := context.Background()

	mockRedis.On("Get", ctx, "testKey").Return("", errors.New("key not found"))

	val, err := cm.Get(ctx, "testKey")

	assert.Error(t, err)
	assert.Empty(t, val)
	mockRedis.AssertExpectations(t)
}
