package main

import (
    "fmt"
    "log"
    "net/http"
    "strings"
    "encoding/json"
    "compress/gzip"
)

type Headers map[string]string
type ResponseDict map[string]interface{}

func homepage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to gohttpbin!")
}

func getIp(r *http.Request)string {
    return r.RemoteAddr[0:len(r.RemoteAddr)-strings.LastIndex(r.RemoteAddr, ":")-1]
}

func ip(w http.ResponseWriter, r *http.Request) {
    response, err := json.MarshalIndent(map[string]string{"origin": getIp(r)}, "", "  ")
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(w, string(response))
}

func useragent(w http.ResponseWriter, r *http.Request) {
    response, err := json.MarshalIndent(map[string]string{"user-agent": r.UserAgent()}, "", "  ")
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(w, string(response))
}

func getHeaders(r *http.Request)Headers {
    headers := Headers{}
    for k, v := range r.Header {
        headers[k] = v[0]
    }
    return headers
}

func headers(w http.ResponseWriter, r *http.Request) {
    response, err := json.MarshalIndent(map[string]Headers{"headers": getHeaders(r)}, "", "  ")
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(w, string(response))
}

func getArgs(r *http.Request)map[string]string {
    args := map[string]string{}
    for k, v := range r.URL.Query() {
        args[k] = v[0]
    }
    return args
}

func get(w http.ResponseWriter, r *http.Request) {
    res := buildResponseDict(r, []string{"headers", "url", "args", "origin"})
    response, err := json.MarshalIndent(res, "", "  ")
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(w, string(response))
}

func gzipHandler(w http.ResponseWriter, r *http.Request) {
    res := buildResponseDict(r, []string{"headers", "origin", "gzipped", "method"})
    response, err := json.MarshalIndent(res, "", "  ")
    if err != nil { log.Fatal(err) }
    w.Header().Set("Content-Encoding", "gzip")
    w.Header().Set("Content-Type", "application/json")
    writer := gzip.NewWriter(w)
    if err != nil { log.Fatal(err) }
    fmt.Fprintf(writer, string(response))
    defer writer.Close()
}

func buildResponseDict(r *http.Request, items []string) ResponseDict {
    res := make(ResponseDict)
    for _, item := range items {
        switch item {
            case "headers": res[item] = getHeaders(r)
            case "url": res[item] = "http://" + r.Host + r.URL.String()
            case "args": res[item] = getArgs(r)
            case "origin": res[item] = getIp(r)
            case "gzipped": res[item] = true
            case "method": res[item] = r.Method
        }
    }
    return res
}

func main() {
    http.HandleFunc("/", homepage)
    http.HandleFunc("/ip", ip)
    http.HandleFunc("/user-agent", useragent)
    http.HandleFunc("/headers", headers)
    http.HandleFunc("/get", get)
    http.HandleFunc("/gzip", gzipHandler)

    fmt.Printf("Listening on port 8000...\n")
    err := http.ListenAndServe(":8000", nil)
    if err != nil { log.Fatal(err) }
}
