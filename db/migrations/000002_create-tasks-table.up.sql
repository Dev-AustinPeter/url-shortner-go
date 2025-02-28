CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    task_id UUID UNIQUE NOT NULL,  -- Unique task ID
    status VARCHAR(20) DEFAULT 'pending',  -- Task status: pending, processing, completed, failed
    result JSONB DEFAULT NULL,  -- Stores the processed URLs (JSON format)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_id ON tasks(task_id);
CREATE INDEX idx_status ON tasks(status);
CREATE INDEX idx_created_at_task ON tasks(created_at);