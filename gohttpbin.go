package main

import (
    "fmt"
    "log"
    "net/http"
    "strings"
    "encoding/json"
)

func homepage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to gohttpbin!")
}

func ip(w http.ResponseWriter, r *http.Request) {
    addr := r.RemoteAddr[0:len(r.RemoteAddr)-strings.LastIndex(r.RemoteAddr, ":")-1]
    response, err := json.MarshalIndent(map[string]string{"origin": addr}, "", "  ")
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(w, string(response))
}

func main() {
    http.HandleFunc("/", homepage)
    http.HandleFunc("/ip", ip)

    fmt.Printf("Listening on port 8000...\n")
    err := http.ListenAndServe(":8000", nil)
    if err != nil { log.Fatal(err) }
}
