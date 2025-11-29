package main

import (
	"bytes"
	"encoding/gob"
	"math/big"

	mpi "github.com/sbromberger/gompi"
)

// Global communicator for the "World"
var world *mpi.Communicator

func MpiInit() {
	// Start MPI (false = not multithreaded)
	mpi.Start(false)
	// Create the World communicator (nil = MPI_COMM_WORLD)
	world = mpi.NewCommunicator(nil)
}

func MpiFinalize() {
	mpi.Stop()
}

func MpiRank() int {
	return world.Rank()
}

func MpiSize() int {
	return world.Size()
}

func MpiBarrier() {
	world.Barrier()
}

func MpiSendBytes(data []byte, dest int, tag int) {
	if len(data) == 0 {
		return
	}
	world.SendBytes(data, dest, tag)
}

func MpiRecvBytes(source int, tag int) []byte {
	// The library uses Probe internally to determine size, allocates the buffer, and receives.
	val, _ := world.RecvBytes(source, tag)
	return val
}

func MpiBcastInt(val int, root int) int {
	vals := []int64{int64(val)}
	world.BcastInt64s(vals, root)
	return int(vals[0])
}

func MpiBcastBytes(data []byte, length int, root int) []byte {
	rank := MpiRank()
	worldSize := MpiSize()

	// Use a unique tag for internal broadcasts to avoid collision
	const BCAST_TAG = 32767

	if rank == root {
		// Root sends to everyone else
		for i := 0; i < worldSize; i++ {
			if i != root {
				if len(data) > 0 {
					world.SendBytes(data, i, BCAST_TAG)
				}
			}
		}
		return data
	} else {
		// Receivers wait for data from root
		if length > 0 {
			buf := make([]byte, length)
			// We use RecvPreallocBytes here because we already know the length from the
			// previous MpiBcastInt call in BcastPoly.
			world.RecvPreallocBytes(buf, root, BCAST_TAG)
			return buf
		}
		return []byte{}
	}
}

func serialize(data []*big.Int) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserialize(data []byte) ([]*big.Int, error) {
	var res []*big.Int
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func SendPoly(poly []*big.Int, dest int, tag int) {
	data, err := serialize(poly)
	if err != nil {
		panic(err)
	}

	MpiSendBytes(data, dest, tag)
}

func RecvPoly(source int, tag int) []*big.Int {

	data := MpiRecvBytes(source, tag)
	poly, err := deserialize(data)
	if err != nil {
		panic(err)
	}
	return poly
}

func BcastPoly(poly []*big.Int, root int) []*big.Int {
	rank := MpiRank()
	var length int
	var data []byte

	if rank == root {
		var err error
		data, err = serialize(poly)
		if err != nil {
			panic(err)
		}
		length = len(data)
	}

	length = MpiBcastInt(length, root)
	data = MpiBcastBytes(data, length, root)

	if rank != root {
		res, err := deserialize(data)
		if err != nil {
			panic(err)
		}
		return res
	}
	return poly
}
