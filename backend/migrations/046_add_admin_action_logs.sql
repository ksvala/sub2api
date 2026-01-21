CREATE TABLE IF NOT EXISTS admin_action_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_id BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(64) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id BIGINT DEFAULT NULL,
    payload TEXT DEFAULT NULL,
    ip_address VARCHAR(64) DEFAULT NULL,
    user_agent TEXT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_action_logs_admin_id ON admin_action_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_action_logs_resource_type ON admin_action_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_admin_action_logs_created_at ON admin_action_logs(created_at);

COMMENT ON TABLE admin_action_logs IS '管理员操作日志';
COMMENT ON COLUMN admin_action_logs.action IS '操作类型';
COMMENT ON COLUMN admin_action_logs.resource_type IS '资源类型';
COMMENT ON COLUMN admin_action_logs.resource_id IS '资源ID';
