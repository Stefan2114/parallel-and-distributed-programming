package main

import (
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

func MpiSendInts(data []int64, dest int, tag int) {
	if len(data) == 0 {
		return
	}
	world.SendInt64s(data, dest, tag)
}

func MpiRecvInts(source int, tag int) []int64 {
	// The library uses Probe internally to determine size, allocates the buffer, and receives.
	val, _ := world.RecvInt64s(source, tag)
	return val
}

func SendPoly(poly []int64, dest int, tag int) {
	MpiSendInts(poly, dest, tag)
}

func RecvPoly(source int, tag int) []int64 {
	return MpiRecvInts(source, tag)
}

func BcastPoly(poly []int64, root int) []int64 {
	rank := world.Rank()

	lenBuf := make([]int64, 1)

	if rank == root {
		lenBuf[0] = int64(len(poly))
	}

	world.BcastInt64s(lenBuf, root)
	polyLen := lenBuf[0]

	if rank != root {
		if polyLen == 0 {
			return []int64{}
		}
		poly = make([]int64, polyLen)
	}

	if polyLen > 0 {
		world.BcastInt64s(poly, root)
	}

	return poly
}
