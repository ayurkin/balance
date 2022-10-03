#!/bin/bash

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_ADMIN_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  create role app with login password 'secret';

  create database app owner app;
EOSQL
