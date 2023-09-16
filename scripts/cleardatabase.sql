
-- run with:
-- psql -U postgres -d nanonymousdb -f cleardatabase.sql

DROP TABLE IF EXISTS seeds CASCADE;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS blacklist;
DROP TABLE IF EXISTS transaction;
DROP TABLE IF EXISTS profit_record;
DROP EXTENSION IF EXISTS pgcrypto CASCADE;
