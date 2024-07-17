CREATE TABLE IF NOT EXISTS payments (
  id VARCHAR(24) PRIMARY KEY,
  user_id VARCHAR(24) REFERENCES users(id) ON DELETE CASCADE,
  order_id VARCHAR(24) REFERENCES orders(id) ON DELETE CASCADE,
  total_payment DECIMAL(10, 2) NOT NULL,
  payment_date DATE NOT NULL,
  status VARCHAR(255) NOT NULL CHECK (status IN ('success', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id);
