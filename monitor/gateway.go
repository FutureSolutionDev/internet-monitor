package monitor

import (
	"bufio"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// defaultGatewayIPv4 returns the IPv4 default-route gateway. Implemented via
// /proc/net/route on Linux; returns ("", false) on other OSes or if not found.
func defaultGatewayIPv4() (string, bool) {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return "", false
	}
	defer f.Close()
	return parseProcRoute(f)
}

// parseProcRoute extracts the default-route gateway from /proc/net/route data.
func parseProcRoute(r io.Reader) (string, bool) {
	sc := bufio.NewScanner(r)
	sc.Scan() // skip header
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 4 {
			continue
		}
		if f[1] != "00000000" { // Destination 0.0.0.0 == default route
			continue
		}
		flags, err := strconv.ParseInt(f[3], 16, 64)
		if err != nil || flags&0x2 == 0 { // RTF_GATEWAY
			continue
		}
		if ip, ok := hexToIPv4(f[2]); ok {
			return ip, true
		}
	}
	return "", false
}

// hexToIPv4 decodes a little-endian 8-hex-digit address (as in /proc/net/route).
func hexToIPv4(h string) (string, bool) {
	if len(h) != 8 {
		return "", false
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return "", false
	}
	return net.IPv4(byte(v), byte(v>>8), byte(v>>16), byte(v>>24)).String(), true
}
