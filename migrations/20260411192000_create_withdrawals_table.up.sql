CREATE TABLE IF NOT EXISTS withdrawals (
    id            SERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number  VARCHAR NOT NULL,
    sum           BIGINT NOT NULL,
    processed_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
