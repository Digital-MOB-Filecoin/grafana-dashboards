-- Gas usage & Message costs, fees, burned FIL 

-- Base fee over time
-- Tables: block_headers

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height)), to_timestamp(height_to_unix(0))) as time,
      avg(df.parent_base_fee::numeric) as fee
FROM block_headers df
WHERE df.height between unix_to_height($__unixEpochFrom()) and unix_to_height($__unixEpochTo())
GROUP BY time

Gas breakdown (exit_code=0)
Tables: derived_gas_outputs

WITH gases AS (SELECT height_to_unix(height) AS ts,
                     sum(gas_used)          AS gas_used,
                     sum(gas_limit)         AS gas_limit,
                     sum(gas_burned)        AS gas_burned,
                     sum(gas_refund)        AS gas_refund
              FROM derived_gas_outputs
              WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                AND exit_code = 0
              GROUP BY ts
              ORDER BY ts)
SELECT time_bucket('$bucket_size', to_timestamp(ts)) as "time",
      avg(gas_used),
      avg(gas_limit),
      avg(gas_burned),
      avg(gas_refund),
      avg(gas_used * 100 / gas_limit)               as "gas_used_pct"
FROM gases
GROUP BY time
ORDER BY time

-- Gas fill ratio
-- Tables: block_headers, derived_gas_outputs

WITH gas_limit AS (SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height)),
                                     to_timestamp(height_to_unix(0))) as time,
                         avg(epoch_gas_limit)                         as gas_limit_theoretical,
                         avg(epoch_gas_limit_sustainable)             as gas_limit_sustainable
                  FROM (SELECT height,
                               count(cid) * 10e9 as epoch_gas_limit,
                               count(cid) * 5e9  as epoch_gas_limit_sustainable
                        FROM block_headers
                        WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                        GROUP BY height) blocks
                  GROUP BY time
                  ORDER BY time),
    gas_used AS (SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height)),
                                    to_timestamp(height_to_unix(0))) as time,
                        avg(gas_used)                                as gas_used
                 FROM (SELECT height, sum(gas_used) as gas_used
                       FROM derived_gas_outputs
                       WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                         AND exit_code = 0
                       GROUP BY height
                       ORDER BY height) gpu
                 GROUP BY time
                 ORDER BY time)
SELECT gu.time,
      gas_limit_theoretical,
      gas_limit_sustainable,
      gu.gas_used                               as avg_gas_used,
      100 * gu.gas_used / gas_limit_theoretical as fill_ratio,
      100 * gu.gas_used / gas_limit_sustainable as fill_ratio_sustainable
FROM gas_used gu
        JOIN gas_limit gl ON gu.time = gl.time

-- Gas per tx (simplified)
-- Tables: derived_gas_outputs

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as "time",
      avg(gas_used),
      avg(gas_limit),
      avg(gas_burned),
      avg(gas_refund),
      avg(gas_used * 100 / gas_limit)                                   as "gas_used_pct"
FROM derived_gas_outputs
WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
 AND exit_code = 0
GROUP BY time
ORDER BY time

-- Gas used by actor/method
-- Tables: derived_gas_outputs, actor_methods

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(dg.height))) as time,
      sum(gas_used)                                                        as gas_used,
      dg.actor_family,
      am.method_name                                                       as actor_method
FROM derived_gas_outputs dg
        JOIN actor_methods am ON am.method = dg.method AND am.family = dg.actor_family
WHERE exit_code = 0
 AND dg.height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time, actor_family, actor_method
ORDER BY time

-- Gas % by actor/method (>0.5%)
-- Tables: derived_gas_outputs, actor_methods

WITH gas_per_method AS (SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(dg.height))) as time,
                              sum(gas_used)                                                        as gas_used,
                              dg.actor_family,
                              dg.method
                       FROM derived_gas_outputs dg
                       WHERE exit_code = 0
                         AND dg.height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                       GROUP BY time, actor_family, method
                       ORDER BY time),
    total_gas AS (SELECT time, sum(gas_used) as total_gas_used FROM gas_per_method GROUP BY time)
SELECT *
FROM (SELECT gm.time, gm.gas_used * 100 / tg.total_gas_used AS percent, gm.actor_family, am.method_name
     FROM gas_per_method gm
              JOIN total_gas tg ON gm.time = tg.time
              JOIN actor_methods am ON am.method = gm.method AND am.family = gm.actor_family) t
WHERE percent > 0.5
ORDER BY time

-- Total FIL gas cost breakdown
-- Tables: derived_gas_outputs

