package service

import (
	"log"
	"net/http"
	"strings"
)

type HttpHandler struct {
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Println("<<< ", r.Method, path)
	if path == "" || path == "/" {
		writeResponse(w, http.StatusOK, "Welcome")
	} else if path == "/favicon.ico" {
		writeResponse(w, http.StatusNotFound, "")
	} else if path == "call" && strings.ToLower(r.Method) == "post" {

	} else {
		writeResponse(w, http.StatusNotFound, "")
	}
}

func writeResponse(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "plain/text")
	if len(body)>0 {
		_, _ = w.Write([]byte(body))
	}
	log.Printf(">>> %d %s", status, body)
}
