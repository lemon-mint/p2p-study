package main

import (
	"crypto/rand"
	mrand "math/rand"
	"time"

	"github.com/lemon-mint/godotenv"
)

func main() {
	godotenv.Load()
}

type Address struct {
}

func getNodeID() uint64 {
	mrand.Seed(int64(time.Now().Nanosecond()))
	var b [4096]byte
	rand.Read(b[:])
	var id uint64 = mrand.Uint64()
	for i := 0; i < len(b)/8; i++ {
		id ^= uint64(b[i*8]) |
			uint64(b[i*8+1])<<8 |
			uint64(b[i*8+2])<<16 |
			uint64(b[i*8+3])<<24 |
			uint64(b[i*8+4])<<32 |
			uint64(b[i*8+5])<<40 |
			uint64(b[i*8+6])<<48 |
			uint64(b[i*8+7])<<56
	}
	return id
}

type Node struct {
	NodeID uint64
}
