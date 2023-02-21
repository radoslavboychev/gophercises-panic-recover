package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(":3000", recoverMw(mux, true)))
}

// recoverMw
func recoverMw(app http.Handler, dev bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				http.Error(w, "Something went wrong - Panic!", http.StatusInternalServerError)
				stack := debug.Stack()
				log.Println(string(stack))
				if !dev {
					http.Error(w, "Something went wrong", http.StatusInternalServerError)
					return
				}
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, string(stack))
			}
		}()
		nw := &responseWriter{ResponseWriter: w}
		nw.flush()
		app.ServeHTTP(w, r)
	}
}

// responseWriter
type responseWriter struct {
	http.ResponseWriter
	writes [][]byte
	status int
}

// Write
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.writes = append(rw.writes, b)
	return len(b), nil
}

// WriteHeader
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

// flush
func (rw *responseWriter) flush() error {
	if rw.status != 0 {
		rw.ResponseWriter.WriteHeader(rw.status)
	}
	for _, write := range rw.writes {
		_, err := rw.ResponseWriter.Write(write)
		if err != nil {
			return nil
		}
	}
	return nil
}

// panicDemo triggers panic
func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

// panicAfterDemo
func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

// funcThatPanics triggers a panic state
func funcThatPanics() {
	panic("OHHHHH")
}

// hello default home page
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}
