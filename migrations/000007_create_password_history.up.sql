-- Create password_history table
CREATE TABLE IF NOT EXISTS password_history (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create password_policies table
CREATE TABLE IF NOT EXISTS password_policies (
    id VARCHAR(36) PRIMARY KEY,
    min_length INT DEFAULT 8,
    max_length INT DEFAULT 128,
    require_uppercase BOOLEAN DEFAULT TRUE,
    require_lowercase BOOLEAN DEFAULT TRUE,
    require_numbers BOOLEAN DEFAULT TRUE,
    require_special_chars BOOLEAN DEFAULT TRUE,
    history_count INT DEFAULT 5,
    expiry_days INT DEFAULT 90,
    prevent_common BOOLEAN DEFAULT TRUE,
    prevent_username BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_deleted_at (deleted_at)
);

-- Insert default password policy
INSERT INTO password_policies (id, min_length, max_length, require_uppercase, require_lowercase, require_numbers, require_special_chars, history_count, expiry_days, prevent_common, prevent_username, is_active) 
VALUES ('policy-001', 8, 128, TRUE, TRUE, TRUE, TRUE, 5, 90, TRUE, TRUE, TRUE);