WITH fil_used AS (SELECT height_to_unix(height)                                                                  as ts,
                        sum(base_fee_burn::numeric + miner_tip::numeric + over_estimation_burn::numeric) /
                        1e18                                                                                    as total_cost,
                        sum(base_fee_burn::numeric) / 1e18                                                      as base_fee_burn,
                        sum(miner_tip::numeric) / 1e18                                                          as miner_tip,
                        sum(refund::numeric) / 1e18                                                             as refund,
                        sum(over_estimation_burn::numeric) / 1e18                                               as over_estimation_burn
                 FROM derived_gas_outputs
                 WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                   AND exit_code = 0
                 GROUP BY ts
                 ORDER BY ts)
SELECT time_bucket('$bucket_size', to_timestamp(ts)) AS "time",
      sum(total_cost),
      sum(base_fee_burn),
      sum(miner_tip),
      sum(over_estimation_burn),
      sum(refund)
FROM fil_used
GROUP BY time
ORDER BY time

-- Base fee, gas premium, miner penalty
-- Tables: derived_gas_outputs

WITH data AS (SELECT height_to_unix(height)              as ts,
                    avg(parent_base_fee::numeric / 1e9) as base_fee,
                    sum(miner_penalty::numeric / 1e9)   as miner_penalty,
                    sum(gas_premium::numeric / 1e9)     as gas_premium
             FROM derived_gas_outputs
             WHERE exit_code = 0
               AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
             GROUP BY ts
             ORDER BY ts)
SELECT time_bucket('$bucket_size', to_timestamp(ts)) AS "time",
      avg(gas_premium)                              AS gas_premium_avg,
      avg(base_fee)                                 as base_fee_avg,
      avg(miner_penalty)                            as miner_penalty_avg
FROM data
GROUP BY time
ORDER BY time

-- Avg FIL cost per tx
-- Tables: derived_gas_outputs

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height)))                       as time,
      avg(base_fee_burn::numeric + miner_tip::numeric + over_estimation_burn::numeric) / 1e18 as total_cost,
      avg(base_fee_burn::numeric) / 1e18                                                      as base_fee_burn,
      avg(miner_tip::numeric) / 1e18                                                          as miner_tip,
      avg(refund::numeric) / 1e18                                                             as refund,
      avg(over_estimation_burn::numeric) / 1e18                                               as over_estimation_burn
FROM derived_gas_outputs
WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
 AND exit_code = 0
GROUP BY time
ORDER BY time

FIL burn rate (f099 balance delta) 
Tables: actors
SELECT time_bucket('$bucket_size', to_timestamp(b.ts)) as time, avg(current - prev) as "total_fil_burn"
FROM (SELECT height_to_unix(height)                              as ts,
            height,
            balance::numeric / 1e18                             as current,
            lag(balance::numeric / 1e18) OVER (ORDER BY height) as prev,
            lag(height) OVER (ORDER BY height)                  as prev_height
     FROM actors
     WHERE id = 'f099'
       AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
     ORDER BY ts) b
WHERE (height - prev_height) = 1
GROUP BY time

-- PreCommit + ProveCommit update for new message types or replace
-- Tables: derived_gas_outputs

WITH data AS (SELECT height_to_unix(height)                                                                  as ts,
                    sum(base_fee_burn::numeric + miner_tip::numeric + over_estimation_burn::numeric) /
                    1e18                                                                                    as total_cost
             FROM derived_gas_outputs
             WHERE exit_code = 0
               AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
               AND actor_family = 'storageminer'
               AND (method = 7 OR method = 6)
             GROUP BY ts
             ORDER BY ts)
SELECT time_bucket('$bucket_size', to_timestamp(ts)) as "time",
      avg(total_cost)                               AS pre_commit_plus_prove_commit_burn
FROM data
GROUP BY time
ORDER BY time


-- Burnt FIL balance (f099)
-- Tables: actors

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as time,
      avg(balance::numeric / 1e18)                                      as burnt_fil_f099
FROM actors
WHERE id = 'f099'
 AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time
ORDER BY time


Burnt FIL from chain economics
Tables: chain_economics, block_headers

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(ce.height))) as time,
      max(burnt_fil::numeric / 1e18)                                       as burned_fil
FROM chain_economics ce
        JOIN block_headers bh ON ce.height = bh.height
WHERE ce.height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time
ORDER BY time

-- Burnt FIL rate + avg since launch
-- Tables: chain_economics

SELECT time_bucket('$bucket_size', to_timestamp(b.ts))           as time,
      avg(current - prev)                                       as "burnt_fil",
      avg((current - 3530755.411577548567) / (height - 148888)) as "avg_since_launch"
FROM (SELECT height_to_unix(height)                                as ts,
            height,
            burnt_fil::numeric / 1e18                             as current,
            lag(burnt_fil::numeric / 1e18) OVER (ORDER BY height) as prev,
            lag(height) OVER (ORDER BY height)                    as prev_height
     FROM chain_economics
     WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
       AND height > 148888
     ORDER BY ts) b
