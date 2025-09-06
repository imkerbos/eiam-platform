-- 删除插入的默认设置
DELETE FROM `system_settings` WHERE `id` IN (
  'setting-001', 'setting-002', 'setting-003', 'setting-004', 'setting-005',
  'setting-006', 'setting-007', 'setting-008', 'setting-009', 'setting-010',
  'setting-011', 'setting-012', 'setting-013', 'setting-014', 'setting-015', 'setting-016'
);

-- 注意：不删除表结构，因为可能被其他功能使用
