#!/bin/bash
: ${DB_USER:=db_USER}
: ${DB_PASSWORD:=db_pass}
: ${DB_NAME:=db_name}
: ${DB_ENCODING:=UTF-8}
: ${DB_PG_SCHEMA_FILE:=/tmp/schema.sql}


{
	gosu postgres psql <<-EOSQL
	CREATE USER "$DB_USER" WITH PASSWORD '$DB_PASSWORD';
EOSQL
} && {
	gosu postgres psql < ${DB_PG_SCHEMA_FILE}
}
