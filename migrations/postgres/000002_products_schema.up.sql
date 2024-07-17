CREATE TABLE IF NOT EXISTS products (
  id VARCHAR(24) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  price DECIMAL(10, 2) NOT NULL,
  category VARCHAR(255) NOT NULL,
  amount INT NOT NULL,
  added_at DATE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);
CREATE INDEX IF NOT EXISTS idx_products_name ON products (name);