package serverpackets

import (
	"encoding/binary"
	"net"
)

const ServerListOpcode = 0x04

// ServerInfo holds data for one game server entry in the ServerList packet.
type ServerInfo struct {
	ID             byte
	IP             net.IP // 4 bytes
	Port           int32
	AgeLimit       byte
	PvP            bool
	CurrentPlayers int16
	MaxPlayers     int16
	Status         byte // 0 = down, 1 = up
	ServerType     int32
	Brackets       bool
	CharCount      byte
}

// ServerList writes the ServerList packet (opcode 0x04) into buf.
// Returns the number of bytes written.
func ServerList(buf []byte, servers []ServerInfo, lastServer byte) int {
	off := 0

	buf[off] = ServerListOpcode
	off++
	buf[off] = byte(len(servers))
	off++
	buf[off] = lastServer
	off++

	for _, s := range servers {
		buf[off] = s.ID
		off++

		ip := s.IP.To4()
		if ip == nil {
			ip = net.IPv4(127, 0, 0, 1).To4()
		}
		copy(buf[off:], ip[:4])
		off += 4

		binary.LittleEndian.PutUint32(buf[off:], uint32(s.Port))
		off += 4

		buf[off] = s.AgeLimit
		off++

		if s.PvP {
			buf[off] = 1
		} else {
			buf[off] = 0
		}
		off++

		binary.LittleEndian.PutUint16(buf[off:], uint16(s.CurrentPlayers))
		off += 2
		binary.LittleEndian.PutUint16(buf[off:], uint16(s.MaxPlayers))
		off += 2

		buf[off] = s.Status
		off++

		binary.LittleEndian.PutUint32(buf[off:], uint32(s.ServerType))
		off += 4

		if s.Brackets {
			buf[off] = 1
		} else {
			buf[off] = 0
		}
		off++
	}

	// Characters section
	binary.LittleEndian.PutUint16(buf[off:], 0) // unknown
	off += 2

	buf[off] = byte(len(servers))
	off++
	for _, s := range servers {
		buf[off] = s.ID
		off++
		buf[off] = s.CharCount
		off++
		buf[off] = 0 // delete count
		off++
	}

	return off
}
