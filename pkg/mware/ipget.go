package mware

import (
	"log"
	"net/http"
	"net/netip"
)

func GetIPMiddleware(h http.HandlerFunc, trustedSubnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if trustedSubnet == "" {
			log.Println("doing without trustedSubnet")
			h.ServeHTTP(w, r)
			return
		}

		ipStr := r.Header.Get("X-Real-IP")

		network, err := netip.ParsePrefix(trustedSubnet)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ip, err := netip.ParseAddr(ipStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !network.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// передаём управление хендлеру
		h.ServeHTTP(w, r)
	}
}
