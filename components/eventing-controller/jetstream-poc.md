### Phase 1 

What are the minimum change we need to have to keep our current NATS-based backend but move to JetStream to be able to 
have at-least-once guarantees?

- [F] Should we use Operator or Helm? How do we setup a cluster of size 1 or 3, that can use a PV?
  
- [F,P] How to configure streams and consumers?
  
- [F] What changes are required in our NATS-based reconciler/dispatcher to use the JetStream backend?
  + How many consumers do we need? One per subscription? 
  + Pull or push-based? Push-based seems to be the closest to the current model we have with NATS.
  
- [F,P] How to migrate a cluster from a NATS-based to a JetStream-based backend, and how would that impact existing subscriptions?
  
- [P] What are the options for encryption when using file-based streams?
  
- [P] Benchmark/estimate following using 1 stream for all and 1 stream per `app.*.*.*` and a memory- and file-based cluster:
  + Storage size requirement
  + Performance of publishing and subscribing in terms of throughput and latency
  + Cluster stability in the long run

### Phase 2

When moving to JetStream, what extra features/options can we provide as part of the NATS-based backend?

- What kind of ack models should we support?
- Delivery policy: should we allow receiving messages form the start of the topic?
- FlowControl: enforcing max-inflight (or max-ack-pending)
- JS supports different accounts on the same cluster, to support multi-tenancy on the same JS cluster. Do we need this? Probably not!
- Should we expose two different NATS-based "backends", one in-memory, and one file-based? Or if we create one Stream per
  `app.*.*.*`, should we allow choosing per Stream?


## Notes

What is a stream and how many do we need? A stream is a collection of subjects, and has its own configs such 
as: storage, replicas, retention, deduplication.

**Configuring a JetStream cluster**
- The replica count must be less than the maximum number of servers(!?) -> Do we need 4 servers for 3 replicas?
- The replica count cannot be edited once configured.

- Should we put all subjects in one stream?
We could have one default stream which is persisted to file which uses following subjects:
```
DEFAULT.*.*.*.*, DEFAULT.*.*.*
```
Or we could have one stream per app as in `app.obj.verb.version`
