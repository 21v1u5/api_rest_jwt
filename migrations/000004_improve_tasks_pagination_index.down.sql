DROP INDEX IF EXISTS idx_tasks_user_created_id;

CREATE INDEX idx_tasks_user_id_created_at ON tasks (user_id, created_at DESC);