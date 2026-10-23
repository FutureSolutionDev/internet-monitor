package monitor

import (
	"bufio"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func defaultGatewayIPv4() (string, bool) {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return "", false
	}
	defer f.Close()
	return parseProcRoute(f)
}

func parseProcRoute(r io.Reader) (string, bool) {
	sc := bufio.NewScanner(r)
	sc.Scan()
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 4 {
			continue
		}
		if f[1] != "00000000" {
			continue
		}
		flags, err := strconv.ParseInt(f[3], 16, 64)
		if err != nil || flags&0x2 == 0 {
			continue
		}
		if ip, ok := hexToIPv4(f[2]); ok {
			return ip, true
		}
	}
	return "", false
}

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
