// ICMP echo (real ping) with graceful failure so callers can fall back to TCP.
package monitor

import (
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// icmpPing sends one ICMP echo to host and returns (latencyMs, true) on a reply.
// It uses an unprivileged datagram socket, which works on Linux (when
// net.ipv4.ping_group_range permits) and macOS, but not on Windows; on any
// platform/permission where it's unavailable it returns (0, false) so the
// caller falls back to TCP. ICMP is therefore strictly an enhancement.
func icmpPing(host string, timeout time.Duration) (int64, bool) {
	ip, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return 0, false
	}
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return 0, false
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("im-ping")},
	}
	b, err := msg.Marshal(nil)
	if err != nil {
		return 0, false
	}

	start := time.Now()
	if _, err := conn.WriteTo(b, &net.UDPAddr{IP: ip.IP}); err != nil {
		return 0, false
	}
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, false
	}
	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return 0, false
	}
	rm, err := icmp.ParseMessage(1, reply[:n]) // 1 = IPv4 ICMP protocol number
	if err != nil {
		return 0, false
	}
	if rm.Type == ipv4.ICMPTypeEchoReply {
		return time.Since(start).Milliseconds(), true
	}
	return 0, false
}
