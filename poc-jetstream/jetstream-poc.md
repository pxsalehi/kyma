## Phase 1 

What are the minimum change we need to have to keep our current NATS-based backend but move to JetStream to be able to 
have at-least-once guarantees?

- [F] Should we use Operator or Helm? How do we setup a cluster of size 1 or 3, that can use a PV?
  + NATS Operator officially discourages to use nats-operator for new deployments. Also support for JetStreams by nats-operator is questionable.
    > The recommended way of running NATS on Kubernetes is by using the Helm charts. If looking for JetStream support, this is supported in the Helm charts. The NATS Operator is not recommended to be used for new deployments.
([Reference](https://github.com/nats-io/nats-operator#nats-operator))
  + Clustering needs to be enabled and defined in helm chart.
  + Existing PVC can also be assigned for JetStream fileStorage as described [here](https://github.com/nats-io/k8s/tree/main/helm/charts/nats#using-with-an-existing-persistentvolumeclaim). But this needs to be explored further.

- [F,P] Do we need [NACK](https://github.com/nats-io/nack#getting-started) for configuration of streams?
  + NACK allows to manage JetStream streams and consumers using k8s CRDs.
  + Using NACK is optional, instead we can use the NATS Go client for JetStream management.

- [F,P] What is the current NATS workload works using Jetstream?
  + The concept of streams and consumers is new in Jetstream.
  + Streams basically provide a persistence for the events. 
  + We need to create Stream, and assign subjects to this stream. Any event published to these subjects will be received and stored by the stream.

- [F,P] How to configure streams and consumers?
  + Do we need one consumer per subscription?
    - We need atleast one consumer per subscription.
  + Do we need one stream for the whole backend, or multiple streams? A stream has its own configs 
    such as: storage, replicas, retention, deduplication. Each subject can only be in one stream. 
    We have two option:
    - Use one stream for all event types, which means one stream uses a wild-card (`>`) to match all
      event types. The issue is that the previous wild-card also matches internal JetStream events
      which can end up taking up space. The solution is to use the stream name as a prefix.
      + It is still possible to set one retention limit that applies independently to each event type.
      + If we do not need to have these configs customizable per event type, no need to use multiple streams.
    - Use one stream for each `app.*.*.*`. We would need to dynamically create the stream for each event type.
  
- [F] What changes are required in our NATS-based reconciler/dispatcher to use the JetStream backend?
  + How many consumers do we need? One per subscription? 
    - We need atleast one consumer per subscription.
  + Pull or push-based? Push-based seems to be the closest to the current model we have with NATS.
    - Push-based would add matching publications of a consumer to a separate subj.
  + The workflow for defining Streams for subscriptions needs to be implemented.
  + The subscribers creation part (or dispatcher) needs to be replaced by JetStream context-based subscribers/consumers.
  + The publisher-proxy also needs to be updated to use JetStream context-based publisher.
  
- [F,P] How to migrate a cluster from a NATS-based to a JetStream-based backend, and how would that impact existing subscriptions? 
  + Considering we have no persistent messages in NATS, if we switch during a maintenance window when there are no new
  publications, we should not need any other action. Subscription migration would not be necessary.
  
- [P] What are the options for encryption when using file-based streams?
  > Performance considerations: As expected, encryption is likely to decrease performance, but by how much is hard 
  > to define. In some performance tests on a MacbookPro 2.8 GHz Intel Core i7 with SSD, we have observed as little 
  > as 1% decrease to more than 30%. In addition to CPU cycles required for encryption, the encrypted files may be 
  > larger, which results in more data being stored or read.
  + They recommend using file-system level encryption
  + JS supports also NATS-level encryption. Not much info on key rotation, etc!
  + Question: Is it a deal-breaker to not use encryption by default?
  
- [P] Benchmark/estimate:
  + Cluster stability in the long run
  + Performance:
    + **One stream for all** and 1 stream per `app.*.*.*`
    + Memory- and **file-based** streams
    + Replication (cluster size 1 or **cluster size 3**)
  
## Benchmarks

- Gardener Cluster with 3 Nodes each 4 CPU and 15Gi Memory
- Each NATS server has 1 CPU and 2 Gi Memory
- Ordered push-based consumers

---

### Cluster stability in the long run

- 3 x (2 publishers and 10 subscribers), 3 streams, 3 replicas, file-based
- Around 12 hours

<details><summary>Commands and results</summary>

```
pods["pod1"]="./natscli bench --stream=bench1 --pub=2 --sub=10 --js --msgs=8640 --replicas=3 --size=10240 --storage=file --publishInterval=5000 bench-subj-1"
pods["pod2"]="./natscli bench --stream=bench2 --pub=2 --sub=10 --js --msgs=8640 --replicas=3 --size=10240 --storage=file --publishInterval=5000 bench-subj-2"
pods["pod3"]="./natscli bench --stream=bench3 --pub=2 --sub=10 --js --msgs=8640 --replicas=3 --size=10240 --storage=file --publishInterval=5000 bench-subj-3"
```

</details>

- Objective was to check if the consensus/replication in a cluster is stable in a long period of eventing activity.
- Did not see any server crash or consensus problem.
- Storage requirement seems to be not more than `num-of-msgs * replication-factor * msg-size`

---

### One stream for all vs one stream per `app.*.*.*`

**Publish to the same stream**

<details><summary>3 x (2 publishers and 10 subscribers), 10k messages each 10KB, 3 replicas, Sync publishers, file-based</summary>

- Avg pub msg/sec: 177
- Avg sub msg/sec: ~340

<details><summary>Commands and results</summary>

Stream must be created beforehand to set the list of subjects to `STREAM.>`: 
```
nats str add --subjects='default.>' --replicas=3 --storage=file default
```

```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app3.bench3.subj.v1"
```

```
NATS Pub/Sub stats: 3,487 msgs/sec ~ 34.06 MB/sec
 Pub stats: 317 msgs/sec ~ 3.10 MB/sec
  [1] 158 msgs/sec ~ 1.55 MB/sec (5000 msgs)
  [2] 158 msgs/sec ~ 1.55 MB/sec (5000 msgs)
  min 158 | avg 158 | max 158 | stddev 0 msgs
 Sub stats: 3,170 msgs/sec ~ 30.96 MB/sec
  [1] 317 msgs/sec ~ 3.10 MB/sec (10000 msgs)
  [10] 317 msgs/sec ~ 3.10 MB/sec (10000 msgs)
  min 317 | avg 317 | max 317 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 4,313 msgs/sec ~ 42.13 MB/sec
 Pub stats: 392 msgs/sec ~ 3.83 MB/sec
  [1] 197 msgs/sec ~ 1.93 MB/sec (5000 msgs)
  [2] 196 msgs/sec ~ 1.91 MB/sec (5000 msgs)
  min 196 | avg 196 | max 197 | stddev 0 msgs
 Sub stats: 3,922 msgs/sec ~ 38.30 MB/sec
  [1] 392 msgs/sec ~ 3.83 MB/sec (10000 msgs)
  [10] 392 msgs/sec ~ 3.83 MB/sec (10000 msgs)
  min 392 | avg 392 | max 392 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,499 msgs/sec ~ 34.18 MB/sec
 Pub stats: 318 msgs/sec ~ 3.11 MB/sec
  [1] 197 msgs/sec ~ 1.93 MB/sec (5000 msgs)
  [2] 159 msgs/sec ~ 1.55 MB/sec (5000 msgs)
  min 159 | avg 178 | max 197 | stddev 19 msgs
 Sub stats: 3,181 msgs/sec ~ 31.07 MB/sec
  [1] 318 msgs/sec ~ 3.11 MB/sec (10000 msgs)
  [10] 318 msgs/sec ~ 3.11 MB/sec (10000 msgs)
  min 318 | avg 318 | max 318 | stddev 0 msgs
```

</details>
</details>

<details><summary>3 x (2 publishers and 10 subscribers), 10k messages each 10KB, 3 replicas, Async publishers, file-based</summary>

- Avg pub msg/sec: 628
- Avg sub msg/sec: 

<details><summary>Commands and results</summary>

```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --no-progress --storage=file default.app3.bench3.subj.v1"
```

```
NATS Pub/Sub stats: 8,853 msgs/sec ~ 86.46 MB/sec
 Pub stats: 977 msgs/sec ~ 9.55 MB/sec
  [1] 592 msgs/sec ~ 5.78 MB/sec (5000 msgs)
  [2] 489 msgs/sec ~ 4.78 MB/sec (5000 msgs)
  min 489 | avg 540 | max 592 | stddev 51 msgs
 Sub stats: 8,058 msgs/sec ~ 78.70 MB/sec
  [1] 979 msgs/sec ~ 9.57 MB/sec (10000 msgs)
  [10] 806 msgs/sec ~ 7.88 MB/sec (10000 msgs)
  min 806 | avg 924 | max 979 | stddev 75 msgs
```
```
NATS Pub/Sub stats: 8,704 msgs/sec ~ 85.00 MB/sec
 Pub stats: 1,127 msgs/sec ~ 11.01 MB/sec
  [1] 583 msgs/sec ~ 5.70 MB/sec (5000 msgs)
  [2] 564 msgs/sec ~ 5.51 MB/sec (5000 msgs)
  min 564 | avg 573 | max 583 | stddev 9 msgs
 Sub stats: 7,914 msgs/sec ~ 77.29 MB/sec
  [1] 1,130 msgs/sec ~ 11.04 MB/sec (10000 msgs)
  [10] 791 msgs/sec ~ 7.73 MB/sec (10000 msgs)
  min 791 | avg 991 | max 1,130 | stddev 147 msgs
```
```
NATS Pub/Sub stats: 15,027 msgs/sec ~ 146.76 MB/sec
 Pub stats: 1,414 msgs/sec ~ 13.81 MB/sec
  [1] 835 msgs/sec ~ 8.16 MB/sec (5000 msgs)
  [2] 707 msgs/sec ~ 6.91 MB/sec (5000 msgs)
  min 707 | avg 771 | max 835 | stddev 64 msgs
 Sub stats: 13,843 msgs/sec ~ 135.19 MB/sec
  [1] 1,460 msgs/sec ~ 14.26 MB/sec (10000 msgs)
  [10] 1,442 msgs/sec ~ 14.09 MB/sec (10000 msgs)
  min 1,432 | avg 1,456 | max 1,493 | stddev 24 msgs
```
</details>
</details>

<details><summary>3 x (2 publishers and 10 subscribers), 10k messages each 1KB, 3 replicas, Async publishers, file-based</summary>

- Avg pub msg/sec: 1810
- Avg sub msg/sec:

<details><summary>Commands and results</summary>

```
NATS Pub/Sub stats: 35,289 msgs/sec ~ 34.46 MB/sec
 Pub stats: 3,320 msgs/sec ~ 3.24 MB/sec
  [1] 2,077 msgs/sec ~ 2.03 MB/sec (5000 msgs)
  [2] 1,662 msgs/sec ~ 1.62 MB/sec (5000 msgs)
  min 1,662 | avg 1,869 | max 2,077 | stddev 207 msgs
 Sub stats: 32,149 msgs/sec ~ 31.40 MB/sec
  [1] 3,331 msgs/sec ~ 3.25 MB/sec (10000 msgs)
  [10] 3,218 msgs/sec ~ 3.14 MB/sec (10000 msgs)
  min 3,218 | avg 3,274 | max 3,333 | stddev 55 msgs
```
```
NATS Pub/Sub stats: 35,610 msgs/sec ~ 34.78 MB/sec
 Pub stats: 3,608 msgs/sec ~ 3.52 MB/sec
  [1] 1,804 msgs/sec ~ 1.76 MB/sec (5000 msgs)
  [2] 1,805 msgs/sec ~ 1.76 MB/sec (5000 msgs)
  min 1,804 | avg 1,804 | max 1,805 | stddev 0 msgs
 Sub stats: 32,399 msgs/sec ~ 31.64 MB/sec
  [1] 3,616 msgs/sec ~ 3.53 MB/sec (10000 msgs)
  [10] 3,240 msgs/sec ~ 3.16 MB/sec (10000 msgs)
  min 3,240 | avg 3,324 | max 3,616 | stddev 109 msgs
```
```
NATS Pub/Sub stats: 37,846 msgs/sec ~ 36.96 MB/sec
 Pub stats: 3,440 msgs/sec ~ 3.36 MB/sec
  [1] 1,795 msgs/sec ~ 1.75 MB/sec (5000 msgs)
  [2] 1,722 msgs/sec ~ 1.68 MB/sec (5000 msgs)
  **min 1,722 | avg 1,758 | max 1,795 | stddev 36 msgs**
 Sub stats: 35,486 msgs/sec ~ 34.65 MB/sec
  [1] 3,550 msgs/sec ~ 3.47 MB/sec (10000 msgs)
  [10] 3,670 msgs/sec ~ 3.58 MB/sec (10000 msgs)
  min 3,550 | avg 3,612 | max 3,670 | stddev 54 msgs
```
</details>
</details>

<details><summary>6 x (2 publishers and 10 subscribers), 10k messages each 10KB, 3 replicas, Sync publishers, file-based</summary>

- Avg pub msg/sec: 97
- Avg sub msg/sec: ~175

<details><summary>Commands and results</summary>

```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file default.app6.bench3.subj.v1"
```

```
NATS Pub/Sub stats: 1,944 msgs/sec ~ 18.99 MB/sec
 Pub stats: 176 msgs/sec ~ 1.73 MB/sec
  [1] 108 msgs/sec ~ 1.06 MB/sec (5000 msgs)
  [2] 88 msgs/sec ~ 884.08 KB/sec (5000 msgs)
  min 88 | avg 98 | max 108 | stddev 10 msgs
 Sub stats: 1,768 msgs/sec ~ 17.27 MB/sec
  min 176 | avg 176 | max 176 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 1,862 msgs/sec ~ 18.19 MB/sec
 Pub stats: 169 msgs/sec ~ 1.65 MB/sec
  [1] 104 msgs/sec ~ 1.02 MB/sec (5000 msgs)
  [2] 84 msgs/sec ~ 846.53 KB/sec (5000 msgs)
  min 84 | avg 94 | max 104 | stddev 10 msgs
 Sub stats: 1,693 msgs/sec ~ 16.53 MB/sec
  min 169 | avg 169 | max 169 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 1,932 msgs/sec ~ 18.87 MB/sec
 Pub stats: 175 msgs/sec ~ 1.72 MB/sec
  [1] 97 msgs/sec ~ 971.05 KB/sec (5000 msgs)
  [2] 87 msgs/sec ~ 878.22 KB/sec (5000 msgs)
  min 87 | avg 92 | max 97 | stddev 5 msgs
 Sub stats: 1,756 msgs/sec ~ 17.16 MB/sec
  min 175 | avg 175 | max 175 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 2,321 msgs/sec ~ 22.67 MB/sec
 Pub stats: 211 msgs/sec ~ 2.06 MB/sec
  [1] 106 msgs/sec ~ 1.04 MB/sec (5000 msgs)
  [2] 105 msgs/sec ~ 1.03 MB/sec (5000 msgs)
  min 105 | avg 105 | max 106 | stddev 0 msgs
 Sub stats: 2,112 msgs/sec ~ 20.63 MB/sec
  min 211 | avg 211 | max 211 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 2,235 msgs/sec ~ 21.84 MB/sec
 Pub stats: 203 msgs/sec ~ 1.99 MB/sec
  [1] 102 msgs/sec ~ 1020.03 KB/sec (5000 msgs)
  [2] 101 msgs/sec ~ 1016.43 KB/sec (5000 msgs)
  min 101 | avg 101 | max 102 | stddev 0 msgs
 Sub stats: 2,032 msgs/sec ~ 19.85 MB/sec
  min 203 | avg 203 | max 203 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 2,083 msgs/sec ~ 20.34 MB/sec
 Pub stats: 189 msgs/sec ~ 1.85 MB/sec
  [1] 95 msgs/sec ~ 953.62 KB/sec (5000 msgs)
  [2] 94 msgs/sec ~ 947.04 KB/sec (5000 msgs)
  min 94 | avg 94 | max 95 | stddev 0 msgs
 Sub stats: 1,893 msgs/sec ~ 18.50 MB/sec
  min 189 | avg 189 | max 189 | stddev 0 msgs
```

</details>
</details>

<details><summary>6 x (2 publishers and 10 subscribers), 1M messages each 512B, 3 replicas, Async publishers, file-based</summary>

- Avg pub msg/sec: 387
- Avg sub msg/sec:

<details><summary>Commands and results</summary>

```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app6.bench3.subj.v1"
```
```
*****Pod1*****
--- PUB ---
  [2] 393 msgs/sec ~ 196.92 KB/sec (500000 msgs)
  min 393 | avg 393 | max 393 | stddev 0 msgs
--- SUB ---
  [10] 787 msgs/sec ~ 393.83 KB/sec (1000000 msgs)
  min 787 | avg 787 | max 787 | stddev 0 msgs
*****Pod2*****
--- PUB ---
  [2] 382 msgs/sec ~ 191.12 KB/sec (500000 msgs)
  min 382 | avg 382 | max 382 | stddev 0 msgs
--- SUB ---
  [10] 764 msgs/sec ~ 382.25 KB/sec (1000000 msgs)
  min 764 | avg 764 | max 764 | stddev 0 msgs
*****Pod3*****
--- PUB ---
  [2] 392 msgs/sec ~ 196.37 KB/sec (500000 msgs)
  min 392 | avg 392 | max 393 | stddev 0 msgs
--- SUB ---
  [10] 785 msgs/sec ~ 392.77 KB/sec (1000000 msgs)
  min 785 | avg 785 | max 785 | stddev 0 msgs
*****Pod4*****
--- PUB ---
  [2] 382 msgs/sec ~ 191.47 KB/sec (500000 msgs)
  min 382 | avg 387 | max 392 | stddev 5 msgs
--- SUB ---
  [10] 765 msgs/sec ~ 382.94 KB/sec (1000000 msgs)
  min 765 | avg 765 | max 765 | stddev 0 msgs
*****Pod5*****
--- PUB ---
  [2] 382 msgs/sec ~ 191.27 KB/sec (500000 msgs)
  min 382 | avg 387 | max 393 | stddev 5 msgs
--- SUB ---
  [10] 765 msgs/sec ~ 382.54 KB/sec (1000000 msgs)
  min 765 | avg 765 | max 765 | stddev 0 msgs
*****Pod6*****
--- PUB ---
  [2] 382 msgs/sec ~ 191.32 KB/sec (500000 msgs)
  min 382 | avg 382 | max 382 | stddev 0 msgs
--- SUB ---
  [10] 765 msgs/sec ~ 382.67 KB/sec (1000000 msgs)
  min 765 | avg 765 | max 765 | stddev 0 msgs
```

</details>

</details>

**Publish to one stream per `app.*.*.*`**

<details><summary>3 x (2 publishers and 10 subscribers), 10k messages each 10KB, 3 replicas, Sync publishers</summary>

- Avg pub msg/sec: 210
- Avg sub msg/sec: ~360

<details><summary>Commands and results</summary>

Create the three streams beforehand: app1, app2, app3, e.g.:
`nats str add --subjects='app1.>' --replicas=3 --storage=file app1`

```
pods["pod1"]="./natscli bench --stream=app1 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=app2 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=app3 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app3.bench3.subj.v1"
```

```
NATS Pub/Sub stats: 4,309 msgs/sec ~ 42.09 MB/sec
 Pub stats: 391 msgs/sec ~ 3.83 MB/sec
  [1] 280 msgs/sec ~ 2.74 MB/sec (5000 msgs)
  [2] 195 msgs/sec ~ 1.91 MB/sec (5000 msgs)
  min 195 | avg 237 | max 280 | stddev 42 msgs
 Sub stats: 3,918 msgs/sec ~ 38.26 MB/sec
  [1] 391 msgs/sec ~ 3.83 MB/sec (10000 msgs)
  [10] 391 msgs/sec ~ 3.83 MB/sec (10000 msgs)
  min 391 | avg 391 | max 391 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,756 msgs/sec ~ 36.68 MB/sec
 Pub stats: 341 msgs/sec ~ 3.33 MB/sec
  [1] 231 msgs/sec ~ 2.26 MB/sec (5000 msgs)
  [2] 170 msgs/sec ~ 1.67 MB/sec (5000 msgs)
  min 170 | avg 200 | max 231 | stddev 30 msgs
 Sub stats: 3,415 msgs/sec ~ 33.35 MB/sec
  [1] 341 msgs/sec ~ 3.34 MB/sec (10000 msgs)
  [10] 341 msgs/sec ~ 3.34 MB/sec (10000 msgs)
  min 341 | avg 341 | max 341 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,969 msgs/sec ~ 38.76 MB/sec
 Pub stats: 360 msgs/sec ~ 3.52 MB/sec
  [1] 211 msgs/sec ~ 2.07 MB/sec (5000 msgs)
  [2] 180 msgs/sec ~ 1.76 MB/sec (5000 msgs)
  min 180 | avg 195 | max 211 | stddev 15 msgs
 Sub stats: 3,608 msgs/sec ~ 35.24 MB/sec
  [1] 360 msgs/sec ~ 3.52 MB/sec (10000 msgs)
  [10] 360 msgs/sec ~ 3.52 MB/sec (10000 msgs)
  min 360 | avg 360 | max 360 | stddev 0 msgs
```

</details>
</details>

<details><summary>6 x (2 publishers and 10 subscribers), 10k messages each 10KB, 3 replicas, Sync publishers, file-based</summary>

- Avg pub msg/sec: 173
- Avg sub msg/sec: ~350

<details><summary>Commands and results</summary>

```
pods["pod1"]="./natscli bench --stream=app1 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=app2 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=app3 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=app4 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=app5 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=app6 --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=10240 --syncpub --no-progress --storage=file app6.bench3.subj.v1"
```

```
NATS Pub/Sub stats: 3,005 msgs/sec ~ 29.35 MB/sec
 Pub stats: 273 msgs/sec ~ 2.67 MB/sec
  [1] 136 msgs/sec ~ 1.33 MB/sec (5000 msgs)
  [2] 136 msgs/sec ~ 1.33 MB/sec (5000 msgs)
  min 136 | avg 136 | max 136 | stddev 0 msgs
 Sub stats: 2,732 msgs/sec ~ 26.68 MB/sec
  min 273 | avg 273 | max 273 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 7,091 msgs/sec ~ 69.26 MB/sec
 Pub stats: 644 msgs/sec ~ 6.30 MB/sec
  [1] 324 msgs/sec ~ 3.17 MB/sec (5000 msgs)
  [2] 322 msgs/sec ~ 3.15 MB/sec (5000 msgs)
  min 322 | avg 323 | max 324 | stddev 1 msgs
 Sub stats: 6,448 msgs/sec ~ 62.97 MB/sec
  min 644 | avg 644 | max 644 | stddev 0 msgs
```
```

NATS Pub/Sub stats: 2,423 msgs/sec ~ 23.66 MB/sec
 Pub stats: 220 msgs/sec ~ 2.15 MB/sec
  [1] 147 msgs/sec ~ 1.44 MB/sec (5000 msgs)
  [2] 110 msgs/sec ~ 1.08 MB/sec (5000 msgs)
  min 110 | avg 128 | max 147 | stddev 18 msgs
 Sub stats: 2,203 msgs/sec ~ 21.51 MB/sec
  [1] 220 msgs/sec ~ 2.15 MB/sec (10000 msgs)
  [10] 220 msgs/sec ~ 2.15 MB/sec (10000 msgs)
  min 220 | avg 220 | max 220 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 2,495 msgs/sec ~ 24.37 MB/sec
 Pub stats: 226 msgs/sec ~ 2.22 MB/sec
  [1] 128 msgs/sec ~ 1.26 MB/sec (5000 msgs)
  [2] 113 msgs/sec ~ 1.11 MB/sec (5000 msgs)
  min 113 | avg 120 | max 128 | stddev 7 msgs
 Sub stats: 2,269 msgs/sec ~ 22.16 MB/sec
  [1] 226 msgs/sec ~ 2.22 MB/sec (10000 msgs)
  [10] 226 msgs/sec ~ 2.22 MB/sec (10000 msgs)
  min 226 | avg 226 | max 226 | stddev 0 msgs
```
```

NATS Pub/Sub stats: 2,927 msgs/sec ~ 28.58 MB/sec
 Pub stats: 266 msgs/sec ~ 2.60 MB/sec
  [1] 133 msgs/sec ~ 1.30 MB/sec (5000 msgs)
  [2] 133 msgs/sec ~ 1.30 MB/sec (5000 msgs)
  min 133 | avg 133 | max 133 | stddev 0 msgs
 Sub stats: 2,661 msgs/sec ~ 25.99 MB/sec
  [1] 266 msgs/sec ~ 2.60 MB/sec (10000 msgs)
  [10] 266 msgs/sec ~ 2.60 MB/sec (10000 msgs)
  min 266 | avg 266 | max 266 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 4,337 msgs/sec ~ 42.36 MB/sec
 Pub stats: 394 msgs/sec ~ 3.85 MB/sec
  [1] 199 msgs/sec ~ 1.95 MB/sec (5000 msgs)
  [2] 197 msgs/sec ~ 1.93 MB/sec (5000 msgs)
  min 197 | avg 198 | max 199 | stddev 1 msgs
 Sub stats: 3,943 msgs/sec ~ 38.51 MB/sec
  [1] 394 msgs/sec ~ 3.85 MB/sec (10000 msgs)
  [10] 394 msgs/sec ~ 3.86 MB/sec (10000 msgs)
  min 394 | avg 394 | max 394 | stddev 0 msgs
```

</details>
</details>

---

<details><summary>What could this mean?</summary>

- It seems sharing the same stream reduces the throughput of publishers. When increasing the pods, 
  sharing the stream seems to reduce performance more than using multiple streams.
- We should aim for using async publishing in the publisher proxy.
- Depending on what event rate we should provide, we might be able to get away with one Stream for all event types 
  if we use async publishing.
- Using only one stream simplifies the code in the publisher proxy.
- We should agree on some realistic parameters (e.g. retention period, msg size) and re-run this evaluation.

</details>

---

### Memory vs file-based

<details><summary>6 x (2 publishers and 10 subscribers), 10k messages each 1KB, 3 replicas, Sync publishers, 1 stream</summary> 

**In-memory**

- Avg pub msg/sec: 197
- Avg sub msg/sec: ~360

<details><summary>Commands and results</summary>

Create the stream:
`nats str add --subjects='default.>' --replicas=3 --storage=memory default`

```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=memory default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=memory default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=memory default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=memory default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=memory default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=memory default.app6.bench3.subj.v1"
```

```
NATS Pub/Sub stats: 3,921 msgs/sec ~ 3.83 MB/sec
 Pub stats: 356 msgs/sec ~ 356.47 KB/sec
  [1] 236 msgs/sec ~ 236.42 KB/sec (5000 msgs)
  [2] 178 msgs/sec ~ 178.24 KB/sec (5000 msgs)
  min 178 | avg 207 | max 236 | stddev 29 msgs
 Sub stats: 3,564 msgs/sec ~ 3.48 MB/sec
  [1] 356 msgs/sec ~ 356.49 KB/sec (10000 msgs)
  [10] 356 msgs/sec ~ 356.49 KB/sec (10000 msgs)
  min 356 | avg 356 | max 356 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,962 msgs/sec ~ 3.87 MB/sec
 Pub stats: 360 msgs/sec ~ 360.26 KB/sec
  [1] 235 msgs/sec ~ 235.65 KB/sec (5000 msgs)
  [2] 180 msgs/sec ~ 180.13 KB/sec (5000 msgs)
  min 180 | avg 207 | max 235 | stddev 27 msgs
 Sub stats: 3,602 msgs/sec ~ 3.52 MB/sec
  [1] 360 msgs/sec ~ 360.29 KB/sec (10000 msgs)
  [10] 360 msgs/sec ~ 360.28 KB/sec (10000 msgs)
  min 360 | avg 360 | max 360 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,803 msgs/sec ~ 3.71 MB/sec
 Pub stats: 345 msgs/sec ~ 345.79 KB/sec
  [1] 215 msgs/sec ~ 215.72 KB/sec (5000 msgs)
  [2] 172 msgs/sec ~ 172.89 KB/sec (5000 msgs)
  min 172 | avg 193 | max 215 | stddev 21 msgs
 Sub stats: 3,458 msgs/sec ~ 3.38 MB/sec
  [1] 345 msgs/sec ~ 345.84 KB/sec (10000 msgs)
  [10] 345 msgs/sec ~ 345.83 KB/sec (10000 msgs)
  min 345 | avg 345 | max 345 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 4,181 msgs/sec ~ 4.08 MB/sec
 Pub stats: 380 msgs/sec ~ 380.17 KB/sec
  [1] 235 msgs/sec ~ 235.59 KB/sec (5000 msgs)
  [2] 190 msgs/sec ~ 190.12 KB/sec (5000 msgs)
  min 190 | avg 212 | max 235 | stddev 22 msgs
 Sub stats: 3,802 msgs/sec ~ 3.71 MB/sec
  [1] 380 msgs/sec ~ 380.21 KB/sec (10000 msgs)
  [10] 380 msgs/sec ~ 380.21 KB/sec (10000 msgs)
  min 380 | avg 380 | max 380 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,939 msgs/sec ~ 3.85 MB/sec
 Pub stats: 358 msgs/sec ~ 358.13 KB/sec
  [1] 223 msgs/sec ~ 223.34 KB/sec (5000 msgs)
  [2] 179 msgs/sec ~ 179.06 KB/sec (5000 msgs)
  min 179 | avg 201 | max 223 | stddev 22 msgs
 Sub stats: 3,581 msgs/sec ~ 3.50 MB/sec
  [1] 358 msgs/sec ~ 358.15 KB/sec (10000 msgs)
  [10] 358 msgs/sec ~ 358.16 KB/sec (10000 msgs)
  min 358 | avg 358 | max 358 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,632 msgs/sec ~ 3.55 MB/sec
 Pub stats: 330 msgs/sec ~ 330.26 KB/sec
  [1] 165 msgs/sec ~ 165.22 KB/sec (5000 msgs)
  [2] 165 msgs/sec ~ 165.13 KB/sec (5000 msgs)
  min 165 | avg 165 | max 165 | stddev 0 msgs
 Sub stats: 3,302 msgs/sec ~ 3.23 MB/sec
  [1] 330 msgs/sec ~ 330.83 KB/sec (10000 msgs)
  [10] 330 msgs/sec ~ 330.28 KB/sec (10000 msgs)
  min 330 | avg 330 | max 330 | stddev 0 msgs
```
</details>

**File-based**

- Avg pub msg/sec: 164
- Avg sub msg/sec: ~330

<details><summary>Commands and results</summary>

Create the stream:
`nats str add --subjects='default.>' --replicas=3 --storage=file default`

Run the Pods:
```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app6.bench3.subj.v1"
```

Results:
```
NATS Pub/Sub stats: 4,079 msgs/sec ~ 3.98 MB/sec
 Pub stats: 370 msgs/sec ~ 370.86 KB/sec
  [1] 186 msgs/sec ~ 186.02 KB/sec (5000 msgs)
  [2] 185 msgs/sec ~ 185.69 KB/sec (5000 msgs)
  min 185 | avg 185 | max 186 | stddev 0 msgs
 Sub stats: 3,708 msgs/sec ~ 3.62 MB/sec
  [1] 370 msgs/sec ~ 370.90 KB/sec (10000 msgs)
  [10] 370 msgs/sec ~ 370.89 KB/sec (10000 msgs)
  min 370 | avg 370 | max 371 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,323 msgs/sec ~ 3.25 MB/sec
 Pub stats: 302 msgs/sec ~ 302.11 KB/sec
  [1] 152 msgs/sec ~ 152.41 KB/sec (5000 msgs)
  [2] 151 msgs/sec ~ 151.05 KB/sec (5000 msgs)
  min 151 | avg 151 | max 152 | stddev 0 msgs
 Sub stats: 3,021 msgs/sec ~ 2.95 MB/sec
  [1] 302 msgs/sec ~ 302.13 KB/sec (10000 msgs)
  [10] 302 msgs/sec ~ 302.13 KB/sec (10000 msgs)
  min 302 | avg 302 | max 302 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,290 msgs/sec ~ 3.21 MB/sec
 Pub stats: 299 msgs/sec ~ 299.17 KB/sec
  [1] 185 msgs/sec ~ 185.26 KB/sec (5000 msgs)
  [2] 149 msgs/sec ~ 149.61 KB/sec (5000 msgs)
  min 149 | avg 167 | max 185 | stddev 18 msgs
 Sub stats: 2,991 msgs/sec ~ 2.92 MB/sec
  [1] 299 msgs/sec ~ 299.20 KB/sec (10000 msgs)
  [10] 299 msgs/sec ~ 299.19 KB/sec (10000 msgs)
  min 299 | avg 299 | max 299 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 4,090 msgs/sec ~ 3.99 MB/sec
 Pub stats: 371 msgs/sec ~ 371.83 KB/sec
  [1] 185 msgs/sec ~ 185.99 KB/sec (5000 msgs)
  [2] 186 msgs/sec ~ 186.35 KB/sec (5000 msgs)
  min 185 | avg 185 | max 186 | stddev 0 msgs
 Sub stats: 3,726 msgs/sec ~ 3.64 MB/sec
  [1] 372 msgs/sec ~ 372.72 KB/sec (10000 msgs)
  [10] 372 msgs/sec ~ 372.70 KB/sec (10000 msgs)
  min 372 | avg 372 | max 372 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,256 msgs/sec ~ 3.18 MB/sec
 Pub stats: 296 msgs/sec ~ 296.09 KB/sec
  [1] 148 msgs/sec ~ 148.06 KB/sec (5000 msgs)
  [2] 148 msgs/sec ~ 148.04 KB/sec (5000 msgs)
  min 148 | avg 148 | max 148 | stddev 0 msgs
 Sub stats: 2,961 msgs/sec ~ 2.89 MB/sec
  [1] 296 msgs/sec ~ 296.13 KB/sec (10000 msgs)
  [10] 296 msgs/sec ~ 296.12 KB/sec (10000 msgs)
  min 296 | avg 296 | max 296 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,194 msgs/sec ~ 3.12 MB/sec
 Pub stats: 290 msgs/sec ~ 290.45 KB/sec
  [1] 152 msgs/sec ~ 152.78 KB/sec (5000 msgs)
  [2] 145 msgs/sec ~ 145.23 KB/sec (5000 msgs)
  min 145 | avg 148 | max 152 | stddev 3 msgs
 Sub stats: 2,904 msgs/sec ~ 2.84 MB/sec
  [1] 290 msgs/sec ~ 290.48 KB/sec (10000 msgs)
  [10] 290 msgs/sec ~ 290.48 KB/sec (10000 msgs)
  min 290 | avg 290 | max 290 | stddev 0 msgs
```

</details>

</details>

---

<details><summary>What could this mean?</summary>

- File-based storage doesn't seem to be much slower than in-memory. Maybe due to the replication latency.
- We'd any way probably use file-based!

</details>

---

### Replication

**No replication**

<details><summary>6 x (2 publishers and 10 subscribers), 10k messages each 1KB, file-based, Sync publishers, 1 stream</summary>

- Avg pub msg/sec: 173
- Avg sub msg/sec: ~320

<details><summary>Commands and results</summary>

Create the stream:
`nats str add --subjects='default.>' --replicas=1 --storage=file default`

Run the Pods:
```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=1 --size=1024 --syncpub --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=1 --size=1024 --syncpub --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=1 --size=1024 --syncpub --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=1 --size=1024 --syncpub --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=1 --size=1024 --syncpub --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=1 --size=1024 --syncpub --no-progress --storage=file default.app6.bench3.subj.v1"
```

Results:
```
NATS Pub/Sub stats: 3,351 msgs/sec ~ 3.27 MB/sec
 Pub stats: 304 msgs/sec ~ 304.72 KB/sec
  [1] 209 msgs/sec ~ 209.03 KB/sec (5000 msgs)
  [2] 152 msgs/sec ~ 152.37 KB/sec (5000 msgs)
  min 152 | avg 180 | max 209 | stddev 28 msgs
 Sub stats: 3,047 msgs/sec ~ 2.98 MB/sec
  [1] 304 msgs/sec ~ 304.73 KB/sec (10000 msgs)
  [10] 304 msgs/sec ~ 304.72 KB/sec (10000 msgs)
  min 304 | avg 304 | max 304 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,348 msgs/sec ~ 3.27 MB/sec
 Pub stats: 304 msgs/sec ~ 304.44 KB/sec
  [1] 152 msgs/sec ~ 152.64 KB/sec (5000 msgs)
  [2] 152 msgs/sec ~ 152.23 KB/sec (5000 msgs)
  min 152 | avg 152 | max 152 | stddev 0 msgs
 Sub stats: 3,044 msgs/sec ~ 2.97 MB/sec
  [1] 304 msgs/sec ~ 304.46 KB/sec (10000 msgs)
  [10] 304 msgs/sec ~ 304.44 KB/sec (10000 msgs)
  min 304 | avg 304 | max 304 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,343 msgs/sec ~ 3.27 MB/sec
 Pub stats: 303 msgs/sec ~ 303.98 KB/sec
  [1] 197 msgs/sec ~ 197.89 KB/sec (5000 msgs)
  [2] 151 msgs/sec ~ 151.99 KB/sec (5000 msgs)
  min 151 | avg 174 | max 197 | stddev 23 msgs
 Sub stats: 3,040 msgs/sec ~ 2.97 MB/sec
  [1] 304 msgs/sec ~ 304.01 KB/sec (10000 msgs)
  [10] 304 msgs/sec ~ 304.03 KB/sec (10000 msgs)
  min 304 | avg 304 | max 304 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,386 msgs/sec ~ 3.31 MB/sec
 Pub stats: 307 msgs/sec ~ 307.87 KB/sec
  [1] 155 msgs/sec ~ 155.20 KB/sec (5000 msgs)
  [2] 153 msgs/sec ~ 153.93 KB/sec (5000 msgs)
  min 153 | avg 154 | max 155 | stddev 1 msgs
 Sub stats: 3,078 msgs/sec ~ 3.01 MB/sec
  [1] 307 msgs/sec ~ 307.91 KB/sec (10000 msgs)
  [10] 307 msgs/sec ~ 307.90 KB/sec (10000 msgs)
  min 307 | avg 307 | max 307 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 4,532 msgs/sec ~ 4.43 MB/sec
 Pub stats: 412 msgs/sec ~ 412.06 KB/sec
  [1] 206 msgs/sec ~ 206.63 KB/sec (5000 msgs)
  [2] 206 msgs/sec ~ 206.05 KB/sec (5000 msgs)
  min 206 | avg 206 | max 206 | stddev 0 msgs
 Sub stats: 4,120 msgs/sec ~ 4.02 MB/sec
  [1] 412 msgs/sec ~ 412.09 KB/sec (10000 msgs)
  [10] 412 msgs/sec ~ 412.10 KB/sec (10000 msgs)
  min 412 | avg 412 | max 412 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,293 msgs/sec ~ 3.22 MB/sec
 Pub stats: 299 msgs/sec ~ 299.40 KB/sec
  [1] 196 msgs/sec ~ 196.35 KB/sec (5000 msgs)
  [2] 149 msgs/sec ~ 149.72 KB/sec (5000 msgs)
  min 149 | avg 172 | max 196 | stddev 23 msgs
 Sub stats: 2,994 msgs/sec ~ 2.92 MB/sec
  [1] 299 msgs/sec ~ 299.42 KB/sec (10000 msgs)
  [10] 299 msgs/sec ~ 299.42 KB/sec (10000 msgs)
  min 299 | avg 299 | max 299 | stddev 0 msgs
```

</details>
</details>

<details><summary>6 x (2 publishers and 10 subscribers), 1M messages each 512B, file-based, Async publishers, 1 stream</summary>

- Avg pub msg/sec: 332
- Avg sub msg/sec:

<details><summary>Commands and results</summary>

```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=1 --size=512 --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=1 --size=512 --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=1 --size=512 --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=1 --size=512 --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=1 --size=512 --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=1 --size=512 --no-progress --storage=file default.app6.bench3.subj.v1"
```
```
*****Pod1*****
--- PUB ---
  [2] 326 msgs/sec ~ 163.05 KB/sec (500000 msgs)
  min 326 | avg 326 | max 326 | stddev 0 msgs
--- SUB ---
  [10] 652 msgs/sec ~ 326.12 KB/sec (1000000 msgs)
  min 652 | avg 652 | max 652 | stddev 0 msgs
*****Pod2*****
--- PUB ---
  [2] 326 msgs/sec ~ 163.03 KB/sec (500000 msgs)
  min 326 | avg 335 | max 344 | stddev 9 msgs
--- SUB ---
  [10] 652 msgs/sec ~ 326.06 KB/sec (1000000 msgs)
  min 652 | avg 652 | max 652 | stddev 0 msgs
*****Pod3*****
--- PUB ---
  [2] 325 msgs/sec ~ 162.77 KB/sec (500000 msgs)
  min 325 | avg 325 | max 325 | stddev 0 msgs
--- SUB ---
  [10] 651 msgs/sec ~ 325.58 KB/sec (1000000 msgs)
  min 651 | avg 651 | max 651 | stddev 0 msgs
*****Pod4*****
--- PUB ---
  [2] 327 msgs/sec ~ 163.85 KB/sec (500000 msgs)
  min 327 | avg 332 | max 338 | stddev 5 msgs
--- SUB ---
  [10] 655 msgs/sec ~ 327.72 KB/sec (1000000 msgs)
  min 655 | avg 655 | max 655 | stddev 0 msgs
*****Pod5*****
--- PUB ---
  [2] 338 msgs/sec ~ 169.16 KB/sec (500000 msgs)
  min 338 | avg 339 | max 340 | stddev 1 msgs
--- SUB ---
  [10] 676 msgs/sec ~ 338.35 KB/sec (1000000 msgs)
  min 676 | avg 676 | max 676 | stddev 0 msgs
*****Pod6*****
--- PUB ---
  [2] 337 msgs/sec ~ 168.92 KB/sec (500000 msgs)
  min 337 | avg 337 | max 338 | stddev 0 msgs
--- SUB ---
  [10] 675 msgs/sec ~ 337.85 KB/sec (1000000 msgs)
  min 675 | avg 675 | max 675 | stddev 0 msgs
```

</details>
</details>

**Replication = 3**

<details><summary>6 x (2 publishers and 10 subscribers), 10k messages each 1KB, file-based, Sync publishers, 1 stream</summary>

- Avg pub msg/sec: 164
- Avg sub msg/sec: ~330

<details><summary>Commands and results</summary>

Create the stream:
`nats str add --subjects='default.>' --replicas=3 --storage=file default`

Run the Pods:
```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=10000 --replicas=3 --size=1024 --syncpub --no-progress --storage=file default.app6.bench3.subj.v1"
```

Results:
```
NATS Pub/Sub stats: 4,079 msgs/sec ~ 3.98 MB/sec
 Pub stats: 370 msgs/sec ~ 370.86 KB/sec
  [1] 186 msgs/sec ~ 186.02 KB/sec (5000 msgs)
  [2] 185 msgs/sec ~ 185.69 KB/sec (5000 msgs)
  min 185 | avg 185 | max 186 | stddev 0 msgs
 Sub stats: 3,708 msgs/sec ~ 3.62 MB/sec
  [1] 370 msgs/sec ~ 370.90 KB/sec (10000 msgs)
  [10] 370 msgs/sec ~ 370.89 KB/sec (10000 msgs)
  min 370 | avg 370 | max 371 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,323 msgs/sec ~ 3.25 MB/sec
 Pub stats: 302 msgs/sec ~ 302.11 KB/sec
  [1] 152 msgs/sec ~ 152.41 KB/sec (5000 msgs)
  [2] 151 msgs/sec ~ 151.05 KB/sec (5000 msgs)
  min 151 | avg 151 | max 152 | stddev 0 msgs
 Sub stats: 3,021 msgs/sec ~ 2.95 MB/sec
  [1] 302 msgs/sec ~ 302.13 KB/sec (10000 msgs)
  [10] 302 msgs/sec ~ 302.13 KB/sec (10000 msgs)
  min 302 | avg 302 | max 302 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,290 msgs/sec ~ 3.21 MB/sec
 Pub stats: 299 msgs/sec ~ 299.17 KB/sec
  [1] 185 msgs/sec ~ 185.26 KB/sec (5000 msgs)
  [2] 149 msgs/sec ~ 149.61 KB/sec (5000 msgs)
  min 149 | avg 167 | max 185 | stddev 18 msgs
 Sub stats: 2,991 msgs/sec ~ 2.92 MB/sec
  [1] 299 msgs/sec ~ 299.20 KB/sec (10000 msgs)
  [10] 299 msgs/sec ~ 299.19 KB/sec (10000 msgs)
  min 299 | avg 299 | max 299 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 4,090 msgs/sec ~ 3.99 MB/sec
 Pub stats: 371 msgs/sec ~ 371.83 KB/sec
  [1] 185 msgs/sec ~ 185.99 KB/sec (5000 msgs)
  [2] 186 msgs/sec ~ 186.35 KB/sec (5000 msgs)
  min 185 | avg 185 | max 186 | stddev 0 msgs
 Sub stats: 3,726 msgs/sec ~ 3.64 MB/sec
  [1] 372 msgs/sec ~ 372.72 KB/sec (10000 msgs)
  [10] 372 msgs/sec ~ 372.70 KB/sec (10000 msgs)
  min 372 | avg 372 | max 372 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,256 msgs/sec ~ 3.18 MB/sec
 Pub stats: 296 msgs/sec ~ 296.09 KB/sec
  [1] 148 msgs/sec ~ 148.06 KB/sec (5000 msgs)
  [2] 148 msgs/sec ~ 148.04 KB/sec (5000 msgs)
  min 148 | avg 148 | max 148 | stddev 0 msgs
 Sub stats: 2,961 msgs/sec ~ 2.89 MB/sec
  [1] 296 msgs/sec ~ 296.13 KB/sec (10000 msgs)
  [10] 296 msgs/sec ~ 296.12 KB/sec (10000 msgs)
  min 296 | avg 296 | max 296 | stddev 0 msgs
```
```
NATS Pub/Sub stats: 3,194 msgs/sec ~ 3.12 MB/sec
 Pub stats: 290 msgs/sec ~ 290.45 KB/sec
  [1] 152 msgs/sec ~ 152.78 KB/sec (5000 msgs)
  [2] 145 msgs/sec ~ 145.23 KB/sec (5000 msgs)
  min 145 | avg 148 | max 152 | stddev 3 msgs
 Sub stats: 2,904 msgs/sec ~ 2.84 MB/sec
  [1] 290 msgs/sec ~ 290.48 KB/sec (10000 msgs)
  [10] 290 msgs/sec ~ 290.48 KB/sec (10000 msgs)
  min 290 | avg 290 | max 290 | stddev 0 msgs
```

</details>
</details>

<details><summary>6 x (2 publishers and 10 subscribers), 1M messages each 512B, file-based, Async publishers, 1 stream</summary>

- Avg pub msg/sec: 408
- Avg sub msg/sec:

<details><summary>Commands and results</summary>

```
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512 --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512 --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512 --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512 --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512 --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512 --no-progress --storage=file default.app6.bench3.subj.v1"
```
```
*****Pod1*****
--- PUB ---
  [2] 399 msgs/sec ~ 199.89 KB/sec (500000 msgs)
  min 399 | avg 399 | max 399 | stddev 0 msgs
--- SUB ---
  [10] 799 msgs/sec ~ 399.77 KB/sec (1000000 msgs)
  min 799 | avg 799 | max 799 | stddev 0 msgs
*****Pod2*****
--- PUB ---
  [2] 401 msgs/sec ~ 200.91 KB/sec (500000 msgs)
  min 401 | avg 412 | max 423 | stddev 11 msgs
--- SUB ---
  [10] 803 msgs/sec ~ 401.83 KB/sec (1000000 msgs)
  min 803 | avg 803 | max 803 | stddev 0 msgs
*****Pod3*****
--- PUB ---
  [2] 418 msgs/sec ~ 209.15 KB/sec (500000 msgs)
  min 418 | avg 418 | max 418 | stddev 0 msgs
--- SUB ---
  [10] 836 msgs/sec ~ 418.38 KB/sec (1000000 msgs)
  min 836 | avg 836 | max 836 | stddev 0 msgs
*****Pod4*****
--- PUB ---
  [2] 400 msgs/sec ~ 200.21 KB/sec (500000 msgs)
  min 400 | avg 410 | max 420 | stddev 10 msgs
--- SUB ---
  [10] 800 msgs/sec ~ 400.47 KB/sec (1000000 msgs)
  min 800 | avg 800 | max 800 | stddev 0 msgs
*****Pod5*****
--- PUB ---
  [2] 401 msgs/sec ~ 200.72 KB/sec (500000 msgs)
  min 401 | avg 410 | max 420 | stddev 9 msgs
--- SUB ---
  [10] 802 msgs/sec ~ 401.45 KB/sec (1000000 msgs)
  min 802 | avg 802 | max 802 | stddev 0 msgs
*****Pod6*****
--- PUB ---
  [2] 400 msgs/sec ~ 200.21 KB/sec (500000 msgs)
  min 400 | avg 402 | max 404 | stddev 2 msgs
--- SUB ---
  [10] 800 msgs/sec ~ 400.46 KB/sec (1000000 msgs)
  min 800 | avg 800 | max 800 | stddev 0 msgs
```

</details>
</details>

---

<details><summary>What could this mean?</summary>

- Replication doesn't seem to have a huge impact on the throughput.
- Results seem a bit unstable!

</details>

---

## Phase 2

When moving to JetStream, what extra features/options can we provide as part of the NATS-based backend?

- What kind of ack models should we support? 
- Delivery policy: should we allow receiving messages form the start of the topic?
- FlowControl: enforcing max-inflight (or max-ack-pending)
- JS supports different accounts on the same cluster, to support multi-tenancy on the same JS cluster. 
  Do we need this? Probably not!
- We might want to look into the KV store that JetStream offers, which might allow delegating 
  state of our dispatchers (if we have any) to JetStream.
- JetStream offers a deduplication feature (per window of time, e.g. 2m), which allows providing exactly-once 
  delivery guarantee.
  