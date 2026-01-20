CREATE TABLE IF NOT EXISTS plans (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(120) NOT NULL,
    description TEXT DEFAULT NULL,
    price DECIMAL(20,8) NOT NULL DEFAULT 0,
    group_name VARCHAR(80) NOT NULL DEFAULT 'default',
    group_sort INTEGER NOT NULL DEFAULT 0,
    daily_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    total_quota DECIMAL(20,8) NOT NULL DEFAULT 0,
    purchase_qr_url TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plans_group_name ON plans(group_name);
CREATE INDEX IF NOT EXISTS idx_plans_group_sort ON plans(group_sort);
CREATE INDEX IF NOT EXISTS idx_plans_enabled ON plans(enabled);
CREATE INDEX IF NOT EXISTS idx_plans_sort_order ON plans(sort_order);

COMMENT ON TABLE plans IS '套餐配置';
COMMENT ON COLUMN plans.title IS '套餐标题';
COMMENT ON COLUMN plans.description IS '套餐描述';
COMMENT ON COLUMN plans.price IS '套餐价格（CNY）';
COMMENT ON COLUMN plans.group_name IS '分组名称';
COMMENT ON COLUMN plans.group_sort IS '分组排序';
COMMENT ON COLUMN plans.daily_quota IS '每日额度';
COMMENT ON COLUMN plans.total_quota IS '总额度';
COMMENT ON COLUMN plans.purchase_qr_url IS '购买二维码URL';
COMMENT ON COLUMN plans.enabled IS '是否启用';
COMMENT ON COLUMN plans.sort_order IS '排序';
