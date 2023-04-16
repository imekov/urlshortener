package middlewares

import (
	"net"
	"net/http"
	"strings"
)

// IPSubnet хранит данные для проверки подсети.
type IPSubnet struct {
	IP string
}

// CheckIP проверяет входит ли IP-адрес клиента в доверенную подсеть.
func (h IPSubnet) CheckIP(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realIP := net.ParseIP(r.Header.Get("X-Real-IP"))
		if realIP == nil {
			ips := r.Header.Get("X-Forwarded-For")
			ipStrs := strings.Split(ips, ",")
			ipStr := ipStrs[0]
			realIP = net.ParseIP(ipStr)
		}

		_, trustedSubnet, err := net.ParseCIDR(h.IP)
		if h.IP == "" || err != nil || !trustedSubnet.Contains(realIP) {
			http.Error(w, "the client IP address is not on a trusted subnet", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})

}
