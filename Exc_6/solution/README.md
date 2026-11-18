# Exercise 6 – Card Sentinel

## 1. Decision

I choose log-based CDC with Debezium. It is the only option that leaves the
certified payment path untouched (R1): no code change like dual write or outbox,
and no extra queries like polling. Since Debezium reads the WAL, it gets every
committed payment in commit order and never sees rolled-back ones (R2), and
because the events land in Kafka, every model (or a new archive consumer) just
reads with its own consumer group. The cost is the replication slot: while the
connector is down, PostgreSQL keeps all unread WAL, and the disk can fill up
and stop the database. In production I would monitor the connector status and
the slot lag (pg_replication_slots, retained WAL size) with an alert long
before the disk is full, and set max_slot_wal_keep_size as a safety limit.
If our team owned the payment path, I would choose the outbox pattern,
because then we control the event format ourselves and do not depend on the
table schema.

## 2. The three runs

### Run 1 – Debezium defaults

Prediction: Debezium creates the topic with 1 partition, so only one
velocity instance works and it cannot keep up with 1,000 payments/s.

Scoreboard (velocity model):

- R3 frauds caught: 3/8 (38%), FAIL
- R3 false alerts: 0, PASS
- R4 alert latency p50/p95: 69.7 s / 92.5 s, FAIL

Why: The topic has only partition 0 and the key is the payment id. In a
consumer group one partition is read by exactly one instance, so vel-1 got
partition 0 and vel-2 and vel-3 got none (0 transactions). vel-1 handles only
about 440 payments/s, so its lag grew by about 3 s every 5 s, up to 90 s. R3
fails not because the model is wrong, but because the payments of most frauds
were still waiting in the backlog. The anomaly model passed (9/9, p95 0.8 s),
because one instance at 0.3 ms per payment is fast enough.

### Run 2 – 6 partitions

Prediction: all three instances work and R4 passes, but the key is still
the payment id, so the payments of one card are spread and R3 fails.

Scoreboard (velocity model):

- R3 frauds caught: 0/9 (0%), FAIL
- R3 false alerts: 0, PASS
- R4 alert latency p50/p95: 0.5 s / 0.5 s, PASS

Why: Each instance gets two partitions (0-1, 2-3 and 4-5), reads about 55,000
payments and keeps up with a lag of about 1 s. But the partition is chosen by
the hash of the payment id, so the 12 test payments of a stolen card land on
all instances. Each one counts only about 4 of them in 60 s and never reaches
the threshold of 8. The anomaly model still passes only because it runs as one
instance that reads all partitions.

### Run 3 – cardholder as key

Prediction: the key becomes the user_id, all payments of one cardholder go to
the same instance, so R3 and R4 both pass.

Scoreboard (velocity model):

- R3 frauds caught: 11/11 (100%), PASS
- R3 false alerts: 0, PASS
- R4 alert latency p50/p95: 0.7 s / 0.7 s, PASS

Why: The key is now the entity whose state the model keeps. All payments of a
cardholder are in one partition, read by one instance in commit order, so the
velocity count is complete. Each instance again had two partitions and a lag
of about 0.6 s. An occasional false alert would be an honest user who paid 8
times in a minute.

## 3. Closing answer

Debezium chose 1 partition and the primary key as message key. In production,
with 1 partition the velocity model could never scale: the lag would grow and
alerts would come minutes late, so stolen cards are blocked too late. With the
primary key the system would look healthy (no errors, no lag) but simply miss
almost all card-testing frauds once more than one instance runs. Nothing
crashes in either case, so the team could only notice through monitoring:
consumer lag and idle instances without partitions for the first problem, and
the detection rate compared to confirmed frauds (chargebacks) or a test with
injected frauds, like our scoreboard, for the second one.

## 4. Final connector

See fraud-connector.json in this folder.
