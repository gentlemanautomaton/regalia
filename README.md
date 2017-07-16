# regalia
A replicable hierarchical key/value store tuned for file system metadata.

This is not a production library; it is a thought experiment with a space for
prototyping.

Regalia is an attempt to work out the correct goals, overall design and API
boundaries of an ideal distributed file system metadata key/value store.

Unlike many distributed key/value stores, this library is not intended for
not intended to operate at massive cloud scale (exceeding 100 nodes).

## Design Goals (achieved or facilitated by this package)

1. Hierarchical key/value store with multiple data streams per key
2. Resistant to corrupted/botched writes (copy on write, validated reads)
3. Complete transaction history (with truncation and archival options)
4. Capable of identifying key and value degradation (hashed hierarchy)
5. Capable of identifying malicious data manipulation (cryptographically hashed hierarchy)
6. Amenable to replication and delta streaming of the entire hierarchy (root transaction log)
7. Amenable to replication and delta streaming of subhierarchies (bucket transaction log)
8. Tolerant of forking and merging (capable of causality determination)
8. Fast retrieval of small values (inline)
9. Compact representation of small values (length-prefixed value encoding, binary storage)
10. Compact representation of large values (referenced, chunked, de-duplicated, compressed, erasure encoded, binary storage)
11. Fast retrieval of large values (striped or mirrored)
12. Iterative migration of data formats and algorithms (iterative hash migration)
13. Knowledge of peers
14. Knowledge of replicas
15. Knowledge of actors

## External Design Goals (not in this package, but not obstructed by this package)

* Peer to peer communication
* Persistence of data to non-volatile media
* Volatile caching
* Non-volatile caching
* Separate persistence schemes for metadata and value storage
* Protect data by storing multiple replicas
* Every peer has a complete set of metadata and some recent history for the hierarchies it participates in
* A minimum number of replicas must be enforced
* Archive transaction history by moving it from hot storage to cold storage
* Move hot blocks from busy storage to less busy or idle storage
* Store copies of data accessed at the local site
* Select source peers for replication by a CRUSH algorithm that naturally distributes the load
* Move rarely used data to peers with high capacity (but probably slower) storage
* Storage transport (remove a disk from one machine and plug it into another)
* Replicate the most recent data first, then work backward
* Capable of operating as a single node
* Capable of operating as a distributed set of peers
* Read-only peers (data mirroring)
* Write-only peers (data archiving)

## Design Proposal

* Limit locks and mutable data structures to the head pointers
* Inline very small data blocks (<= 512 bytes)
* Reference larger data blocks stored
* Erasure encode very large data blocks that don't change much (> 64MB)
* Store key maps in finite state transducers
* Use append-only write patterns as much as possible
* Declare an inlining/referencing threshold per data stream type (so that ACLs are always in metadata, for example)
* Use interval tree clocks to determine causality

## Multiple Data Stream

Why are multiple data streams important? They let us do cool things in a way that
feels natural:

* Separate content from metadata
* Store things like access control lists, file origin information and author
* Perform iterative data processing and conversion
* Map ordinal identifiers to data stream names
