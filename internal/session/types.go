package session

type CommandKind string
type Phase string
type PauseReason string

const (
	CommandPause       CommandKind = "pause"
	CommandResume      CommandKind = "resume"
	CommandPauseAll    CommandKind = "pauseAll"
	CommandResumeBatch CommandKind = "resumeBatch"
	CommandVerify      CommandKind = "verify"
	CommandPriority    CommandKind = "priority"
	CommandRemove      CommandKind = "remove"
	CommandSeed        CommandKind = "seed"
	CommandFirst       CommandKind = "first"
	CommandNotified    CommandKind = "notified"
	CommandRelocate    CommandKind = "relocate"
	CommandTags        CommandKind = "tags"
	PhaseMetadata      Phase       = "metadata"
	PhaseWaiting       Phase       = "waiting"
	PhaseDownloading   Phase       = "downloading"
	PhaseVerifying     Phase       = "verifying"
	PhaseSeeding       Phase       = "seeding"
	PhaseQueued        Phase       = "queued"
	PhasePaused        Phase       = "paused"
	PhaseCompleted     Phase       = "completed"
	PhaseRemoving      Phase       = "removing"
	PhaseError         Phase       = "error"
	PauseManual        PauseReason = "manual"
	PauseRatio         PauseReason = "ratio"
	PauseVerify        PauseReason = "verify"
	PauseDuration      PauseReason = "duration"
)
