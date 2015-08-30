-- vim: ft=sql:ts=4:sw=4:et
-- try6 DB SCHEMA
--

-------------------------------------------------------
--                                                   --
-- Instala el esquema por defecto para la base de    --
-- datos try6                                        --
--                                                   --
-- IMPORTANTE! DEBE EJECUTARSE COMO USUARIO postgres --
--     su - postgres -c "psql < schema.pgsql"        --
-------------------------------------------------------

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = off;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET escape_string_warning = off;

CREATE DATABASE try6db WITH ENCODING = 'UTF8';
ALTER DATABASE try6db OWNER TO try6adm;

\connect try6db

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;
COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';
CREATE EXTENSION IF NOT EXISTS "hstore" WITH SCHEMA public;
SET search_path = public, pg_catalog;
SET default_tablespace = '';
SET default_with_oids = false;

-- ----------------------------
--  Table structure for "tenants"
-- ----------------------------
CREATE TABLE IF NOT EXISTS tenants (
    id        UUID NOT NULL DEFAULT uuid_generate_v4(),
    label     VARCHAR(200),
    status    VARCHAR(50) NOT NULL DEFAULT 'active',
    created   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated   TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted   TIMESTAMP,

    CONSTRAINT tenants_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE tenants OWNER TO try6adm;
CREATE INDEX tenant_idx ON tenants USING btree (id);

-- ----------------------------
--  Table structure for "scopes"
-- ----------------------------
CREATE TABLE IF NOT EXISTS scopes (
    id        UUID NOT NULL DEFAULT uuid_generate_v4(),
    tenant_id UUID,
    label     VARCHAR(200),
    description VARCHAR(200),
    status    VARCHAR(50) NOT NULL DEFAULT 'active',
    created   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated   TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted   TIMESTAMP,

    CONSTRAINT scopes_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE scopes OWNER TO try6adm;
CREATE INDEX scope_idx ON scopes USING btree (id);
CREATE INDEX scope_tenantid_idx ON scopes USING btree (tenant_id);

-- ----------------------------
--  Table structure for "directories"
-- ----------------------------
CREATE TABLE IF NOT EXISTS directories (
    id          UUID NOT NULL DEFAULT uuid_generate_v4(),
    tenant_uid  UUID,
    label       VARCHAR(200),
    description VARCHAR(200),
    status      VARCHAR(50) NOT NULL DEFAULT 'active',
    created     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated     TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted     TIMESTAMP,

    CONSTRAINT directories_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE directories OWNER TO try6adm;
CREATE INDEX directories_idx ON directories USING btree (id);
CREATE INDEX directories_tenantid_idx ON directories USING btree (tenant_uid);

-- ----------------------------
-- Table structure for "passoword policies"
-- ----------------------------
CREATE TABLE IF NOT EXISTS password_creation_policies (
    id             UUID NOT NULL DEFAULT uuid_generate_v4(),
    directory_id   UUID,
    min_pass_len   INT,
    max_pass_len   INT,
    min_req_lcase  INT,
    min_req_ucase  INT,
    min_req_num    INT,
    min_req_sym    INT,
    min_req_dia    INT,
    created        TIMESTAMP NOT NULL DEFAULT NOW(),
    updated        TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted        TIMESTAMP,

    CONSTRAINT password_creation_policies_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE password_creation_policies OWNER TO try6adm;
CREATE INDEX password_creation_policies_idx ON password_creation_policies USING btree (id);
CREATE INDEX password_creation_policies_directoryid_idx ON password_creation_policies USING btree (directory_id);

-- ----------------------------
--  Table structure for "directory scope mapping"
-- ----------------------------
CREATE TABLE IF NOT EXISTS directory_scope (
    id                       UUID NOT NULL DEFAULT uuid_generate_v4(),
    directory_id             UUID,
    scope_id                 UUID,
    priority                 INT,
    is_default_account_store BOOLEAN,
    is_default_group_store   BOOLEAN,
    is_default_rbac_store    BOOLEAN,
    created                  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated                  TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted                  TIMESTAMP,

    CONSTRAINT directory_scope_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE directory_scope OWNER TO try6adm;
CREATE INDEX directory_scope_idx ON directory_scope USING btree (id);
CREATE INDEX directory_scope_directoryid_idx ON directory_scope USING btree (directory_id);
CREATE INDEX directory_scope_scopeid_idx ON directory_scope USING btree (scope_id);

-- ----------------------------
--  Table structure for "accounts"
-- ----------------------------
CREATE TABLE IF NOT EXISTS accounts (
    id        UUID NOT NULL DEFAULT uuid_generate_v4(),
    email     VARCHAR(100),
    name      VARCHAR(200),
    password  VARCHAR(60),
    status    VARCHAR(50) NOT NULL DEFAULT 'active',
    created   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated   TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted   TIMESTAMP,

    CONSTRAINT accounts_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE accounts OWNER TO try6adm;
CREATE INDEX account_idx ON accounts USING btree (id);
CREATE INDEX account_email_idx ON accounts USING btree (email);

-- ----------------------------
--  Table structure for "directory account mapping"
-- ----------------------------
CREATE TABLE IF NOT EXISTS directory_account (
    directory_id             UUID,
    account_id               UUID,
    created                  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated                  TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted                  TIMESTAMP,

    CONSTRAINT directory_account_pkey PRIMARY KEY (directory_id, account_id)
)
WITH (OIDS=FALSE);
ALTER TABLE directory_account OWNER TO try6adm;

--------------------------------------------------
-- Table structure for "keys"
--------------------------------------------------
CREATE TABLE IF NOT EXISTS keys (
  id          UUID NOT NULL DEFAULT uuid_generate_v4(),
  account_id  UUID,
  priv_key    CHARACTER VARYING,
  pub_key     CHARACTER VARYING,
  status      VARCHAR(50) NOT NULL DEFAULT 'active',
  created     TIMESTAMP NOT NULL DEFAULT now(),
  updated     TIMESTAMP DEFAULT NULL,
  deleted     TIMESTAMP DEFAULT NULL,

  CONSTRAINT keys_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE keys OWNER TO try6adm;
CREATE INDEX keys_idx ON keys USING btree (id, kid, pub_key, active);
CREATE INDEX keys_account_idx ON keys USING btree (account_id);

--------------------------------------------------
-- Table structure for "jwt"
--------------------------------------------------
CREATE TABLE IF NOT EXISTS jwt (
  id             UUID NOT NULL DEFAULT uuid_generate_v4(),
  signing_method CHARACTER VARYING,
  expires        TIMESTAMP DEFAULT NULL,
  status         VARCHAR(50) NOT NULL DEFAULT 'active',
  created        TIMESTAMP NOT NULL DEFAULT now(),
  updated        TIMESTAMP DEFAULT NULL,
  deleted        TIMESTAMP DEFAULT NULL,

  CONSTRAINT jwt_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE jwt OWNER TO try6adm;
CREATE INDEX jwt_idx ON jwt USING btree (id, active);

--------------------------------------------------
-- Table structure for "account_custom_data"
--------------------------------------------------
CREATE TABLE IF NOT EXISTS account_custom_data (
  id          UUID NOT NULL DEFAULT uuid_generate_v4(),
  account_id  UUID,
  data        HSTORE,

  CONSTRAINT account_custom_data_pkey PRIMARY KEY (id)
)
WITH (OIDS=FALSE);
ALTER TABLE account_custom_data OWNER TO try6adm;
CREATE INDEX account_custom_data_account_idx ON account_custom_data USING btree (account_id);

-- CREATE TABLE rbac_role (
--     id SERIAL NOT NULL PRIMARY KEY,
--     slug VARCHAR(256) UNIQUE NOT NULL,
--     name VARCHAR(256),
--     description TEXT DEFAULT '',
--     parameters JSONB DEFAULT '[]',
--     created TIMESTAMP  NOT NULL DEFAULT NOW(),
--     updated TIMESTAMP  NOT NULL
-- )
-- WITH (OIDS=FALSE);
-- ALTER TABLE public.rbac_role OWNER TO try6adm;
-- CREATE INDEX rbac_role_idx ON rbac_role USING btree (id, name);
--
-- CREATE TABLE rbac_grant (
--     id SERIAL NOT NULL PRIMARY KEY,
--     from_role INT,
--     to_role INT,
--     assigment JSONB NOT NULL DEFAULT '{}',
--
--     CONSTRAINT memberships_granted_fkey
--         FOREIGN KEY (from_role)
--         REFERENCES rbac_role (id)
--         ON DELETE CASCADE NOT DEFERRABLE,
--     CONSTRAINT members_fkey
--         FOREIGN KEY (to_role)
--         REFERENCES rbac_role (id)
--         ON DELETE CASCADE NOT DEFERRABLE
-- )
-- WITH (OIDS=FALSE);
-- ALTER TABLE public.rbac_grant OWNER TO try6adm;






--
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: try6adm
--

-- COPY accounts (id, uid, name, email, password, active, gravatar, created, updated) FROM stdin (DELIMITER ',');
-- 1,ce30ed61-6b5d-4136-95a3-ab11e3e97d87,Test account,account@test.com,$2a$10$T/tj9OCnQ4XUf7qcVsQsIuV9AxQgHaoaNxSOEnvdGdm.BEPpEG56e,true,\N,2013-08-18 17:46:23.748705,2013-08-18 17:46:23.748705
-- \.
--
-- ALTER SEQUENCE IF EXISTS accounts_id_seq RESTART WITH 2;

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT  ALL ON SCHEMA public TO   postgres;
GRANT  ALL ON SCHEMA public TO   PUBLIC;

COMMIT;
