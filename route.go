package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-clog/clog"
	"github.com/gorilla/mux"
)

// ResponseLogger is a middleware used to keep an copy of Response.StatusCode.
//
type ResponseLogger struct {
	w          http.ResponseWriter
	StatusCode int
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
func (m *ResponseLogger) Header() http.Header {
	return m.w.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (m *ResponseLogger) Write(data []byte) (int, error) {
	return m.w.Write(data)
}

// WriteHeader sends an HTTP response header with status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (m *ResponseLogger) WriteHeader(status int) {
	if m.StatusCode == 0 {
		// since status code can only write once.
		m.StatusCode = status
	}
	m.w.WriteHeader(status)
}

// IndexHandle for index page
//
func IndexHandle(w http.ResponseWriter, r *http.Request) {
	root, _ := filepath.Abs("./public")
	index := filepath.Join(root, "index.html")

	if tmpl, err := template.ParseFiles(index); err == nil {
		tmpl.Execute(w, nil)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
	}
}

// StaticHandler serve public files, exclude folder
//
func StaticHandler() http.Handler {
	root, _ := filepath.Abs("./public")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if request folder, return not found
		if strings.TrimRight(r.RequestURI, "/") != r.RequestURI {
			clog.Warn("[Res] Access (%s) forbidden.", r.RequestURI)
			http.NotFound(w, r)
		} else {
			clog.Info("[Res] Access %s", r.RequestURI)
			http.FileServer(http.Dir(root)).ServeHTTP(w, r)
		}
	})
}

// LogHandler print request trace log
//
func LogHandler(h http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w := &ResponseLogger{w: resp}
		h.ServeHTTP(w, r)

		// "GET / HTTP/1.1" 200 2552 UserAgent
		clog.Info("[API] %s - %d %s %s %s - %s",
			r.RemoteAddr,
			w.StatusCode,
			r.Proto,
			r.Method,
			r.RequestURI,
			time.Since(start))
	})
}

// NewRouter return the registered router
//
func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.StrictSlash(true)

	// static files handler
	router.
		PathPrefix("/public/").
		Handler(LogHandler(StaticHandler(), "log"))

	return router
}
