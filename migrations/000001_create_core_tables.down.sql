-- Drop core tables for EIAM Platform (in reverse order)

-- Drop many-to-many relationship tables
DROP TABLE IF EXISTS organization_users;
DROP TABLE IF EXISTS role_application_groups;
DROP TABLE IF EXISTS role_applications;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS group_roles;
DROP TABLE IF EXISTS user_applications;
DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS user_roles;

-- Drop core tables
DROP TABLE IF EXISTS user_otp_records;
DROP TABLE IF EXISTS user_login_logs;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS application_groups;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;
