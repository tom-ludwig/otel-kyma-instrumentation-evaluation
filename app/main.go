package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/rolldice", rolldice)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	result := doRoll()
	fmt.Fprintf(w, "rolled: %d\n", result)
}

func doRoll() int {
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	return rand.Intn(6) + 1
}
