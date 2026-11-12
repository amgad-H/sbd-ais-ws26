"""The anomaly detector as a Kafka service. YOUR CODE (Part 2, Stage C2).

The plumbing is here, the decisions (marked TODO) are yours. Contract with
the maintenance team: every reading goes through model.score(); every
reading with score >= model.threshold becomes ONE anomaly event,
model.anomaly_event(reading, score), as JSON, published to the topic
`anomalies`.

    python3 detector_service.py --name det-1
    python3 detector_service.py --name det-2          # one more instance


Decisions (see solution/solution.ipynb):
  - in-topic `readings` (key = machine_id, 6 partitions), group `line-guardian`:
    all instances share one group, so every partition (and therefore every
    machine) is read by exactly one instance, in order.
  - On a rebalance, the model forgets the machines of the partitions it loses.
    Otherwise, if a partition comes back later, the model would continue with
    a stale state (old window, old next_seq -> wrong "missing" counts).
  - Offsets are committed before partitions are handed over, so the new owner
    continues exactly where this instance stopped (no gaps, few duplicates).
  - Anomaly key = machine_id: all alerts of a machine stay in order.
    
"""
import argparse
import json
import time
from collections import defaultdict

from confluent_kafka import Consumer, Producer, KafkaException

from model import BearingModel

ap = argparse.ArgumentParser()
ap.add_argument("--name", default="det-1", help="instance name (shows up in the scoreboard)")
ap.add_argument("--in-topic", default="readings", help="the topic the gateways write to")
ap.add_argument("--group", default="line-guardian", help="the consumer group shared by all detector instances")
ap.add_argument("--out-topic", default="anomalies")
ap.add_argument("--bootstrap", default="localhost:9092")
args = ap.parse_args()
if not args.in_topic or not args.group:
    raise SystemExit("Please set --in-topic and --group (or give them defaults above)")

model = BearingModel(instance=args.name)
consumer = Consumer({
    "bootstrap.servers": args.bootstrap,
    "group.id": args.group,
    "client.id": args.name,
    # "latest": a NEW group starts with live data. "earliest" would replay up to
    # 7 days of old readings first: the alerts would come hours late (R3).
    # Once the group has committed offsets, it continues from there anyway.

    "auto.offset.reset": "latest",     # why "latest" and not "earliest" for a live detector?
})
producer = Producer({
    "bootstrap.servers": args.bootstrap,
    "partitioner": "murmur2_random",
    "acks": "all",                 # an alert must not get lost
    "enable.idempotence": True,    # no duplicate / reordered alerts on retries
    "linger.ms": 5,
})

machines_of = defaultdict(set)     # partition -> machines this instance has seen in it


def on_assign(c, partitions):
    print(f"[{args.name}] assigned partitions {sorted(p.partition for p in partitions)}", flush=True)


def on_revoke(c, partitions):
    try:
        c.commit(asynchronous=False)            # hand over our exact position
    except KafkaException:
        pass                                    # nothing to commit yet
    for p in partitions:                        # we no longer own these machines: drop their state
        for m in machines_of.pop(p.partition, ()):
            model.forget(m)
    print(f"[{args.name}] revoked partitions {sorted(p.partition for p in partitions)}", flush=True)


consumer.subscribe([args.in_topic], on_assign=on_assign, on_revoke=on_revoke)
print(f"[{args.name}] {args.in_topic} -> {args.out_topic} (Ctrl+C to stop)", flush=True)

readings = anomalies = 0
report_at = time.time() + 5
try:
    while True:
        msg = consumer.poll(1.0)
        if msg is None:
            continue
        if msg.error():
            print(f"[{args.name}] {msg.error()}")
            continue
        reading = json.loads(msg.value())
        machines_of[msg.partition()].add(reading["machine_id"])
        score = model.score(reading)
        readings += 1
        if score is not None and score >= model.threshold:
            anomalies += 1
            event = model.anomaly_event(reading, score)
            producer.produce(args.out_topic, key=reading["machine_id"],       # machine id is key
                             value=json.dumps(event))
        producer.poll(0)
        if time.time() >= report_at:
            print(f"[{args.name}] {readings:>9,} readings  {anomalies:>6,} anomalies  "
                  f"missing {model.missing}  out-of-order {model.out_of_order}", flush=True)
            report_at += 5
except KeyboardInterrupt:
    pass
finally:
    producer.flush(10)
    consumer.close()
    print(f"[{args.name}] stopped after {readings:,} readings, {anomalies:,} anomalies")
