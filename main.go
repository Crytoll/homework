package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "200")
}

func webRoot(w http.ResponseWriter, r *http.Request) {
	h := r.Header
	version := os.Getenv("VERSION")
	fmt.Println("version:" + version)
	h.Add("VERSION", version)
	for k, v := range h {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	clientIP := ClientIP(r)
	log.Printf("agent ip: %v, status code: %d\n", clientIP, 200)
	w.Write([]byte("hello world\n"))
}

func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func main() {
	http.HandleFunc("/healthz", healthCheck)

	http.HandleFunc("/", webRoot)

	//在80端口监听客户端请求
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}
