# Proposal for multiple active Eventing backends

### Context and scope


### Goals and Non-goals


### Constraints

- The state of each EventingBackendType and the chosen default/active backend should not be overwritten by the Reconciler.

### Design

- We add a new `EventingBackendType` CRD, which can include backend-specific configs.
- We evolve the current `EventingBackend` CRD to point to an `EventingBackendType` CR.

```

```


### APIs


### Alternatives Considered