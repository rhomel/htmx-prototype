package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"
)

var index = `
<!DOCTYPE html>
<html>
<head>
<title>htmx prototype</title>
<script src="https://unpkg.com/htmx.org@1.9.9"></script>
</head>
<body>
<h1>htmx prototype</h1>
<div hx-get="/count" hx-trigger="load" hx-swap="outerHTML"></div>
<button hx-post="/increment" hx-swap="none">Increment</button>
</body>
</html>
`

// FIXME: the incremented event on the client gets triggered twice causing 2
// requests for /count
var count = `
<p hx-get="/count" hx-trigger="incremented from:button">Count: {{.Count}}</p>
`

func main() {
	start()
}

func start() {
	s := &http.Server{
		Addr:           "127.0.0.1:8888",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	var err error
	s.Handler, err = newServer()
	if err != nil {
		log.Fatalf("server failed to initialize: %v", err)
	}
	log.Printf("http://%s", s.Addr)
	err = s.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server terminated with error: %v", err)
	}
}

var _ http.Handler = (*server)(nil)

type server struct {
	templates *templates
	mux       *http.ServeMux
	state     *state
}

func newServer() (*server, error) {
	t, err := newTemplates()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	s := &server{
		templates: t,
		mux:       mux,
		state:     &state{},
	}
	mux.HandleFunc("/increment", s.Increment)
	mux.HandleFunc("/count", s.Count)
	mux.HandleFunc("/", s.Index)
	return s, nil
}

func (s *server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	log.Printf("%s: %s", request.Method, request.URL.Path)
	s.mux.ServeHTTP(response, request)
}

func (s *server) Index(response http.ResponseWriter, request *http.Request) {
	s.templates.index.Execute(response, nil)
}

func (s *server) Increment(response http.ResponseWriter, request *http.Request) {
	s.state.Increment()
	response.Header().Set("HX-Trigger", "incremented")
	response.WriteHeader(http.StatusOK)
}

func (s *server) Count(response http.ResponseWriter, request *http.Request) {
	s.templates.count.Execute(response, s.state)
}

type state struct {
	Count int
}

func (s *state) Increment() {
	s.Count++
}

type templates struct {
	index *template.Template
	count *template.Template
}

func newTemplates() (*templates, error) {
	var err error
	t := &templates{}
	t.index, err = template.New("index").Parse(index)
	if err != nil {
		return nil, err
	}
	t.count, err = template.New("count").Parse(count)
	if err != nil {
		return nil, err
	}
	return t, nil
}
