#!/usr/bin/env bash

# Reset database

PGPASSWORD=postgres
echo $PGPASSWORD

if (
psql -U postgres <<EOF
  select 1 from pg_database where datname='got';
EOF
) | grep -q 1; then

if (
psql -U postgres <<EOF
  drop database got;
EOF
) 2>&1 | grep ERROR; then
  exit 1
fi
fi

if (
psql -U postgres <<EOF
  create database got;
EOF
) 2>&1 | grep ERROR; then
  exit 1
fi

if (
psql -U postgres -d got <<EOF
  create table users (
    id serial primary key,
    email varchar unique not null,
    password varchar,
    name varchar,
    avatar_url varchar,
  );

  create table oauth (
    id serial primary key,
    user_id integer references users(id),
    provider varchar not null,
    provider_user_id varchar not null,
    unique(provider, provider_user_id)
  );

EOF
) 2>&1 | grep ERROR; then
  exit 1
fi
