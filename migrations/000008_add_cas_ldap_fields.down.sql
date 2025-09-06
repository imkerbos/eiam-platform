-- 回滚CAS和LDAP配置字段
ALTER TABLE applications DROP COLUMN bind_password;
ALTER TABLE applications DROP COLUMN bind_dn;
ALTER TABLE applications DROP COLUMN base_dn;
ALTER TABLE applications DROP COLUMN ldap_url;
ALTER TABLE applications DROP COLUMN renew;
ALTER TABLE applications DROP COLUMN gateway;
ALTER TABLE applications DROP COLUMN service_url;
