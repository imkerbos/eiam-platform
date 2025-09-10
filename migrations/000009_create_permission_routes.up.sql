-- Create permission_routes table
CREATE TABLE IF NOT EXISTS permission_routes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_code (code),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create permission_route_applications table (many-to-many)
CREATE TABLE IF NOT EXISTS permission_route_applications (
    permission_route_id VARCHAR(36) NOT NULL,
    application_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_route_id, application_id),
    FOREIGN KEY (permission_route_id) REFERENCES permission_routes(id) ON DELETE CASCADE,
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create permission_route_application_groups table (many-to-many)
CREATE TABLE IF NOT EXISTS permission_route_application_groups (
    permission_route_id VARCHAR(36) NOT NULL,
    application_group_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_route_id, application_group_id),
    FOREIGN KEY (permission_route_id) REFERENCES permission_routes(id) ON DELETE CASCADE,
    FOREIGN KEY (application_group_id) REFERENCES application_groups(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create permission_route_users table (many-to-many)
CREATE TABLE IF NOT EXISTS permission_route_users (
    permission_route_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_route_id, user_id),
    FOREIGN KEY (permission_route_id) REFERENCES permission_routes(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create permission_route_organizations table (many-to-many)
CREATE TABLE IF NOT EXISTS permission_route_organizations (
    permission_route_id VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_route_id, organization_id),
    FOREIGN KEY (permission_route_id) REFERENCES permission_routes(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
