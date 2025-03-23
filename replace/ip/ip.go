package ip

import (
	"net"

	"github.com/c2pc/config-migrate/internal/replacer"
)

func init() {
	replacer.Register("___ip_address___", ipReplacer)
}

func ipReplacer() string {
	tt, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return "localhost"
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}

			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 {
				continue
			}
			return v4.String()
		}
	}

	return "localhost"
}
