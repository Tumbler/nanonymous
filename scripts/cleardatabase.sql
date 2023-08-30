
-- run with:
-- psql postgres -d nanonymousdb -f cleardatabase.sql

DROP TABLE IF EXISTS seeds CASCADE;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS blacklist;
DROP EXTENSION IF EXISTS pgcrypto CASCADE;
