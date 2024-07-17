CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(24) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  registration_date DATE NOT NULL,
  role VARCHAR(255) NOT NULL CHECK (role IN ('admin', 'client'))
);

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
