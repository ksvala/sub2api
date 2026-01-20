ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_code VARCHAR(6) DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_invite_code_unique
    ON users(invite_code)
    WHERE invite_code IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_invites (
    id BIGSERIAL PRIMARY KEY,
    inviter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_code VARCHAR(6) NOT NULL,
    reward_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    confirmed_by BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    confirmed_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_invites_invitee_id
    ON user_invites(invitee_id);
CREATE INDEX IF NOT EXISTS idx_user_invites_inviter_id
    ON user_invites(inviter_id);
CREATE INDEX IF NOT EXISTS idx_user_invites_status
    ON user_invites(status);
CREATE INDEX IF NOT EXISTS idx_user_invites_created_at
    ON user_invites(created_at);

CREATE TABLE IF NOT EXISTS invite_logs (
    id BIGSERIAL PRIMARY KEY,
    invite_id BIGINT NOT NULL REFERENCES user_invites(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL,
    inviter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admin_id BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    reward_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_logs_action
    ON invite_logs(action);
CREATE INDEX IF NOT EXISTS idx_invite_logs_created_at
    ON invite_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_invite_logs_inviter_id
    ON invite_logs(inviter_id);
CREATE INDEX IF NOT EXISTS idx_invite_logs_invitee_id
    ON invite_logs(invitee_id);
CREATE INDEX IF NOT EXISTS idx_invite_logs_invite_id
    ON invite_logs(invite_id);

COMMENT ON TABLE user_invites IS '用户邀请关系';
COMMENT ON COLUMN user_invites.inviter_id IS '邀请人用户ID';
COMMENT ON COLUMN user_invites.invitee_id IS '被邀请用户ID';
COMMENT ON COLUMN user_invites.invite_code IS '注册时填写的邀请码';
COMMENT ON COLUMN user_invites.reward_amount IS '注册时快照的奖励额度';
COMMENT ON COLUMN user_invites.status IS '状态: pending, confirmed';
COMMENT ON COLUMN user_invites.confirmed_by IS '确认管理员ID';
COMMENT ON COLUMN user_invites.confirmed_at IS '确认时间';

COMMENT ON TABLE invite_logs IS '邀请操作日志';
COMMENT ON COLUMN invite_logs.action IS '操作类型: bind, confirm';
