package dsm

type VariableState struct {
	ID    int
	Value int
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
}

type UpdateMessage struct {
	VarID    int
	NewValue int
}

type Ack struct{}
