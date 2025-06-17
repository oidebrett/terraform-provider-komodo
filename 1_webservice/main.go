package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
)

var users = make(map[string]string) // Map to store users with id as key and name as value
var mutex = &sync.RWMutex{}         // Mutex to protect access to the map as server is multithreaded

func main() {
	router := chi.NewRouter()

	router.Post("/{id}", func(response http.ResponseWriter, request *http.Request) {
		id := chi.URLParam(request, "id")
		name, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(response, "Failed to read request body", http.StatusBadRequest)
			return
		}
		mutex.Lock()
		defer mutex.Unlock()
		users[id] = string(name)
		fmt.Fprintf(response, "%s", string(name))
		fmt.Println("POST: ", id, " ", string(name))
	})

  router.Get("/{id}", func(response http.ResponseWriter, request *http.Request) {
		id := chi.URLParam(request, "id")
		mutex.RLock()
		defer mutex.RUnlock()
		name, ok := users[id]
		if !ok {
			http.NotFound(response, request)
			return
		}
		fmt.Fprintf(response, "%s", name)
		fmt.Println("GET: ", id, " ", name)
	})

	router.Put("/{id}", func(response http.ResponseWriter, request *http.Request) {
		id := chi.URLParam(request, "id")
		name, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(response, "Failed to read request body", http.StatusBadRequest)
			return
		}
		mutex.Lock()
		defer mutex.Unlock()
		if _, ok := users[id]; !ok {
			http.NotFound(response, request)
			return
		}
		users[id] = string(name)
		fmt.Fprintf(response, "%s", string(name))
		fmt.Println("PUT: ", id, " ", users[id])
	})

	router.Delete("/{id}", func(response http.ResponseWriter, request *http.Request) {
		id := chi.URLParam(request, "id")
		mutex.Lock()
		defer mutex.Unlock()
		name, ok := users[id]
		if !ok {
			http.NotFound(response, request)
			return
		}
		delete(users, id)
		fmt.Fprintf(response, "%s", name)
		fmt.Println("PUT: ", id, " ", users[id])
	})

	http.ListenAndServe(":6251", router)
}