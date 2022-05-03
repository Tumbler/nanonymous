--
-- PostgreSQL database dump
--

-- Dumped from database version 14.2 (Ubuntu 14.2-1.pgdg21.10+1)
-- Dumped by pg_dump version 14.2 (Ubuntu 14.2-1.pgdg21.10+1)

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
    hash bytea NOT NULL
);


ALTER TABLE public.blacklist OWNER TO test;

--
-- Name: seeds; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.seeds (
    id integer NOT NULL,
    seed bytea NOT NULL,
    current_index bigint
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
-- Name: wallets; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.wallets (
    parent_seed integer NOT NULL,
    index bigint NOT NULL,
    balance numeric(40,0) NOT NULL,
    hash bytea NOT NULL
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
\\xf87df2b018b571951de85e46a83e01e2a9ca68895e3d5dd47152d42da4753e70
\\x61e2b27a37be010fa830ec27b419106169fc0f81df71bf1569203655f0463da2
\\x6269043470cbad23d548ff241abd68df0e87071f10a341ba03e47ad78157175e
\.


--
-- Data for Name: seeds; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.seeds (id, seed, current_index) FROM stdin;
1	\\xc30d0407030281e95f9ff270668e76d25101204eb914cbc5f1596b685976759cd794af8e3b96702aa2d3c215267b31a91b719fee6a6f98cf5967e62bd3aeb77e5aea691e2918a8e70571c5f9ddbc46330b4bc75ef3b2343b2866c0fd0e32ae8d6527	2
\.


--
-- Data for Name: wallets; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.wallets (parent_seed, index, balance, hash) FROM stdin;
1	0	0	\\x89cce4e0bf55e2745dc49db8804d4a61510efeee87e229f24ff713b5b8a4cd97
1	1	0	\\xe6b8ca5007d0f5f8c44829f60efc3a8d40fed98ae585b72887256d60ee0cd84b
1	2	0	\\x0aa19bbb5e6e281e89e6d7cb22f9bf5600d196f69a0ae3b47de9d75030d969c5
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

