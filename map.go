package regalia

// Map is a key/value map implemented as a mutable finite state transducer.
// Modifications to the map are enacted through transactions. Transactions are
// recorded in a log that forms a transaction history. Transactions are signed
// by one or more parties.
//
// Transactions are referenced within the map, but are not stored directly.
// A single transaction can affect more than one map.
type Map struct {
	// TODO: Add node ID?
	FST      Stream
	Log      Stream
	Head     uint64 // Offset of initial state within the FST page set.
	Revision uint64 // CRDT sum
}

// Descriptor is a mutable representation of the current state of a map root. It
// effectively holds the HEAD pointer.
type Descriptor struct {
}
