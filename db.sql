-- most table structure is copied over from lily, as well as indexes and panel queries
CREATE TABLE public.actors (
    height     bigint NOT NULL,
    id         text NOT NULL,
    state_root text NOT NULL,
    code       text NOT NULL,
    head       text NOT NULL,
    balance    numeric NOT NULL,
    nonce      bigint NOT NULL,
    state      jsonb,
    code_cid   text
);
ALTER TABLE ONLY public.actors ADD CONSTRAINT actors_pkey PRIMARY KEY (height, id, state_root);
CREATE INDEX actors_height_idx ON public.actors USING btree (height DESC);
SELECT create_hypertable('actors', 'height', chunk_time_interval => 20160, if_not_exists => TRUE);
SELECT set_integer_now_func('actors', 'current_height', replace_if_exists => true);

CREATE TABLE public.actor_methods (
    actor_family      text   NOT NULL,
    method_name text   NOT NULL,
    method      bigint NOT NULL,
    PRIMARY KEY (family, method)
);

-- all fields can be extracted from ChainGetTipSet
CREATE TABLE public.block_headers (
    height            bigint NOT NULL,
    cid               text   NOT NULL, 
    miner_id          text   NOT NULL,
    parent_weight     numeric NOT NULL,
    parent_base_fee   numeric NOT NULL,
    parent_state_root text   NOT NULL,
    timestamp         bigint NOT NULL,
    win_count         bigint,
    fork_signaling    bigint NOT NULL
);

ALTER TABLE ONLY public.block_headers ADD CONSTRAINT block_headers_pkey PRIMARY KEY (height, cid);
CREATE INDEX block_headers_height_idx    ON public.block_headers USING btree (height DESC);
CREATE INDEX block_headers_timestamp_idx ON public.block_headers USING btree (timestamp);
SELECT create_hypertable('block_headers', 'height', chunk_time_interval => 20160, if_not_exists => TRUE);
SELECT set_integer_now_func('block_headers', 'current_height', replace_if_exists => true);

-- all fields, except parent_state_root can be extracted from StateVMCirculatingSupplyInternal
CREATE TABLE public.chain_economics (
    height                bigint  NOT NULL,
    parent_state_root     text    NOT NULL,
    circulating_fil       numeric NOT NULL,
    vested_fil            numeric NOT NULL,
    mined_fil             numeric NOT NULL,
    burnt_fil             numeric NOT NULL,
    locked_fil            numeric NOT NULL,
    fil_reserve_disbursed numeric NOT NULL
);
ALTER TABLE ONLY public.chain_economics ADD CONSTRAINT chain_economics_pk PRIMARY KEY (height, parent_state_root);

-- all fetched from power actor state
CREATE TABLE public.chain_powers (
    height                        bigint  NOT NULL,
    state_root                    text    NOT NULL,
    total_raw_bytes_power         numeric NOT NULL,
    total_raw_bytes_committed     numeric NOT NULL,
    total_qa_bytes_power          numeric NOT NULL,
    total_qa_bytes_committed      numeric NOT NULL,
    total_pledge_collateral       numeric NOT NULL,
    qa_smoothed_position_estimate numeric NOT NULL,
    qa_smoothed_velocity_estimate numeric NOT NULL,
    miner_count                   bigint,
    participating_miner_count     bigint
);
ALTER TABLE ONLY public.chain_powers ADD CONSTRAINT chain_powers_pkey PRIMARY KEY (height, state_root);
CREATE INDEX chain_powers_height_idx ON public.chain_powers USING btree (height DESC);

-- you need to call ChainGetParentReceipts and ChainGetParentMessages
-- from receipt you get gas_used and exit_code, you also get return but that doesn't seem to be needed
-- from message you get to, from, value, nonce, gas_limit, gas_fee_cap, gas_premium, method, cid
-- paren_base_fee can be fetched from the block_headers table
-- actorn_name and actor_family can be populated from the actors / actor methods table
-- fields that are not populated from requests:
-- size_bytes m.ChainLength(), the serialized byte size of the message, we need to get it somehow
-- following fields use gasUsed (from receipt), gasLimit, baseFee (block header), feeCap, gasPremium (from message):
-- base_fee_burn = min(baseFee, feeCap) * gasUsed 
-- over_estimation_burn = min(baseFee, feeCap) * gas_burned
-- miner_penalty (baseFee - feeCap) * gasUsed when feeCap < baseFee (sender underpaid), plus (baseFee - feeCap) * gas_burned for the burned portion. Zero when feeCap >= baseFee.
-- miner_tip = min(gasPremium, feeCap - baseFeeToPay) * gasLimit
-- refund = (gasLimit * feeCap) - base_fee_burn - miner_tip - over_estimation_burn
-- gas_refund = (gasLimit - gasUsed) - gasToBurn 
-- gas_burned = (gasLimit - gasUsed) * over / gasUsed, if gasUsed = 0, gas_burned = gasLimit

