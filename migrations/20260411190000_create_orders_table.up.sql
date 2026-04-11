CREATE TABLE IF NOT EXISTS orders (
    id         SERIAL PRIMARY KEY,
    number     VARCHAR NOT NULL UNIQUE,
    user_id    BIGINT NOT NULL,
    status     VARCHAR NOT NULL,
    accrual    BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
