--
-- PostgreSQL database dump
--

-- Dumped from database version 14.9 (Ubuntu 14.9-0ubuntu0.22.04.1)
-- Dumped by pg_dump version 14.9 (Ubuntu 14.9-0ubuntu0.22.04.1)
-- pg_dump -c gotests > resetTestDatabase.sql

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE ONLY public.wallets DROP CONSTRAINT seed;
ALTER TABLE ONLY public.wallets DROP CONSTRAINT wallets_pkey;
ALTER TABLE ONLY public.seeds DROP CONSTRAINT seeds_pkey;
ALTER TABLE ONLY public.blacklist DROP CONSTRAINT blacklist_pkey;
ALTER TABLE public.seeds ALTER COLUMN id DROP DEFAULT;
DROP TABLE public.wallets;
DROP TABLE public.transaction;
DROP SEQUENCE public.seeds_id_seq;
DROP TABLE public.seeds;
DROP TABLE public.blacklist;
DROP EXTENSION pgcrypto;
--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: blacklist; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.blacklist (
    hash bytea PRIMARY KEY,
    seed_id integer NOT NULL DEFAULT 0
);


ALTER TABLE public.blacklist OWNER TO test;

--
-- Name: seeds; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.seeds (
    id integer NOT NULL,
    seed bytea NOT NULL,
    current_index bigint,
    active boolean DEFAULT true
);


ALTER TABLE public.seeds OWNER TO test;

--
-- Name: seeds_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.seeds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.seeds_id_seq OWNER TO test;

--
-- Name: seeds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.seeds_id_seq OWNED BY public.seeds.id;


--
-- Name: transaction; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.transaction (
    unique_id bigint
);


ALTER TABLE public.transaction OWNER TO test;

--
-- Name: wallets; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.wallets (
    parent_seed integer NOT NULL,
    index bigint NOT NULL,
    balance numeric(40,0) NOT NULL,
    hash bytea NOT NULL,
    in_use boolean DEFAULT true,
    receive_only boolean DEFAULT false,
    mixer boolean DEFAULT false,
    PRIMARY KEY (parent_seed, index)
);


ALTER TABLE public.wallets OWNER TO test;

--
-- Name: seeds id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.seeds ALTER COLUMN id SET DEFAULT nextval('public.seeds_id_seq'::regclass);


--
-- Data for Name: blacklist; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.blacklist (hash) FROM stdin;
\\x229f159f32139fac10a1e2135a9b2d05653f38794690008439dfcc0f3f6dcd1c
\\xa98cfe2664590f3646b0552a82760bf56c4c70732a378a3db0ce161d2a38a2c4
\\x07df75587b148251ea97a898d9a0968d188df87a70e0ee26f15104bb3e69ad5f
\\xad033a213924c894aff483e5bf62da600aa7f46df04826e07d98b5ea5e2a0be5
\.


--
-- Data for Name: seeds; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.seeds (id, seed, current_index, active) FROM stdin;
1	\\xc30d0407030281e95f9ff270668e76d25101204eb914cbc5f1596b685976759cd794af8e3b96702aa2d3c215267b31a91b719fee6a6f98cf5967e62bd3aeb77e5aea691e2918a8e70571c5f9ddbc46330b4bc75ef3b2343b2866c0fd0e32ae8d6527	6	t
\.


--
-- Data for Name: transaction; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.transaction (unique_id) FROM stdin;
-1
\.


--
-- Data for Name: wallets; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.wallets (parent_seed, index, balance, hash, in_use, receive_only, mixer) FROM stdin;
1	0	41000000000000000000000000000000	\\x89cce4e0bf55e2745dc49db8804d4a61510efeee87e229f24ff713b5b8a4cd97	f	f	f
1	1	600000000000000000000000000000	\\xe6b8ca5007d0f5f8c44829f60efc3a8d40fed98ae585b72887256d60ee0cd84b	f	f	f
1	2	3200000000000000000000000000000	\\x0aa19bbb5e6e281e89e6d7cb22f9bf5600d196f69a0ae3b47de9d75030d969c5	f	f	f
1	3	0	\\x29abfdb3f6be7cfd895550c13dae13d0030841b20f3f3a58b121bc12fbb0af0f	f	f	f
1	4	5000000000000000000000000000000	\\xbedfec468394732754fba3d3e2372dd80bb3048a4e1069dc6de3b1b36fb8f8f8	f	f	t
1	5	5000000000000000000000000000000	\\x41b3cd90a0286d00f657af8ba6efc405e69422174817acf2a009f016641bd2ed	f	f	t
1	6	10000000000000000000000000000000	\\x5288413269e90c7a8c05897dfd02de46ee289a2ec68fe602cc7fc9cb77fecf9e	f	f	t
\.


--
-- Name: seeds_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.seeds_id_seq', 1, true);


--
-- Name: blacklist blacklist_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.blacklist
    ADD CONSTRAINT blacklist_pkey PRIMARY KEY (hash);


--
-- Name: seeds seeds_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.seeds
    ADD CONSTRAINT seeds_pkey PRIMARY KEY (id);


--
-- Name: wallets wallets_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (parent_seed, index);


--
-- Name: wallets seed; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT seed FOREIGN KEY (parent_seed) REFERENCES public.seeds(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

