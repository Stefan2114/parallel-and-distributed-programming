package dsm

type VarState struct {
	Value   int
	Version int
}

type WriteRequest struct {
	VarID     int
	NewValue  int
	Requester int
}

type CASRequest struct {
	VarID     int
	OldValue  int
	NewValue  int
	Requester int
}

type CASResponse struct {
	Success      bool
	CurrentValue int
	Version      int
}

type UpdateMessage struct {
	VarID    int
	NewValue int
	Version  int
}

type Ack struct{}
