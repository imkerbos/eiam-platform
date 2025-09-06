-- 添加CAS和LDAP配置字段
ALTER TABLE applications ADD COLUMN service_url VARCHAR(500) AFTER digest_algorithm;
ALTER TABLE applications ADD COLUMN gateway BOOLEAN DEFAULT FALSE AFTER service_url;
ALTER TABLE applications ADD COLUMN renew BOOLEAN DEFAULT FALSE AFTER gateway;
ALTER TABLE applications ADD COLUMN ldap_url VARCHAR(500) AFTER renew;
ALTER TABLE applications ADD COLUMN base_dn VARCHAR(500) AFTER ldap_url;
ALTER TABLE applications ADD COLUMN bind_dn VARCHAR(500) AFTER base_dn;
ALTER TABLE applications ADD COLUMN bind_password VARCHAR(500) AFTER bind_dn;
