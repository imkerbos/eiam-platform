-- Delete sample organizations
DELETE FROM organizations WHERE id IN (
    'org-001', 'org-002', 'org-003', 'org-004', 'org-005', 'org-006', 
    'org-007', 'org-008', 'org-009', 'org-010', 'org-011'
);
