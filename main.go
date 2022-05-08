package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lemon-mint/godotenv"
	"github.com/lemon-mint/p2p-study/types"
)

var t *http.Transport = &http.Transport{
	DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
		addrs := String2Addrs(addr)
		if len(addrs) > 0 {
			conn, err := Dial(addrs)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}
		return net.DialTimeout(network, addr, time.Second*5)
	},
}

var client = &http.Client{Transport: t}

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

	log.Println("My addresses:", Addrs2String(myaddrs))

	myid := getNodeID()
	log.Println("My ID:", myid)

	n := Node{
		NodeID: myid,
		Addrs:  myaddrs,
	}
	go func() {
		err := n.StartServer(ln)
		panic(err)
	}()

	resp, err := client.Get("http://" + Addrs2String(myaddrs) + "/id")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	log.Println("Response:", string(body))

}

func Addrs2String(addrs []types.Address) string {
	var sb strings.Builder
	for i, addr := range addrs {
		if i != 0 {
			sb.WriteString(".")
		}
		strAddr := base64.RawURLEncoding.EncodeToString(addr)
		sb.WriteString(strAddr)
	}
	return sb.String()
}

func String2Addrs(s string) []types.Address {
	var addrs []types.Address
	for _, addr := range strings.Split(s, ".") {
		if addr == "" {
			continue
		}
		byteAddr, err := base64.RawURLEncoding.DecodeString(addr)
		if err != nil {
			continue
		}
		a := types.Address(byteAddr)
		if !a.Vstruct_Validate() {
			continue
		}
		addrs = append(addrs, a)
	}
	return addrs
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

func (p Peer) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID    uint64
		Addrs string
	}{
		ID:    p.ID,
		Addrs: Addrs2String(p.Addrs),
	})
}

func (p Peer) UnmarshalJSON(b []byte) error {
	var s struct {
		ID    uint64
		Addrs string
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	p.ID = s.ID
	p.Addrs = String2Addrs(s.Addrs)
	return nil
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
		log.Println("Dialing:", addr)
		if addr.Protocol() == types.Protocol_TCP {
			conn, err = net.DialTimeout("tcp", addr.Host()+":"+strconv.Itoa(int(addr.Port())), time.Second*1)
			if err != nil {
				continue
			}
			return conn, nil
		}
		if addr.Protocol() == types.Protocol_UDP {
			conn, err = net.DialTimeout("udp", addr.Host()+":"+strconv.Itoa(int(addr.Port())), time.Second*1)
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
	Addrs  []types.Address

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

	// Add self to the list of peers
	n.Peers = append(n.Peers, Peer{
		ID:    n.NodeID,
		Addrs: n.Addrs,
	})

	// Get the list of peers to bootstrap from
	var bootstrapPeers []Peer
	var addrs []Peer
	for _, peer := range n.Peers {
		if peer.ID != n.NodeID {
			resp, err := http.Get("http://" + Addrs2String(peer.Addrs) + "/peers")
			if err != nil {
				continue
			}
			defer resp.Body.Close()
			err = json.NewDecoder(resp.Body).Decode(&addrs)
			if err != nil {
				continue
			}
			bootstrapPeers = append(bootstrapPeers, addrs...)
			addrs = addrs[:0]
		}
	}
	n.Peers = append(n.Peers, bootstrapPeers...)
	sort.Slice(n.Peers, func(i, j int) bool {
		return n.Peers[i].ID < n.Peers[j].ID
	})

	// Remove duplicates
	var uniquePeers []Peer
	var seenPeers = make(map[uint64]bool)
	for _, peer := range n.Peers {
		if _, ok := seenPeers[peer.ID]; !ok {
			seenPeers[peer.ID] = true
			uniquePeers = append(uniquePeers, peer)
		}
	}
	n.Peers = uniquePeers
}

func (n *Node) StartServer(ln net.Listener) error {
	r := httprouter.New()

	r.GET("/peers", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		n.mu.RLock()
		defer n.mu.RUnlock()

		err := json.NewEncoder(w).Encode(n.Peers)
		if err != nil {
			log.Println(err)
		}
	})

	r.GET("/id", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		n.mu.RLock()
		defer n.mu.RUnlock()

		w.Write([]byte(strconv.FormatUint(n.NodeID, 10)))
	})

	return http.Serve(ln, r)
}
