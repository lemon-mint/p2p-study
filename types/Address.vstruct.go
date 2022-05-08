package types

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

type _ = strings.Builder
type _ = unsafe.Pointer

var _ = math.Float32frombits
var _ = math.Float64frombits
var _ = strconv.FormatInt
var _ = strconv.FormatUint
var _ = strconv.FormatFloat
var _ = fmt.Sprint

type AddressType uint8

const (
	AddressType_IPv4     AddressType = 0
	AddressType_IPv6     AddressType = 1
	AddressType_DNS_A    AddressType = 2
	AddressType_DNS_AAAA AddressType = 3
	AddressType_DNS_SRV  AddressType = 4
)

func (e AddressType) String() string {
	switch e {
	case AddressType_IPv4:
		return "IPv4"
	case AddressType_IPv6:
		return "IPv6"
	case AddressType_DNS_A:
		return "DNS_A"
	case AddressType_DNS_AAAA:
		return "DNS_AAAA"
	case AddressType_DNS_SRV:
		return "DNS_SRV"
	}
	return ""
}

func (e AddressType) Match(
	onIPv4 func(),
	onIPv6 func(),
	onDNS_A func(),
	onDNS_AAAA func(),
	onDNS_SRV func(),
) {
	switch e {
	case AddressType_IPv4:
		onIPv4()
	case AddressType_IPv6:
		onIPv6()
	case AddressType_DNS_A:
		onDNS_A()
	case AddressType_DNS_AAAA:
		onDNS_AAAA()
	case AddressType_DNS_SRV:
		onDNS_SRV()
	}
}

func (e AddressType) MatchS(s struct {
	onIPv4     func()
	onIPv6     func()
	onDNS_A    func()
	onDNS_AAAA func()
	onDNS_SRV  func()
}) {
	switch e {
	case AddressType_IPv4:
		s.onIPv4()
	case AddressType_IPv6:
		s.onIPv6()
	case AddressType_DNS_A:
		s.onDNS_A()
	case AddressType_DNS_AAAA:
		s.onDNS_AAAA()
	case AddressType_DNS_SRV:
		s.onDNS_SRV()
	}
}

type Protocol uint8

const (
	Protocol_TCP  Protocol = 0
	Protocol_UDP  Protocol = 1
	Protocol_TLS  Protocol = 2
	Protocol_DTLS Protocol = 3
	Protocol_WS   Protocol = 4
	Protocol_WSS  Protocol = 5
	Protocol_QUIC Protocol = 6
)

func (e Protocol) String() string {
	switch e {
	case Protocol_TCP:
		return "TCP"
	case Protocol_UDP:
		return "UDP"
	case Protocol_TLS:
		return "TLS"
	case Protocol_DTLS:
		return "DTLS"
	case Protocol_WS:
		return "WS"
	case Protocol_WSS:
		return "WSS"
	case Protocol_QUIC:
		return "QUIC"
	}
	return ""
}

func (e Protocol) Match(
	onTCP func(),
	onUDP func(),
	onTLS func(),
	onDTLS func(),
	onWS func(),
	onWSS func(),
	onQUIC func(),
) {
	switch e {
	case Protocol_TCP:
		onTCP()
	case Protocol_UDP:
		onUDP()
	case Protocol_TLS:
		onTLS()
	case Protocol_DTLS:
		onDTLS()
	case Protocol_WS:
		onWS()
	case Protocol_WSS:
		onWSS()
	case Protocol_QUIC:
		onQUIC()
	}
}

func (e Protocol) MatchS(s struct {
	onTCP  func()
	onUDP  func()
	onTLS  func()
	onDTLS func()
	onWS   func()
	onWSS  func()
	onQUIC func()
}) {
	switch e {
	case Protocol_TCP:
		s.onTCP()
	case Protocol_UDP:
		s.onUDP()
	case Protocol_TLS:
		s.onTLS()
	case Protocol_DTLS:
		s.onDTLS()
	case Protocol_WS:
		s.onWS()
	case Protocol_WSS:
		s.onWSS()
	case Protocol_QUIC:
		s.onQUIC()
	}
}

type Address []byte

func (s Address) Type() AddressType {
	return AddressType(s[0])
}

func (s Address) Protocol() Protocol {
	return Protocol(s[1])
}

func (s Address) Port() uint16 {
	_ = s[3]
	var __v uint16 = uint16(s[2]) |
		uint16(s[3])<<8
	return uint16(__v)
}

func (s Address) Host() string {
	_ = s[11]
	var __off0 uint64 = 12
	var __off1 uint64 = uint64(s[4]) |
		uint64(s[5])<<8 |
		uint64(s[6])<<16 |
		uint64(s[7])<<24 |
		uint64(s[8])<<32 |
		uint64(s[9])<<40 |
		uint64(s[10])<<48 |
		uint64(s[11])<<56
	var __v = s[__off0:__off1]

	return *(*string)(unsafe.Pointer(&__v))
}

func (s Address) Vstruct_Validate() bool {
	if len(s) < 12 {
		return false
	}

	_ = s[11]

	var __off0 uint64 = 12
	var __off1 uint64 = uint64(s[4]) |
		uint64(s[5])<<8 |
		uint64(s[6])<<16 |
		uint64(s[7])<<24 |
		uint64(s[8])<<32 |
		uint64(s[9])<<40 |
		uint64(s[10])<<48 |
		uint64(s[11])<<56
	var __off2 uint64 = uint64(len(s))
	return __off0 <= __off1 && __off1 <= __off2
}

func (s Address) String() string {
	if !s.Vstruct_Validate() {
		return "Address (invalid)"
	}
	var __b strings.Builder
	__b.WriteString("Address {")
	__b.WriteString("Type: ")
	__b.WriteString(s.Type().String())
	__b.WriteString(", ")
	__b.WriteString("Protocol: ")
	__b.WriteString(s.Protocol().String())
	__b.WriteString(", ")
	__b.WriteString("Port: ")
	__b.WriteString(strconv.FormatUint(uint64(s.Port()), 10))
	__b.WriteString(", ")
	__b.WriteString("Host: ")
	__b.WriteString(strconv.Quote(s.Host()))
	__b.WriteString("}")
	return __b.String()
}

func Serialize_Address(dst Address, Type AddressType, Protocol Protocol, Port uint16, Host string) Address {
	_ = dst[11]
	dst[0] = byte(Type)
	dst[1] = byte(Protocol)
	dst[2] = byte(Port)
	dst[3] = byte(Port >> 8)

	var __index = uint64(12)
	__tmp_3 := uint64(len(Host)) + __index
	dst[4] = byte(__tmp_3)
	dst[5] = byte(__tmp_3 >> 8)
	dst[6] = byte(__tmp_3 >> 16)
	dst[7] = byte(__tmp_3 >> 24)
	dst[8] = byte(__tmp_3 >> 32)
	dst[9] = byte(__tmp_3 >> 40)
	dst[10] = byte(__tmp_3 >> 48)
	dst[11] = byte(__tmp_3 >> 56)
	copy(dst[__index:__tmp_3], Host)
	return dst
}

func New_Address(Type AddressType, Protocol Protocol, Port uint16, Host string) Address {
	var __vstruct__size = 12 + len(Host)
	var __vstruct__buf = make(Address, __vstruct__size)
	__vstruct__buf = Serialize_Address(__vstruct__buf, Type, Protocol, Port, Host)
	return __vstruct__buf
}
