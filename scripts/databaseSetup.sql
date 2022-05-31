
-- run with:
-- psql -U postgres -f databaseSetup.sql

SELECT 'CREATE DATABASE nanonymousdb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'nanonymousdb')\gexec
\c nanonymousdb

CREATE TABLE IF NOT EXISTS seeds(
   id SERIAL PRIMARY KEY,
   seed BYTEA NOT NULL,
   current_index BIGINT,
   active BOOL DEFAULT true
);


CREATE TABLE IF NOT EXISTS wallets(
   parent_seed INT,
   index BIGINT NOT NULL,
   balance NUMERIC(40, 0) NOT NULL,
   hash BYTEA NOT NULL,
   CONSTRAINT seed
      FOREIGN KEY(parent_seed)
      REFERENCES seeds(id)
      ON DELETE CASCADE,
   pow TEXT,
   in_use BOOL DEFAULT true,
   PRIMARY KEY (parent_seed, index)
);

CREATE TABLE IF NOT EXISTS blacklist(
   hash BYTEA PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS transaction(
   unique_id BIGINT
);
INSERT INTO
   transaction (unique_id)
VALUES
   (-1);

CREATE TABLE IF NOT EXISTS profit_record(
   id SERIAL PRIMARY KEY,
   trans_id INT NOT NULL,
   time TIMESTAMPTZ NOT NULL,
   nano_gained NUMERIC(40,0) NOT NULL,
   nano_usd_value FLOAT8
);

CREATE EXTENSION pgcrypto;

CREATE USER go WITH PASSWORD 'my_password';
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO go;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO go;
GRANT INSERT ON ALL TABLES IN SCHEMA public TO go;
GRANT UPDATE ON ALL TABLES IN SCHEMA public TO go;
GRANT REFERENCES ON ALL TABLES IN SCHEMA public TO go;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO go;

REVOKE ALL ON ALL TABLES IN SCHEMA public FROM test;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM test;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA public FROM test;

