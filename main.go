package main

import (
	"crypto/rand"
	"log"
	mrand "math/rand"
	"net"
	"time"

	"github.com/lemon-mint/godotenv"
	"github.com/lemon-mint/p2p-study/types"
)

func main() {
	godotenv.Load()

	myaddrs, err := MyAddresses()
	if err != nil {
		panic(err)
	}

	for i := range myaddrs {
		log.Println("Discovered address:", myaddrs[i].String())
	}
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

func distance(a, b uint64) uint64 {
	return a ^ b
}

type Peer struct {
	ID    uint64
	Addrs []types.Address
}

func MyAddresses() ([]types.Address, error) {
	var addrs []types.Address

	var addrSet = make(map[string]bool)

	ifaceaddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, ifaceaddr := range ifaceaddrs {
		var ip net.IP
		switch v := ifaceaddr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
			continue
		}

		if v := ip.To4(); v != nil {
			if _, ok := addrSet[v.String()]; !ok {
				addrSet[v.String()] = true
			}
			addrs = append(addrs, types.New_Address(
				types.AddressType_IPv4,
				types.Protocol_TCP,
				0,
				v.String(),
			))
		} else if v := ip.To16(); v != nil {
			if _, ok := addrSet[v.String()]; !ok {
				addrSet[v.String()] = true
			}
			addrs = append(addrs, types.New_Address(
				types.AddressType_IPv6,
				types.Protocol_TCP,
				0,
				v.String(),
			))
		} else {
			continue
		}
	}

	return addrs, nil
}

type Node struct {
	NodeID uint64

	// Peers is the list of peers
	Peers []Peer
}
