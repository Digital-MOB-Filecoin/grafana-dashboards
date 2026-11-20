package model

// ActorType maps a Filecoin actor type to its code CID.
// Singleton builtin actors also carry an ActorID.
type ActorType struct {
	ActorTypeID  int
	ActorID      *int  // nil for non-singletons
	ActorCodeCID CID
}

// ActorCode records the CID used for a given actor type at a specific network version.
type ActorCode struct {
	ActorTypeID    int
	ActorCodeCID   CID
	NetworkVersion int
}

// ActorMethod describes a method exported by an actor type.
type ActorMethod struct {
	ActorTypeID int
	MethodName  string
	Method      uint64
}
