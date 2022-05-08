package main

import (
	"crypto/rand"
	"io"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
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

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	lnip := ln.Addr().String()
	portpos := strings.LastIndex(lnip, ":")
	portStr := lnip[portpos+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	for i := range myaddrs {
		myaddrs[i] = types.New_Address(
			myaddrs[i].Type(),
			myaddrs[i].Protocol(),
			uint16(port),
			myaddrs[i].Host(),
		)
	}

	for i := range myaddrs {
		log.Println("Discovered address:", myaddrs[i].String())
	}

	myid := getNodeID()
	log.Println("My ID:", myid)

	me := Peer{
		ID:    myid,
		Addrs: myaddrs,
	}

	log.Println("Me:", me)
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

	// Get Cloudflare IP
	resp, err := http.Get("https://cloudflare.com/cdn-cgi/trace")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	cfip := godotenv.Parse(string(body))["ip"]
	if cfip != "" {
		if _, ok := addrSet[cfip]; !ok {
			addrSet[cfip] = true
			addrs = append(addrs, types.New_Address(
				types.AddressType_IPv4,
				types.Protocol_TCP,
				0,
				cfip,
			))
		}
	}

	return addrs, nil
}

func Dial(addrs []types.Address) (net.Conn, error) {
	var err error
	var conn net.Conn
	for _, addr := range addrs {
		if addr.Protocol() == types.Protocol_TCP {
			conn, err = net.DialTimeout(addr.Protocol().String(), addr.String(), time.Second*1)
			if err != nil {
				continue
			}
			return conn, nil
		}
	}

	return nil, err
}

type Node struct {
	NodeID uint64

	// Peers is the list of peers
	Peers []Peer

	mu sync.RWMutex
}

func (n *Node) AddPeer(p Peer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Peers = append(n.Peers, p)
	sort.Slice(n.Peers, func(i, j int) bool {
		return n.Peers[i].ID < n.Peers[j].ID
	})
}

func (n *Node) Bootstrap(peers []Peer) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Peers = append(n.Peers, peers...)
	sort.Slice(n.Peers, func(i, j int) bool {
		return n.Peers[i].ID < n.Peers[j].ID
	})
}
