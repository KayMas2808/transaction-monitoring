CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'viewer' CHECK (role IN ('admin', 'analyst', 'viewer')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20),
    location VARCHAR(100),
    risk_score DECIMAL(3,2) DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER/* REFERENCES customers(id)*/,
    amount DECIMAL(12,2) NOT NULL,
    currency CHAR(3) DEFAULT 'USD',
    merchant_id VARCHAR(50),
    merchant_category VARCHAR(50),
    location VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    payment_method VARCHAR(20),
    card_last_four CHAR(4),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'under_review')),
    fraud_score DECIMAL(5,4) DEFAULT 0.0,
    fraud_reasons TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_by INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER REFERENCES transactions(id),
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    message TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'resolved', 'false_positive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by INTEGER REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS fraud_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    rule_type VARCHAR(50) NOT NULL,
    parameters JSONB,
    threshold DECIMAL(10,4),
    weight DECIMAL(3,2) DEFAULT 1.0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id INTEGER,
    details JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_fraud_score ON transactions(fraud_score);
CREATE INDEX IF NOT EXISTS idx_alerts_transaction_id ON alerts(transaction_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

INSERT INTO users (username, email, password_hash, role) 
VALUES ('admin', 'admin@example.com', '$2a$10$GaAlTVe2fvNmPGhhCv2mbe8Tt361IxmvS1fifgICokTe1dzmZU1yy', 'admin')
ON CONFLICT (username) DO NOTHING;

INSERT INTO fraud_rules (name, description, rule_type, parameters, threshold, weight) VALUES
('High Amount Transaction', 'Flags transactions above threshold', 'amount_threshold', '{"currency": "USD"}', 10000.0, 1.0),
('Velocity Check', 'Flags rapid successive transactions', 'velocity', '{"time_window": "1 minute", "max_count": 3}', 3.0, 0.8),
('Location Anomaly', 'Flags transactions from unusual locations', 'location', '{"distance_threshold": 500}', 500.0, 0.6),
('Time Anomaly', 'Flags transactions at unusual hours', 'time_pattern', '{"start_hour": 2, "end_hour": 6}', 0.0, 0.4),
('Amount Pattern', 'Flags round number amounts', 'amount_pattern', '{"round_amounts": true}', 0.0, 0.3)
ON CONFLICT DO NOTHING; 