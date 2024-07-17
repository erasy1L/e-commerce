CREATE TABLE IF NOT EXISTS orders (
  id VARCHAR(24) PRIMARY KEY,
  user_id VARCHAR(24) REFERENCES users(id) ON DELETE CASCADE,
  total_price DECIMAL(10, 2) NOT NULL,
  ordered_date DATE NOT NULL,
  status VARCHAR(255) NOT NULL CHECK (status IN ('new', 'pending', 'completed'))
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);

CREATE TABLE IF NOT EXISTS order_products (
  order_id VARCHAR(24) REFERENCES orders(id) ON DELETE CASCADE,
  product_id VARCHAR(24) REFERENCES products(id) ON DELETE CASCADE,
  amount INT NOT NULL,
  PRIMARY KEY (order_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_order_products_order_id ON order_products (order_id);
CREATE INDEX IF NOT EXISTS idx_order_products_product_id ON order_products (product_id);