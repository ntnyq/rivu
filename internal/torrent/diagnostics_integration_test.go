//go:build integration

package torrent

import (
	"encoding/binary"
	"net"
	"testing"
)

// A local BEP 15 tracker exercises the engine's UDP events, which cannot be
// observed through an HTTP transport or WebSocket status callbacks.
func TestLocalUDPTrackerDiagnostics(t *testing.T) {
	listener, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	t.Cleanup(func() { listener.Close(); <-done })
	go func() {
		defer close(done)
		packet := make([]byte, 2048)
		for {
			n, remote, err := listener.ReadFrom(packet)
			if err != nil {
				return
			}
			if n < 16 {
				continue
			}
			action := binary.BigEndian.Uint32(packet[8:12])
			response := make([]byte, 8)
			binary.BigEndian.PutUint32(response[:4], action)
			copy(response[4:8], packet[12:16])
			switch action {
			case 0: // connect
				response = binary.BigEndian.AppendUint64(response, 1)
			case 1: // announce: interval, leechers, seeders, no peers
				response = binary.BigEndian.AppendUint32(response, 60)
				response = append(response, make([]byte, 8)...)
			default:
				continue
			}
			_, _ = listener.WriteTo(response, remote)
		}
	}()
	checkLocalTrackerDiagnostics(t, "udp4://"+listener.LocalAddr().String()+"/private-passkey?token=secret", "ok")
}
