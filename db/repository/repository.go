package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Dev-AustinPeter/url-shortner-go/constants"
	"github.com/Dev-AustinPeter/url-shortner-go/db"
	"github.com/Dev-AustinPeter/url-shortner-go/types"
	"github.com/gofrs/uuid"
	"golang.org/toolchain/src/math/rand"
)

type Url struct {
	ID        sql.NullInt64  `json:"id"`
	ShortCode sql.NullString `json:"shortCode"`
	LongUrl   sql.NullString `json:"longUrl"`
	CreatedAt sql.NullTime   `json:"createdAt"`
}

type UrlRepository interface {
	CreateUrl(url string) (*string, error)
	GetUrl(shortCode string) (Url, error)
	GetLongUrl(longUrl string) (Url, error)
	CreateTaskId() (*types.Task, error)
	GetTask(taskId string) (types.Task, error)
	UpdateTask(taskId string, status string, result json.RawMessage) error
	GetAllUrls() ([]Url, error)
}

type Repository struct {
	DB db.Database
}

func NewRepository(con db.Database) UrlRepository {
	return &Repository{
		DB: con,
	}
}

func GenerateShortCode(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = constants.LETTER_BYTES[rand.Intn(len(constants.LETTER_BYTES))]
	}
	return string(b)
}

func (r *Repository) CreateUrl(LongUrl string) (*string, error) {
	url, err := r.GetLongUrl(LongUrl)
	if err == nil && url.LongUrl.String == LongUrl {
		return &url.ShortCode.String, nil
	}

	shortCode := GenerateShortCode(6)

	tn := time.Now().UTC()
	_, err = r.DB.Exec("INSERT INTO urls (short_code, long_url, created_at) VALUES ($1, $2, $3)", shortCode, LongUrl, tn)
	if err != nil {
		return nil, err
	}
	return &shortCode, nil
}

func (r *Repository) GetUrl(shortCode string) (Url, error) {
	var url Url
	err := r.DB.QueryRow("SELECT id, short_code, long_url, created_at FROM urls WHERE short_code = $1", shortCode).Scan(&url.ID, &url.ShortCode, &url.LongUrl, &url.CreatedAt)
	if err != nil {
		return Url{}, err
	}
	return url, nil
}

func (r *Repository) GetLongUrl(longUrl string) (Url, error) {
	var url Url
	err := r.DB.QueryRow("SELECT id, short_code, long_url, created_at FROM urls WHERE long_url = $1", longUrl).
		Scan(&url.ID, &url.ShortCode, &url.LongUrl, &url.CreatedAt)
	if err != nil {
		return Url{}, err
	}
	return url, nil

}

func (r *Repository) GetAllUrls() ([]Url, error) {
	rows, err := r.DB.Query("SELECT id, short_code, long_url, created_at FROM urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var urls []Url
	for rows.Next() {
		var url Url
		err := rows.Scan(&url.ID, &url.ShortCode, &url.LongUrl, &url.CreatedAt)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}

	return urls, nil
}

func (r *Repository) CreateTaskId() (*types.Task, error) {
	taskId := uuid.Must(uuid.NewV4()).String()
	stm, err := r.DB.Prepare("INSERT INTO tasks (task_id, status, created_at) VALUES ($1, $2, $3)") // status is default to pending
	if err != nil {
		return nil, err
	}
	defer stm.Close()
	tn := time.Now().UTC()
	_, err = stm.Exec(taskId, "pending", tn)
	if err != nil {
		return nil, err
	}
	return &types.Task{TaskID: taskId, Status: "pending", CreatedAt: tn}, nil
}

func (r *Repository) UpdateTask(taskId string, status string, result json.RawMessage) error {
	stm, err := r.DB.Prepare("UPDATE tasks SET status = $1, result = CASE WHEN $2::text = '' THEN NULL ELSE $2::jsonb END WHERE task_id = $3")
	if err != nil {
		return err
	}
	defer stm.Close()
	_, err = stm.Exec(status, result, taskId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetTask(taskId string) (types.Task, error) {
	var task types.Task
	err := r.DB.QueryRow("SELECT task_id, status, result, created_at FROM tasks WHERE task_id = $1", taskId).Scan(&task.TaskID, &task.Status, &task.Result, &task.CreatedAt)
	if err != nil {
		return types.Task{}, err
	}
	return task, nil
}
