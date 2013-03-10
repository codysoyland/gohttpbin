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

func useragent(w http.ResponseWriter, r *http.Request) {
    response, err := json.MarshalIndent(map[string]string{"user-agent": r.UserAgent()}, "", "  ")
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(w, string(response))
}

func headers(w http.ResponseWriter, r *http.Request) {
    headers := map[string]string{}
    for k, v := range r.Header {
        headers[k] = v[0]
    }
    response, err := json.MarshalIndent(map[string]map[string]string{"headers": headers}, "", "  ")
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(w, string(response))
}


func main() {
    http.HandleFunc("/", homepage)
    http.HandleFunc("/ip", ip)
    http.HandleFunc("/user-agent", useragent)
    http.HandleFunc("/headers", headers)

    fmt.Printf("Listening on port 8000...\n")
    err := http.ListenAndServe(":8000", nil)
    if err != nil { log.Fatal(err) }
}
