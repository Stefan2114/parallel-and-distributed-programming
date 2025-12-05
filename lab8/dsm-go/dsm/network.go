package dsm

import (
	"dsm-go/config"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Transport struct {
	dsm   *DSM
	peers map[int]*rpc.Client
	mu    sync.Mutex
}

func NewTransport(d *DSM, port string) *Transport {
	t := &Transport{
		dsm:   d,
		peers: make(map[int]*rpc.Client),
	}

	rpcHandler := &RPCHandler{dsm: d}
	server := rpc.NewServer()
	server.Register(rpcHandler)

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("Node %d listening on %s", d.SelfID, port)
	go server.Accept(l)

	return t
}

func (t *Transport) ConnectPeers(procs []config.Process) {
	for _, p := range procs {
		if p.ID == t.dsm.SelfID {
			continue
		}

		go func(target config.Process) {
			for {
				t.mu.Lock()
				_, connected := t.peers[target.ID]
				t.mu.Unlock()

				if connected {
					return
				}

				client, err := rpc.Dial("tcp", target.Host+":"+target.Port)
				if err == nil {
					t.mu.Lock()
					t.peers[target.ID] = client
					t.mu.Unlock()
					log.Printf("Connected to Node %d", target.ID)
					return
				}

				time.Sleep(2 * time.Second)
			}
		}(p)
	}
}

func (t *Transport) SendWriteRequest(ownerID int, varID, val int) error {
	t.mu.Lock()
	client, ok := t.peers[ownerID]
	t.mu.Unlock()
	if !ok {
		return fmt.Errorf("connection to owner %d not established", ownerID)
	}

	args := WriteRequest{VarID: varID, NewValue: val, Requester: t.dsm.SelfID}
	var reply Ack
	return client.Call("RPCHandler.HandleWriteRequest", args, &reply)
}

func (t *Transport) SendCASRequest(ownerID int, varID, oldVal, newVal int) (bool, error) {
	t.mu.Lock()
	client, ok := t.peers[ownerID]
	t.mu.Unlock()
	if !ok {
		return false, fmt.Errorf("connection to owner %d not established", ownerID)
	}

	args := CASRequest{VarID: varID, OldValue: oldVal, NewValue: newVal, Requester: t.dsm.SelfID}
	var reply CASResponse
	err := client.Call("RPCHandler.HandleCASRequest", args, &reply)
	return reply.Success, err
}

func (t *Transport) BroadcastUpdate(vConf config.VariableConfig, varID, val int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	args := UpdateMessage{VarID: varID, NewValue: val}

	for _, subID := range vConf.Subscribers {
		if subID == t.dsm.SelfID {
			continue // Don't send to self
		}
		if client, ok := t.peers[subID]; ok {
			// Async call to prevent blocking the owner
			go func(c *rpc.Client, id int) {
				var reply Ack
				if err := c.Call("RPCHandler.ReceiveUpdate", args, &reply); err != nil {
					log.Printf("Failed to update Node %d: %v", id, err)
				}
			}(client, subID)
		}
	}
}

type RPCHandler struct {
	dsm *DSM
}

func (h *RPCHandler) ReceiveUpdate(args UpdateMessage, reply *Ack) error {
	h.dsm.ApplyUpdate(args.VarID, args.NewValue)
	return nil
}

func (h *RPCHandler) HandleWriteRequest(args WriteRequest, reply *Ack) error {
	return h.dsm.coordinateUpdate(args.VarID, args.NewValue)
}

func (h *RPCHandler) HandleCASRequest(args CASRequest, reply *CASResponse) error {
	h.dsm.mu.Lock()
	current := h.dsm.Data[args.VarID]

	if current == args.OldValue {
		h.dsm.Data[args.VarID] = args.NewValue
		h.dsm.mu.Unlock()

		reply.Success = true
		reply.CurrentValue = args.NewValue

		if h.dsm.Callback != nil {
			go h.dsm.Callback(args.VarID, args.OldValue, args.NewValue)
		}

		vConf, _ := h.dsm.Config[args.VarID]
		h.dsm.Transport.BroadcastUpdate(vConf, args.VarID, args.NewValue)
	} else {
		h.dsm.mu.Unlock()
		reply.Success = false
		reply.CurrentValue = current
	}
	return nil
}
