-- Add color field to application_groups table
ALTER TABLE application_groups ADD COLUMN color VARCHAR(20) DEFAULT '#1890ff' AFTER icon;
