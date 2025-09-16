-- SAML断言表
CREATE TABLE saml_assertions (
    id VARCHAR(36) PRIMARY KEY,
    assertion_id VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    username VARCHAR(100) NOT NULL,
    audience VARCHAR(500) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_saml_assertions_user_id (user_id),
    INDEX idx_saml_assertions_assertion_id (assertion_id),
    INDEX idx_saml_assertions_expires_at (expires_at)
);
