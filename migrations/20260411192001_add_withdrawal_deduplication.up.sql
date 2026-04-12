ALTER TABLE withdrawals
ADD CONSTRAINT withdrawals_user_order_unique UNIQUE (user_id, order_number);