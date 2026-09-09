-- will have and endpoint with this info, on network version change, wipe actor_types and actor_methods and repopulate
-- in actor_codes we only append
CREATE TABLE public.actor_types (
    actor_type_id integer
    actor_id integer -- populated only for singleton builtin actors
    actor_code_cid bytea
);

CREATE TABLE public.actor_codes (
    actor_type_id    integer
    actor_code_cid bytea
    network_version integer
);

CREATE TABLE public.actor_methods (
    actor_type_id    integer
    method_name text   NOT NULL,
    method      bigint NOT NULL,
    FOREIGN KEY (actor_type_id) REFERENCES actor_types (actor_type_id),
);

-- all fields can be extracted from ChainGetTipSet
CREATE TABLE public.block_headers (
    height            bigint NOT NULL,
    cid               bytea   NOT NULL, 
    miner_id          integer   NOT NULL,
    parent_weight     numeric NOT NULL,
    parent_base_fee   numeric NOT NULL,
    parent_state_root bytea   NOT NULL,
    timestamp         bigint NOT NULL,
    win_count         bigint,
    fork_signaling    bigint NOT NULL
);

-- some fields can be extracted from StateVMCirculatingSupplyInternal
-- the rest are fetched from power actor state

CREATE TABLE public.chain_economics (
    height                bigint  NOT NULL,
    state_root            bytea   NOT NULL,
    circulating_fil       numeric NOT NULL,
    vested_fil            numeric NOT NULL,
    mined_fil             numeric NOT NULL,
    burnt_fil             numeric NOT NULL,
    locked_fil            numeric NOT NULL,
    fil_reserve_disbursed numeric NOT NULL
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

-- over := gasLimit - (11*gasUsed)/10   // gasLimit/gasUsed - 1 - 0.1, scaled
-- if over < 0 { return gasLimit-gasUsed, 0 }  // no burn if overuse < 10%
-- if over > gasUsed { over = gasUsed }         // cap at 100%
-- gasToBurn = (gasLimit - gasUsed) * over / gasUsed

-- gas_burned = (gasLimit - gasUsed) * over / gasUsed, if gasUsed = 0, gas_burned = gasLimit
-- baseFeeToPay = min(baseFee, feeCap).
-- postgres virtual column -> available from postgres 18

CREATE TABLE public.derived_gas_outputs (
    height               bigint  NOT NULL,
    index_in_block       int,
    cid                  bytea   NOT NULL,
    state_root           bytea   NOT NULL,
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
    base_fee_burn        numeric GENERATED ALWAYS AS (min(parent_base_fee, gas_fee_cap)*gas_used),
    over_estimation_burn numeric GENERATED ALWAYS AS (min(parent_base_fee, gas_fee_cap)*gas_burned),
    miner_penalty        numeric NOT NULL, --conditional formula
    miner_tip            numeric GENERATED ALWAYS AS (min(gas_premium, gas_fee_cap - min(parent_base_fee, gas_fee_cap)) * gas_limit),
    refund               numeric GENERATED ALWAYS AS (gas_limit * gas_fee_cap) - base_fee_burn - miner_tip - over_estimation_burn,
    gas_refund           bigint  NOT NULL, -- (gasLimit - gasUsed) - gasToBurn gasToBurn has a conditional formula
    gas_burned           bigint  NOT NULL, --conditional formula
    actor_type_id        integer    NOT NULL,
    params               text    NOT NULL
);

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

-- we'll have an endpoint that returns all events for a given height
-- we'll filter the resulting json in code using the same values we would of used for filters
-- need to lookup how I did the encoding for fidl. most likely we need to encode all the sector event types for the filter

CREATE TABLE public.miner_sector_events (
    height     bigint                         NOT NULL,
    miner_id   integer                        NOT NULL,
    sector_id  bigint                         NOT NULL,
    state_root bytea                          NOT NULL,
    event      public.miner_sector_event_type NOT NULL
);

-- this can be populated at the same time as derived_gas_outputs
-- curently only for ProveCommitAggregate method seems to need decoded params for the panels, so we'll start only with that
-- decoding will be done via an endpoint similar to StateDecodeParams
CREATE TABLE public.decoded_params (
    height bigint,
    index_in_block int,
    params jsonb
    FOREIGN KEY (height, index_in_block) REFERENCES derived_gas_outputs (height, index_in_block)
);


-- in order to keep the data in sync with the chain we'll have a cron job that runs every 30s
-- and calls ChainGetHead. I will compare the cids/height with the data in block headers and 
-- on mismatch it will delete all data for that height in tables: block_headers, chain_economics, derived_gas_output, miner_sector_events
-- data in decoded params will get deleted because it references derived_gas_outputs through the foreign key
-- on mismatch it will continue and compare the previous height, and repeat this process until what's in the db matches the api call