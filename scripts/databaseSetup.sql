
-- run with:
-- psql postgres -d nanonymousdb -f databaseSetup.sql

CREATE TABLE IF NOT EXISTS seeds(
   id SERIAL PRIMARY KEY,
   seed BYTEA NOT NULL,
   current_index BIGINT
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
   PRIMARY KEY (parent_seed, index)
);

CREATE TABLE IF NOT EXISTS blacklist(
   hash BYTEA PRIMARY KEY
);

CREATE EXTENSION pgcrypto;