-- postgres virtual column -> investigate

CREATE TABLE public.derived_gas_outputs (
    msg_internal_id      bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
    height               bigint  NOT NULL,
    index_in_block       int,
    cid                  text    NOT NULL,
    state_root           text    NOT NULL,
    "from"               text    NOT NULL,
    "to"                 text    NOT NULL,
    value                numeric NOT NULL,
    gas_fee_cap          numeric NOT NULL,
    gas_premium          numeric NOT NULL,
    gas_limit            bigint,
    size_bytes           bigint,
    nonce                bigint,
    method               bigint,
    exit_code            bigint  NOT NULL,
    gas_used             bigint  NOT NULL,
    parent_base_fee      numeric NOT NULL,
    base_fee_burn        numeric NOT NULL,
    over_estimation_burn numeric NOT NULL,
    miner_penalty        numeric NOT NULL,
    miner_tip            numeric NOT NULL,
    refund               numeric NOT NULL,
    gas_refund           bigint  NOT NULL,
    gas_burned           bigint  NOT NULL,
    actor_name           text    NOT NULL,
    actor_family         text    NOT NULL,
    params               text    NOT NULL
);
ALTER TABLE ONLY public.derived_gas_outputs ADD CONSTRAINT derived_gas_outputs_pkey PRIMARY KEY (height, cid, state_root);
CREATE INDEX derived_gas_outputs_exit_code_index    ON public.derived_gas_outputs USING btree (exit_code);
CREATE INDEX derived_gas_outputs_from_index         ON public.derived_gas_outputs USING hash  ("from");
CREATE INDEX derived_gas_outputs_method_index       ON public.derived_gas_outputs USING btree (method);
CREATE INDEX derived_gas_outputs_to_index           ON public.derived_gas_outputs USING hash  ("to");
CREATE INDEX derived_gas_outputs_actor_family_index ON public.derived_gas_outputs USING btree (actor_family);
SELECT create_hypertable('derived_gas_outputs', 'height', chunk_time_interval => 20160, if_not_exists => TRUE);
SELECT set_integer_now_func('derived_gas_outputs', 'current_height', replace_if_exists => true);

CREATE TYPE public.miner_sector_event_type AS ENUM (
    'PRECOMMIT_ADDED',
    'PRECOMMIT_EXPIRED',
    'COMMIT_CAPACITY_ADDED',
    'SECTOR_ADDED',
    'SECTOR_EXTENDED',
    'SECTOR_EXPIRED',
    'SECTOR_FAULTED',
    'SECTOR_RECOVERING',
    'SECTOR_RECOVERED',
    'SECTOR_TERMINATED'
);

-- we'll use mainly GetActorEventsRaw with filters
-- a filter has this structure
-- [
--   {
--     "fields": {
--       "abc": [
--         {
--           "codec": 81,
--           "value": "ZGRhdGE="
--         }
--       ]
--     },
--     "fromHeight": 1010,
--     "toHeight": 1020
--   }
-- ]
-- need to lookup how we did it for fidl. most likely we need to encode all the sector event types for the filter
-- does this api has the same limitations as the lotus api? meaning 1000 / 10000 events per response and no way of paginating them

CREATE TABLE public.miner_sector_events (
    height     bigint                         NOT NULL,
    miner_id   text                           NOT NULL,
    sector_id  bigint                         NOT NULL,
    state_root text                           NOT NULL,
    event      public.miner_sector_event_type NOT NULL
);
ALTER TABLE ONLY public.miner_sector_events ADD CONSTRAINT miner_sector_events_pkey PRIMARY KEY (height, sector_id, event, miner_id, state_root);
CREATE INDEX miner_sector_events_height_idx ON public.miner_sector_events USING btree (height DESC);
SELECT create_hypertable('miner_sector_events', 'height', chunk_time_interval => 20160, if_not_exists => TRUE);
SELECT set_integer_now_func('miner_sector_events', 'current_height', replace_if_exists => true);

-- this can be populated at the same time as derived_gas_outputs
-- curently only for ProveCommitAggregate method seems to need decoded params for the panels, so we'll start only with that
CREATE TABLE public.decoded_params (
    msg_internal_id bigint,
    params jsonb
);

-- tabel meant to track if a state root is canonical or not
-- can be populated using ChainNotify
-- alternative process only tipsets older that 30 epochs
CREATE TABLE public.chain_state (
    height      bigint NOT NULL,
    tipset_key  text   NOT NULL,   
    parent_key  text   NOT NULL,
    state_root  text   NOT NULL,
    canonical   boolean NOT NULL DEFAULT true,
    reverted_at timestamptz
);