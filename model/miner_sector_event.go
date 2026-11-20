package model

// MinerSectorEventType mirrors the public.miner_sector_event_type postgres enum.
type MinerSectorEventType string

const (
	SectorEventPrecommitAdded      MinerSectorEventType = "PRECOMMIT_ADDED"
	SectorEventPrecommitExpired    MinerSectorEventType = "PRECOMMIT_EXPIRED"
	SectorEventCommitCapacityAdded MinerSectorEventType = "COMMIT_CAPACITY_ADDED"
	SectorEventSectorAdded         MinerSectorEventType = "SECTOR_ADDED"
	SectorEventSectorExtended      MinerSectorEventType = "SECTOR_EXTENDED"
	SectorEventSectorExpired       MinerSectorEventType = "SECTOR_EXPIRED"
	SectorEventSectorFaulted       MinerSectorEventType = "SECTOR_FAULTED"
	SectorEventSectorRecovering    MinerSectorEventType = "SECTOR_RECOVERING"
	SectorEventSectorRecovered     MinerSectorEventType = "SECTOR_RECOVERED"
	SectorEventSectorTerminated    MinerSectorEventType = "SECTOR_TERMINATED"
)

// MinerSectorEvent is the DB model for the miner_sector_events table.
type MinerSectorEvent struct {
	Height    int64
	MinerID   int64
	SectorID  int64
	StateRoot CID
	Event     MinerSectorEventType
}
