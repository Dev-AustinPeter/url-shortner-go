package urlshortner

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dev-AustinPeter/url-shortner-go/constants"
	"github.com/Dev-AustinPeter/url-shortner-go/db/repository"
	"github.com/Dev-AustinPeter/url-shortner-go/middleware"
	"github.com/Dev-AustinPeter/url-shortner-go/services/cachemanager"
	"github.com/Dev-AustinPeter/url-shortner-go/types"
	"github.com/Dev-AustinPeter/url-shortner-go/utils"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Handler struct {
	UrlRepository repository.UrlRepository
	Logger        *zerolog.Logger
	CacheManager  *cachemanager.CacheManager
}

func NewHandler(repository repository.UrlRepository, logger *zerolog.Logger, cacheManager *cachemanager.CacheManager) *Handler {
	return &Handler{
		UrlRepository: repository,
		Logger:        logger,
		CacheManager:  cacheManager,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router, middleware *middleware.RateLimiter) {

	r.Handle("/shorten", middleware.Limit(http.HandlerFunc(h.shorten))).Methods("POST")
	r.Handle("/shorten/{shortUrl}", middleware.Limit(http.HandlerFunc(h.getShorten))).Methods("GET")
	r.Handle("/shorten", middleware.Limit(http.HandlerFunc(h.createTaskId))).Methods("GET")
	r.Handle("/task/{taskId}", middleware.Limit(http.HandlerFunc(h.getTaskBaseOnTaskId))).Methods("GET")
}

// shorten handles POST requests to /shorten. It takes a JSON payload with a
// "longUrl" field and returns a JSON response with a "shortCode" field.
// If the payload is invalid, it returns a 400 error. If the URL cannot be
// shortened, it returns a 500 error. Otherwise, it returns a 201 Created
// status with the shortened URL in the response body.
func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		LongUrl string `json:"longUrl"`
	}

	if err := utils.ParseJson(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if payload.LongUrl == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s", "LongUrl is required"))
		return
	}

	sUrl, err := h.UrlRepository.CreateUrl(payload.LongUrl)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJson(w, http.StatusCreated, types.ResponseUrl{
		ShortCode: *sUrl,
		LongUrl:   payload.LongUrl,
	})

}

// getShorten handles GET requests to /shorten/{shortUrl}. It attempts to fetch the URL from the database.
// If the URL is not found, it returns a 404 error. Otherwise, it returns the URL in the response body.
func (h *Handler) getShorten(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortUrl := vars["shortUrl"]

	if shortUrl == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s", "ShortUrl is required"))
		return
	}

	url, err := h.UrlRepository.GetUrl(shortUrl)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("%s", "ShortUrl not found"))
		return
	}

	utils.WriteJson(w, http.StatusOK, types.ResponseUrl{
		ShortCode: url.ShortCode.String,
		LongUrl:   url.LongUrl.String,
		CreatedAt: url.CreatedAt.Time.UTC().String(),
	})
}

// createTaskId handles GET requests to /task. It creates a new task in the database and starts it in a separate goroutine.
// The task is responsible for processing all URLs in the database and storing the result in the task's result field.
// If the task creation fails, it returns a 500 error. Otherwise, it returns the created task in the response body.
func (h *Handler) createTaskId(w http.ResponseWriter, r *http.Request) {
	task, err := h.UrlRepository.CreateTaskId()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	/**
	 * By running it in a separate goroutine, the server can continue to handle other requests while the task is being processed.
	 * in production, you may want to consider using a task queue or a background job processing system to handle long-running tasks.
	 */
	go h.processTask(task.TaskID)

	utils.WriteJson(w, http.StatusCreated, types.Task{
		TaskID:    task.TaskID,
		Status:    task.Status,
		CreatedAt: task.CreatedAt.UTC(),
	})
}

// getTaskBaseOnTaskId handles GET requests to /task/{taskId}. It attempts to fetch the task from the cache first. If the cache
// is a miss, it fetches the task from the database and stores it in the cache for future requests. If the task is not found in
// the database, it returns a 404 error.
func (h *Handler) getTaskBaseOnTaskId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]

	if taskId == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s", "TaskId is required"))
		return
	}

	// Try fetching from cache first
	data, err := h.CacheManager.Get(r.Context(), taskId)
	if err != nil && err != redis.Nil {
		h.Logger.Error().Err(err).Str("task_id", taskId).Msg("Failed to fetch task from cache")
	}

	// If cache hit, return the task immediately
	if err == nil && data != "" {
		var task types.Task
		if json.Unmarshal([]byte(data), &task) == nil { // Avoid redundant error checking
			h.Logger.Info().Str("task_id", taskId).Msg("Task fetched from cache")
			task.CreatedAt = task.CreatedAt.UTC()
			utils.WriteJson(w, http.StatusOK, task)
			return
		}
		h.Logger.Error().Str("task_id", taskId).Msg("Failed to unmarshal task from cache")
	}

	// Fetch from database if cache miss
	task, err := h.UrlRepository.GetTask(taskId)
	if err != nil {
		h.Logger.Warn().Str("task_id", taskId).Msg("Task not found in database")
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("%s", "Task not found"))
		return
	}

	task.CreatedAt = task.CreatedAt.UTC()

	// Store task in cache for future requests
	if jsonTask, err := json.Marshal(task); err == nil {
		if err = h.CacheManager.Set(r.Context(), taskId, string(jsonTask), constants.CACHE_TTL_PERMANENT); err != nil {
			h.Logger.Error().Err(err).Str("task_id", taskId).Msg("Failed to set task in cache")
		}
	} else {
		h.Logger.Error().Err(err).Str("task_id", taskId).Msg("Failed to marshal task for caching")
	}

	utils.WriteJson(w, http.StatusOK, task)
}

// processTask processes a task by updating its status and fetching all URLs from the database.
// It first checks if the provided taskId is valid. If valid, it marks the task status as "processing".
// It then retrieves all URLs, and if none are found, it marks the task as "completed" with an empty result.
// Otherwise, it marshals the URLs into JSON format and updates the task status to "completed" with the result.
// In case of any error during the process, it updates the task status to "failed" and logs the error details.

func (h *Handler) processTask(taskId string) {
	if taskId == "" {
		h.Logger.Error().Msg("Invalid task ID")
		return
	}

	// Mark task as processing
	if err := h.UrlRepository.UpdateTask(taskId, "processing", nil); err != nil {
		h.Logger.Error().Err(err).Msg("Failed to update task status to processing")
		return
	}

	// Fetch all URLs
	urls, err := h.UrlRepository.GetAllUrls()
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to fetch URLs")
		h.UrlRepository.UpdateTask(taskId, "failed", nil)
		return
	}

	// Handle empty result case
	if len(urls) == 0 {
		h.Logger.Warn().Msg("No URLs found, marking task as completed with empty result")
		h.UrlRepository.UpdateTask(taskId, "completed", nil)
		return
	}

	// Pre-allocate slice for efficiency
	urlsMap := make([]map[string]string, len(urls))
	for i, url := range urls {
		urlsMap[i] = map[string]string{"short": url.ShortCode.String, "long": url.LongUrl.String}
	}

	// Convert to JSON
	result, err := json.Marshal(urlsMap)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to marshal URL data")
		h.UrlRepository.UpdateTask(taskId, "failed", nil)
		return
	}

	// Mark task as completed
	if err := h.UrlRepository.UpdateTask(taskId, "completed", result); err != nil {
		h.Logger.Error().Err(err).Msg("Failed to update task status to completed")
		return
	}

	h.Logger.Info().Str("task_id", taskId).Msg("Task completed successfully")
}
