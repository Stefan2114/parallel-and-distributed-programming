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
	Data     map[int]int
	Config   map[int]config.VariableConfig
	Callback CallbackFunc

	mu sync.RWMutex

	Transport *Transport
}

func NewDSM(selfID int, port string, cb CallbackFunc) *DSM {
	d := &DSM{
		SelfID:   selfID,
		Data:     make(map[int]int),
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
			d.Data[vConf.ID] = 0
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
	val, ok := d.Data[varID]
	return val, ok
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
		current := d.Data[varID]
		if current == oldVal {
			d.Data[varID] = newVal
			d.mu.Unlock()

			if d.Callback != nil {
				go d.Callback(varID, oldVal, newVal)
			}

			d.Transport.BroadcastUpdate(vConf, varID, newVal)
			return true, nil
		}
		d.mu.Unlock()
		return false, nil
	}

	return d.Transport.SendCASRequest(vConf.OwnerID, varID, oldVal, newVal)
}

func (d *DSM) coordinateUpdate(varID int, newVal int) error {
	d.mu.Lock()
	oldVal := d.Data[varID]
	d.Data[varID] = newVal
	d.mu.Unlock()

	log.Printf("[DSM] Owner updating Var %d: %d -> %d", varID, oldVal, newVal)

	if d.Callback != nil {
		go d.Callback(varID, oldVal, newVal)
	}

	vConf := d.Config[varID]
	d.Transport.BroadcastUpdate(vConf, varID, newVal)
	return nil
}

func (d *DSM) ApplyUpdate(varID int, newVal int) {
	d.mu.Lock()
	oldVal := d.Data[varID]
	d.Data[varID] = newVal
	d.mu.Unlock()

	log.Printf("[DSM] Replica updated Var %d: %d -> %d", varID, oldVal, newVal)

	if d.Callback != nil {
		d.Callback(varID, oldVal, newVal)
	}
}
