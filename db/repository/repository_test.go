package repository_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Dev-AustinPeter/url-shortner-go/db/repository"
	mocks "github.com/Dev-AustinPeter/url-shortner-go/tests/mock"
	"github.com/Dev-AustinPeter/url-shortner-go/types"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetLongUrl(t *testing.T) {
	mockDB, mock := mocks.NewMockDB()
	defer mockDB.Close()

	repo := repository.NewRepository(mockDB)

	// Test successful query
	t.Run("Success", func(t *testing.T) {
		longUrl := "https://example.com/long-url"
		expectedUrl := repository.Url{
			ID:        sql.NullInt64{Int64: 1, Valid: true},
			ShortCode: sql.NullString{String: "abc123", Valid: true},
			LongUrl:   sql.NullString{String: longUrl, Valid: true},
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		rows := sqlmock.NewRows([]string{"id", "short_code", "long_url", "created_at"}).
			AddRow(expectedUrl.ID.Int64, expectedUrl.ShortCode.String, expectedUrl.LongUrl.String, expectedUrl.CreatedAt.Time)

		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = \\$1").
			WithArgs(longUrl).
			WillReturnRows(rows)

		url, err := repo.GetLongUrl(longUrl)

		assert.NoError(t, err)
		assert.Equal(t, expectedUrl.ID.Int64, url.ID.Int64)
		assert.Equal(t, expectedUrl.ShortCode.String, url.ShortCode.String)
		assert.Equal(t, expectedUrl.LongUrl.String, url.LongUrl.String)
	})

	// Test when URL not found
	t.Run("Not Found", func(t *testing.T) {
		longUrl := "https://example.com/non-existent"

		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = \\$1").
			WithArgs(longUrl).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetLongUrl(longUrl)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	// Test database error
	t.Run("Database Error", func(t *testing.T) {
		longUrl := "https://example.com/error-url"
		dbErr := errors.New("database connection error")

		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = \\$1").
			WithArgs(longUrl).
			WillReturnError(dbErr)

		_, err := repo.GetLongUrl(longUrl)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
	})
}

func TestCreateUrl(t *testing.T) {
	mockDB, mock := mocks.NewMockDB()
	defer mockDB.Close()

	repo := repository.NewRepository(mockDB)

	t.Run("URL Already Exists", func(t *testing.T) {
		// Setup
		longUrl := "https://example.com/existing-url"
		existingShortCode := "abc123"

		// Mock GetLongUrl query - simulate URL already exists
		rows := sqlmock.NewRows([]string{"id", "short_code", "long_url", "created_at"}).
			AddRow(1, existingShortCode, longUrl, time.Now())

		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = \\$1").
			WithArgs(longUrl).
			WillReturnRows(rows)

		// Call the function
		shortCode, err := repo.CreateUrl(longUrl)

		// Assert the results
		assert.NoError(t, err)
		assert.Equal(t, existingShortCode, *shortCode)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create New URL", func(t *testing.T) {
		// Setup
		longUrl := "https://example.com/new-url"

		// Mock GetLongUrl query - simulate URL doesn't exist yet
		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = \\$1").
			WithArgs(longUrl).
			WillReturnError(sql.ErrNoRows)

		// Mock the INSERT query with AnyArg for the short code
		mock.ExpectExec("INSERT INTO urls \\(short_code, long_url, created_at\\) VALUES \\(\\$1, \\$2, \\$3\\)").
			WithArgs(sqlmock.AnyArg(), longUrl, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected

		// Call the function
		shortCode, err := repo.CreateUrl(longUrl)

		// Assert the results
		assert.NoError(t, err)
		assert.NotNil(t, shortCode)
		assert.Len(t, *shortCode, 6) // Assuming GenerateShortCode(6) creates a 6-char code

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Insert Error", func(t *testing.T) {
		// Setup
		longUrl := "https://example.com/error-url"
		dbErr := errors.New("insert error")

		// Mock GetLongUrl query - simulate URL doesn't exist yet
		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = \\$1").
			WithArgs(longUrl).
			WillReturnError(sql.ErrNoRows)

		// Mock the INSERT query with an error
		mock.ExpectExec("INSERT INTO urls \\(short_code, long_url, created_at\\) VALUES \\(\\$1, \\$2, \\$3\\)").
			WithArgs(sqlmock.AnyArg(), longUrl, sqlmock.AnyArg()).
			WillReturnError(dbErr)

		// Call the function
		shortCode, err := repo.CreateUrl(longUrl)

		// Assert the results
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, shortCode)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetUrl(t *testing.T) {
	mockDB, mock := mocks.NewMockDB()
	defer mockDB.Close()

	repo := repository.NewRepository(mockDB)

	// Test successful query
	t.Run("Success", func(t *testing.T) {
		longUrl := "https://example.com/long-url"
		shortCode := "abc123"
		expectedUrl := repository.Url{
			ID:        sql.NullInt64{Int64: 1, Valid: true},
			ShortCode: sql.NullString{String: "abc123", Valid: true},
			LongUrl:   sql.NullString{String: longUrl, Valid: true},
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		rows := sqlmock.NewRows([]string{"id", "short_code", "long_url", "created_at"}).
			AddRow(expectedUrl.ID.Int64, expectedUrl.ShortCode.String, expectedUrl.LongUrl.String, expectedUrl.CreatedAt.Time)

		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE short_code = \\$1").
			WithArgs(shortCode).
			WillReturnRows(rows)

		url, err := repo.GetUrl(shortCode)

		assert.NoError(t, err)
		assert.Equal(t, expectedUrl.ID.Int64, url.ID.Int64)
		assert.Equal(t, expectedUrl.ShortCode.String, url.ShortCode.String)
		assert.Equal(t, expectedUrl.LongUrl.String, url.LongUrl.String)
	})

	// Test when URL not found
	t.Run("Not Found", func(t *testing.T) {
		shortCode := "abc123"
		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE short_code = \\$1").
			WithArgs(shortCode).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetUrl(shortCode)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	// Test database error
	t.Run("Database Error", func(t *testing.T) {
		shortCode := "abc123"
		dbErr := errors.New("database connection error")

		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls WHERE short_code = \\$1").
			WithArgs(shortCode).
			WillReturnError(dbErr)

		_, err := repo.GetUrl(shortCode)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
	})
}

func TestCreateTaskId(t *testing.T) {
	mockDB, mock := mocks.NewMockDB()
	defer mockDB.Close()

	repo := repository.NewRepository(mockDB)

	t.Run("Success", func(t *testing.T) {
		// Mock the Prepare statement
		mockStmt := mock.ExpectPrepare("INSERT INTO tasks \\(task_id, status, created_at\\) VALUES \\(\\$1, \\$2, \\$3\\)")

		// Mock the Exec on the prepared statement
		mockStmt.ExpectExec().
			WithArgs(sqlmock.AnyArg(), "pending", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Call the function
		task, err := repo.CreateTaskId()

		// Assert the results
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, "pending", task.Status)
		assert.NotEmpty(t, task.TaskID)

		// Verify UUID format (assuming it's a UUID v4)
		_, uuidErr := uuid.FromString(task.TaskID)
		assert.NoError(t, uuidErr)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Prepare Error", func(t *testing.T) {
		// Mock a prepare error
		prepareErr := errors.New("prepare statement error")
		mock.ExpectPrepare("INSERT INTO tasks \\(task_id, status, created_at\\) VALUES \\(\\$1, \\$2, \\$3\\)").
			WillReturnError(prepareErr)

		// Call the function
		task, err := repo.CreateTaskId()

		// Assert the results
		assert.Error(t, err)
		assert.Equal(t, prepareErr, err)
		assert.Nil(t, task)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Exec Error", func(t *testing.T) {
		// Mock the Prepare statement
		execErr := errors.New("execution error")
		mockStmt := mock.ExpectPrepare("INSERT INTO tasks \\(task_id, status, created_at\\) VALUES \\(\\$1, \\$2, \\$3\\)")

		// Mock an exec error
		mockStmt.ExpectExec().
			WithArgs(sqlmock.AnyArg(), "pending", sqlmock.AnyArg()).
			WillReturnError(execErr)

		// Call the function
		task, err := repo.CreateTaskId()

		// Assert the results
		assert.Error(t, err)
		assert.Equal(t, execErr, err)
		assert.Nil(t, task)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateTask(t *testing.T) {
	mockDB, mock := mocks.NewMockDB()
	defer mockDB.Close()

	repo := repository.NewRepository(mockDB)

	t.Run("Success With Result", func(t *testing.T) {
		// Setup
		taskId := "abc123-uuid"
		status := "completed"
		result := json.RawMessage(`{"url": "https://short.url/abc123"}`)

		// Mock the Prepare statement
		mockStmt := mock.ExpectPrepare("UPDATE tasks SET status = \\$1, result = CASE WHEN \\$2::text = '' THEN NULL ELSE \\$2::jsonb END WHERE task_id = \\$3")

		// Mock the Exec on the prepared statement
		mockStmt.ExpectExec().
			WithArgs(status, result, taskId).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 0 last insert id, 1 row affected

		// Call the function
		err := repo.UpdateTask(taskId, status, result)

		// Assert the results
		assert.NoError(t, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success With Empty Result", func(t *testing.T) {
		// Setup
		taskId := "abc123-uuid"
		status := "failed"
		var result json.RawMessage // empty result

		// Mock the Prepare statement
		mockStmt := mock.ExpectPrepare("UPDATE tasks SET status = \\$1, result = CASE WHEN \\$2::text = '' THEN NULL ELSE \\$2::jsonb END WHERE task_id = \\$3")

		// Mock the Exec on the prepared statement
		mockStmt.ExpectExec().
			WithArgs(status, result, taskId).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Call the function
		err := repo.UpdateTask(taskId, status, result)

		// Assert the results
		assert.NoError(t, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Prepare Error", func(t *testing.T) {
		// Setup
		taskId := "abc123-uuid"
		status := "completed"
		result := json.RawMessage(`{"url": "https://short.url/abc123"}`)
		prepareErr := errors.New("prepare statement error")

		// Mock a prepare error
		mock.ExpectPrepare("UPDATE tasks SET status = \\$1, result = CASE WHEN \\$2::text = '' THEN NULL ELSE \\$2::jsonb END WHERE task_id = \\$3").
			WillReturnError(prepareErr)

		// Call the function
		err := repo.UpdateTask(taskId, status, result)

		// Assert the results
		assert.Error(t, err)
		assert.Equal(t, prepareErr, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Exec Error", func(t *testing.T) {
		// Setup
		taskId := "abc123-uuid"
		status := "completed"
		result := json.RawMessage(`{"url": "https://short.url/abc123"}`)
		execErr := errors.New("execution error")

		// Mock the Prepare statement
		mockStmt := mock.ExpectPrepare("UPDATE tasks SET status = \\$1, result = CASE WHEN \\$2::text = '' THEN NULL ELSE \\$2::jsonb END WHERE task_id = \\$3")

		// Mock an exec error
		mockStmt.ExpectExec().
			WithArgs(status, result, taskId).
			WillReturnError(execErr)

		// Call the function
		err := repo.UpdateTask(taskId, status, result)

		// Assert the results
		assert.Error(t, err)
		assert.Equal(t, execErr, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetTask(t *testing.T) {
	mockDB, mock := mocks.NewMockDB()
	defer mockDB.Close()

	repo := repository.NewRepository(mockDB)

	t.Run("Success With Result", func(t *testing.T) {
		// Setup
		taskId := "abc123-uuid"
		status := "completed"
		result := json.RawMessage(`{"url": "https://short.url/abc123"}`)
		createdAt := time.Now().UTC()

		// Create expected task
		expectedTask := types.Task{
			TaskID:    taskId,
			Status:    status,
			Result:    result,
			CreatedAt: createdAt,
		}

		// Mock the query
		rows := sqlmock.NewRows([]string{"task_id", "status", "result", "created_at"}).
			AddRow(expectedTask.TaskID, expectedTask.Status, expectedTask.Result, expectedTask.CreatedAt)

		mock.ExpectQuery("SELECT task_id, status, result, created_at FROM tasks WHERE task_id = \\$1").
			WithArgs(taskId).
			WillReturnRows(rows)

		// Call the function
		task, err := repo.GetTask(taskId)

		// Assert the results
		assert.NoError(t, err)
		assert.Equal(t, expectedTask.TaskID, task.TaskID)
		assert.Equal(t, expectedTask.Status, task.Status)
		assert.Equal(t, string(expectedTask.Result), string(task.Result))
		assert.WithinDuration(t, expectedTask.CreatedAt, task.CreatedAt, time.Millisecond)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success With Null Result", func(t *testing.T) {
		// Setup
		taskId := "abc123-uuid"
		status := "pending"
		var result json.RawMessage = nil // null result
		createdAt := time.Now().UTC()

		// Create expected task
		expectedTask := types.Task{
			TaskID:    taskId,
			Status:    status,
			Result:    result,
			CreatedAt: createdAt,
		}

		// Mock the query
		rows := sqlmock.NewRows([]string{"task_id", "status", "result", "created_at"}).
			AddRow(expectedTask.TaskID, expectedTask.Status, []byte(nil), expectedTask.CreatedAt)

		mock.ExpectQuery("SELECT task_id, status, result, created_at FROM tasks WHERE task_id = \\$1").
			WithArgs(taskId).
			WillReturnRows(rows)

		// Call the function
		task, err := repo.GetTask(taskId)

		// Assert the results
		assert.NoError(t, err)
		assert.Equal(t, expectedTask.TaskID, task.TaskID)
		assert.Equal(t, expectedTask.Status, task.Status)
		assert.Nil(t, task.Result)
		assert.WithinDuration(t, expectedTask.CreatedAt, task.CreatedAt, time.Millisecond)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Task Not Found", func(t *testing.T) {
		// Setup
		taskId := "nonexistent-uuid"

		// Mock the query with no rows
		mock.ExpectQuery("SELECT task_id, status, result, created_at FROM tasks WHERE task_id = \\$1").
			WithArgs(taskId).
			WillReturnError(sql.ErrNoRows)

		// Call the function
		task, err := repo.GetTask(taskId)

		// Assert the results
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Equal(t, types.Task{}, task)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		// Setup
		taskId := "abc123-uuid"
		dbErr := errors.New("database error")

		// Mock the query with an error
		mock.ExpectQuery("SELECT task_id, status, result, created_at FROM tasks WHERE task_id = \\$1").
			WithArgs(taskId).
			WillReturnError(dbErr)

		// Call the function
		task, err := repo.GetTask(taskId)

		// Assert the results
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, types.Task{}, task)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestGetAllUrls(t *testing.T) {
	mockDB, mock := mocks.NewMockDB()
	defer mockDB.Close()

	repo := repository.NewRepository(mockDB)

	// Test successful query
	t.Run("Success", func(t *testing.T) {
		longUrl := "https://example.com/long-url"
		expectedUrl := repository.Url{
			ID:        sql.NullInt64{Int64: 1, Valid: true},
			ShortCode: sql.NullString{String: "abc123", Valid: true},
			LongUrl:   sql.NullString{String: longUrl, Valid: true},
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		longUrl2 := "https://example.com/long-url2"

		expectedUrl2 := repository.Url{
			ID:        sql.NullInt64{Int64: 2, Valid: true},
			ShortCode: sql.NullString{String: "abc456", Valid: true},
			LongUrl:   sql.NullString{String: longUrl2, Valid: true},
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}

		rows := sqlmock.NewRows([]string{"id", "short_code", "long_url", "created_at"}).
			AddRow(expectedUrl.ID.Int64, expectedUrl.ShortCode.String, expectedUrl.LongUrl.String, expectedUrl.CreatedAt.Time).
			AddRow(expectedUrl2.ID.Int64, expectedUrl2.ShortCode.String, expectedUrl2.LongUrl.String, expectedUrl2.CreatedAt.Time)

		mock.ExpectQuery("SELECT id, short_code, long_url, created_at FROM urls").
			WillReturnRows(rows)

		url, err := repo.GetAllUrls()

		assert.NoError(t, err)
		assert.Equal(t, expectedUrl.ID.Int64, url[0].ID.Int64)
		assert.Equal(t, expectedUrl.ShortCode.String, url[0].ShortCode.String)
		assert.Equal(t, expectedUrl.LongUrl.String, url[0].LongUrl.String)
	})
}
