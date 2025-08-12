-- Insert sample organizations with tree structure
-- EIAM Company (Headquarters)
INSERT INTO organizations (id, name, code, type, parent_id, level, path, sort, description, manager, location, phone, email, status, created_at, updated_at) VALUES
('org-001', 'EIAM Company', 'EIAM', 1, NULL, 1, '/1', 1, 'EIAM Company Headquarters', 'user-001', 'New York, US', '+1-555-0100', 'hq@eiam.com', 1, NOW(), NOW()),

-- Departments under EIAM Company
('org-002', 'Human Resources', 'HR', 3, 'org-001', 2, '/1/2', 1, 'Human Resources Department', 'user-002', 'New York, US', '+1-555-0101', 'hr@eiam.com', 1, NOW(), NOW()),
('org-003', 'DevOps', 'DEVOPS', 3, 'org-001', 2, '/1/3', 2, 'DevOps Department', 'user-003', 'New York, US', '+1-555-0102', 'devops@eiam.com', 1, NOW(), NOW()),
('org-004', 'Development', 'DEV', 3, 'org-001', 2, '/1/4', 3, 'Development Department', 'user-004', 'New York, US', '+1-555-0103', 'dev@eiam.com', 1, NOW(), NOW()),

-- Teams under HR Department
('org-005', 'Recruitment Team', 'HR-REC', 4, 'org-002', 3, '/1/2/5', 1, 'Recruitment and hiring team', 'user-005', 'New York, US', '+1-555-0104', 'recruitment@eiam.com', 1, NOW(), NOW()),
('org-006', 'Training Team', 'HR-TRAIN', 4, 'org-002', 3, '/1/2/6', 2, 'Employee training and development team', 'user-006', 'New York, US', '+1-555-0105', 'training@eiam.com', 1, NOW(), NOW()),

-- Teams under DevOps Department
('org-007', 'Infrastructure Team', 'DEVOPS-INFRA', 4, 'org-003', 3, '/1/3/7', 1, 'Infrastructure and platform team', 'user-007', 'New York, US', '+1-555-0106', 'infra@eiam.com', 1, NOW(), NOW()),
('org-008', 'Security Team', 'DEVOPS-SEC', 4, 'org-003', 3, '/1/3/8', 2, 'Security and compliance team', 'user-008', 'New York, US', '+1-555-0107', 'security@eiam.com', 1, NOW(), NOW()),

-- Teams under Development Department
('org-009', 'Frontend Team', 'DEV-FE', 4, 'org-004', 3, '/1/4/9', 1, 'Frontend development team', 'user-009', 'New York, US', '+1-555-0108', 'frontend@eiam.com', 1, NOW(), NOW()),
('org-010', 'Backend Team', 'DEV-BE', 4, 'org-004', 3, '/1/4/10', 2, 'Backend development team', 'user-010', 'New York, US', '+1-555-0109', 'backend@eiam.com', 1, NOW(), NOW()),
('org-011', 'QA Team', 'DEV-QA', 4, 'org-004', 3, '/1/4/11', 3, 'Quality assurance team', 'user-011', 'New York, US', '+1-555-0110', 'qa@eiam.com', 1, NOW(), NOW());
