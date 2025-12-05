package dsm

import (
	"dsm-go/config"
	"fmt"
	"log"
	"sync"
	"time"
)

type CallbackFunc func(varID int, oldVal int, newVal int)

type DSM struct {
	SelfID   int
	Data     map[int]VarState
	Config   map[int]config.VariableConfig
	Callback CallbackFunc

	mu sync.RWMutex

	Transport *Transport
}

func NewDSM(selfID int, port string, cb CallbackFunc) *DSM {
	d := &DSM{
		SelfID:   selfID,
		Data:     make(map[int]VarState),
		Config:   make(map[int]config.VariableConfig),
		Callback: cb,
	}

	for _, vConf := range config.Variables {
		isSubscriber := false
		for _, sub := range vConf.Subscribers {
			if sub == selfID {
				isSubscriber = true
				break
			}
		}
		if isSubscriber {
			d.Config[vConf.ID] = vConf
			d.Data[vConf.ID] = VarState{Value: 0, Version: 0}
		}
	}

	d.Transport = NewTransport(d, port)

	go func() {
		time.Sleep(2 * time.Second)
		d.Transport.ConnectPeers(config.AllProcesses)
	}()

	return d
}

func (d *DSM) Get(varID int) (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	state, ok := d.Data[varID]
	return state.Value, ok
}

func (d *DSM) Write(varID int, val int) error {
	vConf, ok := d.Config[varID]
	if !ok {
		return fmt.Errorf("node %d not subscribed to var %d", d.SelfID, varID)
	}

	if vConf.OwnerID == d.SelfID {
		return d.coordinateUpdate(varID, val)
	}

	return d.Transport.SendWriteRequest(vConf.OwnerID, varID, val)
}

func (d *DSM) CompareAndExchange(varID int, oldVal int, newVal int) (bool, error) {
	vConf, ok := d.Config[varID]
	if !ok {
		return false, fmt.Errorf("node %d not subscribed to var %d", d.SelfID, varID)
	}

	if vConf.OwnerID == d.SelfID {
		d.mu.Lock()
		currentState := d.Data[varID]

		if currentState.Value == oldVal {
			newState := VarState{
				Value:   newVal,
				Version: currentState.Version + 1,
			}
			d.Data[varID] = newState
			d.mu.Unlock()

			if d.Callback != nil {
				go d.Callback(varID, currentState.Value, newVal)
			}

			d.Transport.BroadcastUpdate(vConf, varID, newVal, newState.Version)
			return true, nil
		}
		d.mu.Unlock()
		return false, nil
	}

	return d.Transport.SendCASRequest(vConf.OwnerID, varID, oldVal, newVal)
}

func (d *DSM) coordinateUpdate(varID int, newVal int) error {
	d.mu.Lock()

	currentState := d.Data[varID]
	oldVal := currentState.Value

	newState := VarState{
		Value:   newVal,
		Version: currentState.Version + 1,
	}
	d.Data[varID] = newState

	d.mu.Unlock()

	log.Printf("[DSM] Owner updated Var %d: %d -> %d (Ver %d)", varID, oldVal, newVal, newState.Version)

	if d.Callback != nil {
		go d.Callback(varID, oldVal, newVal)
	}

	vConf := d.Config[varID]
	d.Transport.BroadcastUpdate(vConf, varID, newVal, newState.Version)
	return nil
}

func (d *DSM) ApplyUpdate(varID int, newVal int, newVersion int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	currentState := d.Data[varID]

	if newVersion > currentState.Version {
		oldVal := currentState.Value

		d.Data[varID] = VarState{
			Value:   newVal,
			Version: newVersion,
		}

		log.Printf("[DSM] Replica accepted update Var %d: %d (Ver %d)", varID, newVal, newVersion)

		if d.Callback != nil {
			go d.Callback(varID, oldVal, newVal)
		}
	} else {
		log.Printf("[DSM] Replica IGNORED stale update Var %d (Incoming Ver %d <= Local Ver %d)",
			varID, newVersion, currentState.Version)
	}
}