WHERE (height - prev_height) = 1
GROUP BY time

-- Total burnt FIL per period
-- Tables: chain_economics

WITH burnt_fils AS (SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as time,
                          max(burnt_fil::numeric / 1e18)                                    as burnt_fil
                   FROM chain_economics
                   WHERE height > 148888
                     AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                   GROUP BY time
                   ORDER BY time),
    lags AS (SELECT time, burnt_fil, lag(burnt_fil) OVER (ORDER BY time) as burnt_fil_last FROM burnt_fils)
SELECT time, burnt_fil - burnt_fil_last as total_burnt_fil
FROM lags

-- Proof aggregation FIL burn
-- Tables: parsed_messages, block_headers

WITH base_fees AS (SELECT height, max(parent_base_fee::numeric) as base_fee
                  FROM block_headers
                  WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                  GROUP BY height),
    proofs_batched AS (SELECT p.height, sum((params -> 'SectorNumbers' ->> 'elemcount')::numeric) as num_proofs_batched
                       FROM parsed_messages p
                       WHERE p.height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
                         AND method = 'ProveCommitAggregate'
                       GROUP BY p.height),
    gas_burnt AS (SELECT bf.height,
                         (greatest(2e9, base_fee) * 65733296.73 * num_proofs_batched::numeric * 0.05) / 1e18 as burn
                  FROM base_fees bf
                           JOIN proofs_batched pb ON bf.height = pb.height)
SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as time,
      sum(burn)                                                         as sum_proof_aggregation_fil_burn
FROM gas_burnt
GROUP BY time
ORDER BY time

-- Messages

-- Message count by method
-- Tables: parsed_messages, derived_gas_outputs

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(dg.height))) as time,
      count(pm.method)                                                     AS "count",
      pm.method                                                            AS method
FROM parsed_messages pm
        JOIN derived_gas_outputs dg ON dg.height = pm.height AND dg.cid = pm.cid
WHERE exit_code = 0
 AND pm.height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
 AND dg.height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time, pm.method
ORDER BY time, count DESC

-- Failed txs by actor/method
-- Tables: derived_gas_outputs

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(dg.height))) as time,
      actor_family,
      cast(method as text),
      count(dg.cid)                                                        AS "count"
FROM derived_gas_outputs dg
WHERE exit_code != 0
 AND dg.height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY 1, 2, 3
ORDER BY 1

-- Failed DisputeWindowedPoSt (method 25)
-- Tables: derived_gas_outputs

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) AS "time",
      count(cid)                                                        AS "exit",
      exit_code::text                                                   as exit_code
FROM derived_gas_outputs
WHERE exit_code != 0
 AND actor_family = 'storageminer'
 AND method = 25
 AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time, exit_code

-- Tx size distribution
-- Tables: derived_gas_outputs

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as time,
      count(cid)                                                        as bucket,
      to_char(size_bytes, '9000')                                       as name
FROM derived_gas_outputs
WHERE exit_code = 0
 AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time, name
ORDER BY time, name

-- Failed DisputeWindowedPoSt (method 24)
-- Tables: derived_gas_outputs

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) AS "time",
      count(cid)                                                        AS "exit",
      exit_code::text                                                   as exit_code
FROM derived_gas_outputs
WHERE exit_code != 0
 AND actor_family = 'storageminer'
 AND method = 24
 AND height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time, exit_code
ORDER BY time, exit_code

-- Network Power and active miners 

-- Network power growth
-- Tables: chain_powers

WITH powers AS (SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as time,
                      max(total_raw_bytes_power::numeric),
                      max(total_qa_bytes_power::numeric),
                      max(participating_miner_count)
               FROM chain_powers
               WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
               GROUP BY time
               ORDER BY time),
    lags AS (SELECT time, total_raw_bytes_power, lag(total_raw_bytes_power) OVER (ORDER BY time), ...FROM powers)
SELECT time,
      total_raw_bytes_power - total_raw_bytes_power_last,
      total_qa_bytes_power - total_qa_bytes_power_last,
      participating_miner_count
FROM lags

-- Total raw/QA bytes power
-- Tables: chain_powers

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as time,
      max(total_raw_bytes_power::numeric)                               as total_raw_bytes_power,
      max(total_qa_bytes_power::numeric)                                as total_qa_bytes_power
FROM chain_powers
WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time
ORDER BY time

-- Sectors (onboarding rate, events)
-- Sector event type breakdown
-- Tables: miner_sector_events

SELECT time_bucket('$bucket_size', to_timestamp(height_to_unix(height))) as time,
      count(event)                                                      as values,
      event                                                             as metric
FROM miner_sector_events
WHERE height BETWEEN unix_to_height($__unixEpochFrom()) AND unix_to_height($__unixEpochTo())
GROUP BY time, event
ORDER BY time, event

