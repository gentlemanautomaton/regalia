# regalia

A replicable hierarchical key/value store tuned for file system metadata.

This is not a production library; it is a thought experiment with a space for
prototyping.

Regalia is an attempt to work out the correct goals, overall design and API
boundaries of an ideal distributed file system metadata key/value store.

Unlike many distributed key/value stores, it is not a goal of this system to
operate at massive cloud scale (exceeding 100 nodes).

## Design Goals (achieved or facilitated by this package)

1. Hierarchical key/value store with multiple data streams (attributes) per key
2. Resistant to corrupted/botched writes (copy on write, validated reads)
3. Complete transaction history (with truncation and archival options)
4. Capable of identifying key and value degradation (hashed hierarchy)
5. Capable of identifying malicious data manipulation (cryptographically hashed hierarchy)
6. Amenable to replication and delta streaming of the entire hierarchy (root transaction log)
7. Amenable to replication and delta streaming of subhierarchies (bucket transaction log)
8. Tolerant of forking and merging (capable of causality determination)
9. Fast retrieval of small values (inline)
10. Compact representation of small values (length-prefixed value encoding, binary storage)
11. Compact representation of large values (referenced, chunked, de-duplicated, compressed, erasure encoded, binary storage)
12. Fast retrieval of large values (striped or mirrored)
13. Iterative migration of data formats and algorithms (iterative hash migration)
14. Knowledge of peers
15. Knowledge of replicas
16. Knowledge of actors

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

## Attributes (multiple data streams per value)

Storing multiple data streams per value lets us achieve the following goals in
a way that feels natural:

* Separate content from metadata
* Store things like access control lists, file origin information and author
* Perform iterative data processing and conversion
* Map ordinal identifiers to data stream names

Each value in regalia is made up of one or more data streams, commonly called
attributes. Each attribute has an identifier that distinguishes it and an
optional encoding descriptor that can describe the encoding of simple values.

The unique combination of attributes can be thought of as a value's type.
Values of the same type have the same set of attributes. When attributes are
added or deleted, values are automatically transitioned between types.

Types are useful because they facilitate efficient encoding of values. Instead
of repeatedly describing the set of attributes present in each value, a value
need only refer to its type to describe how to parse its attribute set.

When encoding values, regalia stores the type identifier at the start of the
byte stream, followed by an array of offsets, followed by the attribute data
streams.

Here's an example of a value with three attributes:

```
|-------------------------|
| TYPE         | Fixed    |
|-------------------------|
| Attr0 Offset | Fixed    |
|-------------------------|
| Attr1 Offset | Fixed    |
|-------------------------|
| Attr2 Offset | Fixed    |
|-------------------------|
| Attr0        | Variable |
|-------------------------|
| Attr1        | Variable |
|-------------------------|
| Attr2        | Variable |
|-------------------------|
```

When an attribute is modified and a new value is recorded, a new pseudo-type is
created that includes a back-reference to the previous value, plus the updated
attribute(s):

```
|-------------------------|
| TYPE         | Fixed    |
|-------------------------|
| Reference    | Fixed    |
|-------------------------|
| Attr2 Offset | Fixed    |
|-------------------------|
| Attr2        | Variable |
|-------------------------|
```

This design facilitates efficient copy-on-write storage of small changes to
large attribute sets. The drawback is that attribute retrieval may require
additional reads as back-references are followed.

Types with back-references may be chained, thus increasing the indirection
required to locate a particular attribute's data stream. If regalia determines
that an indirection limit would be exceeded by an addition of a new
back-reference, regalia will instead collapse the attribute changes into a
newly allocated copy of the real type without any back-references.

A potential optimization here would be to build a computational machine for
each type that quickly retrieves a particular attribute. Once constructed and
cached in-memory, each retrieval would run the machine associated with the
desired attribute.

The underlying assumption is that regalia is used to store values with common
types that follow common type transitions.

Inspiration for attribute design can be drawn from the following:

* Chrome's V8 JavaScript engine type transitions
* GOB encoding
* NTFS attributes

## Address Space

Most data in regalia is saved into a theoretically limitless append-only byte
stream. Back-references are offsets into that stream used to refer back to
previously written data.

In order to minimize the distance that back-references must go into the past,
regalia will occasionally copy data forward and update values to use the new
back-references. This allows stale data that appeared earlier in the byte
stream to be archived, truncated or moved onto slower storage media. Offsets
within the stream are never reused, even when data has been truncated.

Under consideration is multiplexing of the data stream. The stream could be
divided into chunks of typed substreams: FST nodes, transitions, values,
etc. Multiplexing may improve the locality of reference for FST traversal. A
multiplexing layer added over the top of the root data stream would incur
additional complexity.

Inspiration for data stream design can be drawn from the following:

* NTFS Change journals

## Data Blocks

Data within the byte stream is organized into blocks, with each block holding
one or more cryptographic references to its predecessor. This forms a sort of
block chain.

Blocks are formed after a set of peers reach consensus about the history of
events. Once a block has been formed, peers may refer to data stored within it
via back-references.

Note: Store a base offset at the start of a block, and then encode offsets
as deltas against the block's base offset. Alternatively, include the block
number in each reference; numbering may start at 0 for the previous block,
followed by 1 for the block before that, and so on.

Inspiration for data block design can be drawn from the following:

* git (think of blocks as commits)
* Block chains (bitcoin, ethereum, etc.)
* Video compression (intra-frames vs inter-frames)

## Finite-State Transducers

One of the goals of this project is to explore the use of finite-state
transducers for mapping key data. In FST parlance, the input tape provides the
desired key and the output tape yields an offset or index that locates a value.

Finite-state transducers are well suited for static data, but the complexity of
their creation can create challenges for mutable or slowly changing data sets.

Each node in an FST is an N-way transition map, with N being the number of
transitions for a particular node. When N is small, a simple data structure
such as a list will be used. As N grows, more complex data structures may be
used that are less space efficient but allow for faster retrieval.

Some possible implementations under consideration:

* Compaction of single-transition nodes into a series
* Flat lists for low transition nodes
* 4-bit tries (ala ethereum patricia tries) for sparse transition nodes
* 256-way index for dense nodes (possibly with a 2-tier bitmap)

Inspiration for FST design can be drawn from the following:

* [Ethereum Patricia Tries](https://github.com/ethereum/wiki/wiki/Patricia-Tree)
* [Index 1,600,000,000 Keys with Automata and Rust](http://blog.burntsushi.net/transducers/)

## Indexing

Indexes are built per-attribute. Under consideration is something similar to
SQL Server's columnstore.
