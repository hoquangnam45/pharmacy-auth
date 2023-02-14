CREATE SCHEMA auth;

CREATE TABLE auth.config(
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  key VARCHAR(255) NOT NULL,
  namespace VARCHAR(255),
  value VARCHAR(255)
);
ALTER TABLE auth.config ADD UNIQUE(namespace, key);

CREATE TABLE auth.login_detail(
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  activated BOOLEAN DEFAULT FALSE
);

CREATE TABLE auth.client(
  id INTEGER NOT NULL PRIMARY KEY,
  client_id VARCHAR(255) NOT NULL UNIQUE,
	signing_key VARCHAR(255) NOT NULL,
	verification_key VARCHAR(255) NOT NULL,
	name VARCHAR(255) NOT NULL,
	application_type INTEGER NOT NULL,
	active BOOLEAN DEFAULT TRUE,
  refresh_token_ttl INTERVAL NOT NULL,
  access_token_ttl INTERVAL NOT NULL,
  allow_origin VARCHAR(255) NOT NULL
);

CREATE TABLE auth.refresh_token(
  id VARCHAR(255) NOT NULL PRIMARY KEY,
  client_id VARCHAR(255) NOT NULL REFERENCES auth.client(id),
  subject VARCHAR(255) NOT NULL,
  issued_at TIMESTAMP NOT NULL,
  expired_at TIMESTAMP NOT NULL,
  protected_ticket VARCHAR(255) NOT NULL
);

CREATE TABLE auth.role (
  id VARCHAR(255) NOT NULL UNIQUE
);
INSERT INTO auth.role VALUES('ADMIN');

CREATE TABLE auth.user_role_association (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.login_detail(id) ON DELETE CASCADE,
  role_id VARCHAR(255) NOT NULL REFERENCES auth.role(id) ON DELETE CASCADE
);
ALTER TABLE auth.user_role_association ADD UNIQUE(user_id, role_id);

CREATE TABLE auth.policy (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  allow BOOLEAN DEFAULT FALSE,
  resource_pattern VARCHAR(255) NOT NULL
);
ALTER TABLE auth.policy ADD UNIQUE(allow, resource_pattern);
INSERT INTO auth.policy(allow, resource_pattern) VALUES(true, '*');

CREATE TABLE auth.role_policy_association(
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  role_id UUID NOT NULL REFERENCES auth.role(id) ON DELETE CASCADE,
  policy_id UUID NOT NULL REFERENCES auth.policy(id) ON DELETE CASCADE
);
ALTER TABLE auth.role_policy_association ADD UNIQUE(role_id, policy_id);
INSERT INTO auth.role_policy_association(role_id, policy_id)
SELECT auth.role.id, auth.policy.id FROM auth.role, auth.policy WHERE auth.role.role_name = 'ADMIN' AND auth.policy.allow = TRUE AND auth.policy.resource_pattern = '*';

CREATE TABLE auth.user_policy_association(
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.login_detail(id) ON DELETE CASCADE,
  policy_id UUID NOT NULL REFERENCES auth.policy(id) ON DELETE CASCADE
);
ALTER TABLE auth.role_policy_association ADD UNIQUE(role_id, policy_id);
