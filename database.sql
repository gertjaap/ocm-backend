--
-- PostgreSQL database dump
--

-- Dumped from database version 12.4
-- Dumped by pg_dump version 12.2

-- Started on 2021-02-24 23:05:26 UTC

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 203 (class 1259 OID 958494)
-- Name: blocks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.blocks (
    id bigint NOT NULL,
    height integer,
    hash bytea
);


ALTER TABLE public.blocks OWNER TO postgres;

--
-- TOC entry 202 (class 1259 OID 958492)
-- Name: blocks_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.blocks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.blocks_id_seq OWNER TO postgres;

--
-- TOC entry 2967 (class 0 OID 0)
-- Dependencies: 202
-- Name: blocks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.blocks_id_seq OWNED BY public.blocks.id;


--
-- TOC entry 209 (class 1259 OID 958532)
-- Name: outputs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.outputs (
    id bigint NOT NULL,
    script_id bigint,
    vout bigint,
    value bigint,
    created_in_tx bigint,
    spent_in_tx bigint,
    coinbase boolean
);


ALTER TABLE public.outputs OWNER TO postgres;

--
-- TOC entry 208 (class 1259 OID 958530)
-- Name: outputs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.outputs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.outputs_id_seq OWNER TO postgres;

--
-- TOC entry 2968 (class 0 OID 0)
-- Dependencies: 208
-- Name: outputs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.outputs_id_seq OWNED BY public.outputs.id;


--
-- TOC entry 205 (class 1259 OID 958505)
-- Name: scripts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scripts (
    id bigint NOT NULL,
    script bytea
);


ALTER TABLE public.scripts OWNER TO postgres;

--
-- TOC entry 204 (class 1259 OID 958503)
-- Name: scripts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scripts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.scripts_id_seq OWNER TO postgres;

--
-- TOC entry 2969 (class 0 OID 0)
-- Dependencies: 204
-- Name: scripts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scripts_id_seq OWNED BY public.scripts.id;


--
-- TOC entry 207 (class 1259 OID 958516)
-- Name: transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transactions (
    id bigint NOT NULL,
    block_id bigint,
    hash bytea,
    received timestamp without time zone
);


ALTER TABLE public.transactions OWNER TO postgres;

--
-- TOC entry 206 (class 1259 OID 958514)
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.transactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.transactions_id_seq OWNER TO postgres;

--
-- TOC entry 2970 (class 0 OID 0)
-- Dependencies: 206
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.transactions_id_seq OWNED BY public.transactions.id;


--
-- TOC entry 2813 (class 2604 OID 958497)
-- Name: blocks id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.blocks ALTER COLUMN id SET DEFAULT nextval('public.blocks_id_seq'::regclass);


--
-- TOC entry 2816 (class 2604 OID 958535)
-- Name: outputs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.outputs ALTER COLUMN id SET DEFAULT nextval('public.outputs_id_seq'::regclass);


--
-- TOC entry 2814 (class 2604 OID 958508)
-- Name: scripts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scripts ALTER COLUMN id SET DEFAULT nextval('public.scripts_id_seq'::regclass);


--
-- TOC entry 2815 (class 2604 OID 958519)
-- Name: transactions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);


--
-- TOC entry 2820 (class 2606 OID 958502)
-- Name: blocks blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (id);


--
-- TOC entry 2831 (class 2606 OID 958537)
-- Name: outputs outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.outputs
    ADD CONSTRAINT outputs_pkey PRIMARY KEY (id);


--
-- TOC entry 2823 (class 2606 OID 958513)
-- Name: scripts scripts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scripts
    ADD CONSTRAINT scripts_pkey PRIMARY KEY (id);


--
-- TOC entry 2826 (class 2606 OID 958524)
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- TOC entry 2817 (class 1259 OID 958554)
-- Name: blocks_idx_hash; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX blocks_idx_hash ON public.blocks USING btree (hash);


--
-- TOC entry 2818 (class 1259 OID 958553)
-- Name: blocks_idx_height; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX blocks_idx_height ON public.blocks USING btree (height DESC NULLS LAST);


--
-- TOC entry 2827 (class 1259 OID 958727)
-- Name: outputs_idx_created_vout_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX outputs_idx_created_vout_unique ON public.outputs USING btree (created_in_tx, vout);


--
-- TOC entry 2828 (class 1259 OID 958555)
-- Name: outputs_idx_script; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX outputs_idx_script ON public.outputs USING btree (script_id);


--
-- TOC entry 2829 (class 1259 OID 958557)
-- Name: outputs_idx_spent; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX outputs_idx_spent ON public.outputs USING btree (spent_in_tx);


--
-- TOC entry 2821 (class 1259 OID 958560)
-- Name: scripts_idx_script; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX scripts_idx_script ON public.scripts USING btree (script);


--
-- TOC entry 2824 (class 1259 OID 958920)
-- Name: transaction_hash; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX transaction_hash ON public.transactions USING btree (hash);


--
-- TOC entry 2833 (class 2606 OID 958538)
-- Name: outputs fkey_output_created_tx; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.outputs
    ADD CONSTRAINT fkey_output_created_tx FOREIGN KEY (created_in_tx) REFERENCES public.transactions(id);


--
-- TOC entry 2835 (class 2606 OID 958548)
-- Name: outputs fkey_output_script; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.outputs
    ADD CONSTRAINT fkey_output_script FOREIGN KEY (script_id) REFERENCES public.scripts(id);


--
-- TOC entry 2834 (class 2606 OID 958543)
-- Name: outputs fkey_output_spent_tx; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.outputs
    ADD CONSTRAINT fkey_output_spent_tx FOREIGN KEY (spent_in_tx) REFERENCES public.transactions(id);


--
-- TOC entry 2832 (class 2606 OID 958525)
-- Name: transactions fkey_transaction_block; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT fkey_transaction_block FOREIGN KEY (block_id) REFERENCES public.blocks(id);


-- Completed on 2021-02-24 23:05:26 UTC

--
-- PostgreSQL database dump complete
--

