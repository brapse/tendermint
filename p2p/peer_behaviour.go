package p2p

import (
	"errors"
	"sync"
)

// ErrorPeerBehaviour are types of observable erroneous peer behaviours.
type ErrorPeerBehaviour int

const (
	ErrorPeerBehaviourUnknown = iota
	ErrorPeerBehaviourBadMessage
	ErrorPeerBehaviourMessageOutofOrder
)

// Type of good behaviour a peer can perform.
type GoodPeerBehaviour int

const (
	GoodPeerBehaviourVote = iota + 100
	GoodPeerBehaviourBlockPart
)

// PeerBehaviour provides an interface for reactors to signal the behaviour
// of peers synchronously to other components.
type PeerBehaviour interface {
	Behaved(peerID ID, reason GoodPeerBehaviour) error
	Errored(peerID ID, reason ErrorPeerBehaviour) error
}

type SwitchPeerBehaviour struct {
	sw *Switch
}

// Reports the ErrorPeerBehaviour of peer identified by peerID to the Switch.
func (spb *SwitchPeerBehaviour) Errored(peerID ID, reason ErrorPeerBehaviour) error {
	peer := spb.sw.Peers().Get(peerID)
	if peer == nil {
		return errors.New("Peer not found")
	}

	spb.sw.StopPeerForError(peer, reason)
	return nil
}

// Reports the GoodPeerBehaviour of peer identified by peerID to the Switch.
func (spb *SwitchPeerBehaviour) Behaved(peerID ID, reason GoodPeerBehaviour) error {
	peer := spb.sw.Peers().Get(peerID)
	if peer == nil {
		return errors.New("Peer not found")
	}

	spb.sw.MarkPeerAsGood(peer)
	return nil
}

// Return a new switchPeerBehaviour instance which wraps the Switch.
func NewSwitchPeerBehaviour(sw *Switch) *SwitchPeerBehaviour {
	return &SwitchPeerBehaviour{
		sw: sw,
	}
}

// storedPeerBehaviour serves a mock concrete implementation of the
// PeerBehaviour interface used in reactor tests to ensure reactors
// produce the correct signals in manufactured scenarios.
type StoredPeerBehaviour struct {
	eb  map[ID][]ErrorPeerBehaviour
	gb  map[ID][]GoodPeerBehaviour
	mtx sync.RWMutex
}

// GettablePeerBehaviour provides an interface for accessing ErrorPeerBehaviours
// and GoodPeerBehaviour recorded by an implementation of PeerBehaviour
type GettablePeerBehaviour interface {
	GetErrorBehaviours(peerID ID) []ErrorPeerBehaviour
	GetGoodBehaviours(peerID ID) []GoodPeerBehaviour
}

// NewStoredPeerBehaviour returns a PeerBehaviour which records all observed
// behaviour in memory.
func NewStoredPeerBehaviour() *StoredPeerBehaviour {
	return &StoredPeerBehaviour{
		eb: map[ID][]ErrorPeerBehaviour{},
		gb: map[ID][]GoodPeerBehaviour{},
	}
}

// Errored stores the ErrorPeerBehaviour produced by the peer identified by ID.
func (spb *StoredPeerBehaviour) Errored(peerID ID, reason ErrorPeerBehaviour) {
	spb.mtx.Lock()
	defer spb.mtx.Unlock()
	if _, ok := spb.eb[peerID]; !ok {
		spb.eb[peerID] = []ErrorPeerBehaviour{reason}
	} else {
		spb.eb[peerID] = append(spb.eb[peerID], reason)
	}
}

// ErrorBehaviours returns all erorrs produced by peer identified by ID.
func (spb *StoredPeerBehaviour) GetErrorBehaviours(peerID ID) []ErrorPeerBehaviour {
	spb.mtx.RLock()
	defer spb.mtx.RUnlock()
	if items, ok := spb.eb[peerID]; ok {
		result := make([]ErrorPeerBehaviour, len(items))
		copy(result, items)

		return result
	} else {
		return []ErrorPeerBehaviour{}
	}
}

// Behaved stores the GoodPeerBehaviour of peer identified by ID.
func (spb *StoredPeerBehaviour) Behaved(peerID ID, reason GoodPeerBehaviour) {
	spb.mtx.Lock()
	defer spb.mtx.Unlock()
	if _, ok := spb.gb[peerID]; !ok {
		spb.gb[peerID] = []GoodPeerBehaviour{reason}
	} else {
		spb.gb[peerID] = append(spb.gb[peerID], reason)
	}
}

// GetGoodPeerBehaviours returns all positive behaviours produced by the peer
// identified by peerID.
func (spb *StoredPeerBehaviour) GetGoodBehaviours(peerID ID) []GoodPeerBehaviour {
	spb.mtx.RLock()
	defer spb.mtx.RUnlock()
	if items, ok := spb.gb[peerID]; ok {
		result := make([]GoodPeerBehaviour, len(items))
		copy(result, items)

		return result
	} else {
		return []GoodPeerBehaviour{}
	}
}
