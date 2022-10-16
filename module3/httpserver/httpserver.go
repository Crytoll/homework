package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
)

func index(w http.ResponseWriter, r *http.Request) {
	os.Setenv("VERSION", "v0.0.1")
	version := os.Getenv("VERSION")
	w.Header().Set("VERSION", version)
	for k, v := range r.Header {
		for _, vv := range v {
			fmt.Printf("request header key is %s, value is %s\n", k, vv)
			w.Header().Set(k, vv)
		}
	}
	clientip := getCurrentIP(r)
	log.Printf("Success! Response code: %d", http.StatusOK)
	log.Printf("Success! clientip: %s", clientip)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "working")
}

func getCurrentIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		remoteAddr := r.RemoteAddr
		fmt.Printf("remoteAddr: %s\n", remoteAddr)
		remoteAddr0 := strings.Split(remoteAddr, ":")[0]
		fmt.Printf("X-Forwarded-For IP: %s\n", remoteAddr0)
		ip = remoteAddr0
	}
	return ip
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof", pprof.Index)
	mux.HandleFunc("/debug/pporf/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/", index)
	mux.HandleFunc("/healthz", healthz)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("start http server failed, error:%s\n", err.Error())
	}
}
