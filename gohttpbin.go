package main

import (
    "fmt"
    "log"
    "net/http"
    "strings"
    "encoding/json"
    "compress/gzip"
    "regexp"
    "strconv"
    "time"
    "bytes"
    "io"
)

type Headers map[string]string
type ResponseDict map[string]interface{}

func getIp(r *http.Request)string {
    return r.RemoteAddr[0:len(r.RemoteAddr)-strings.LastIndex(r.RemoteAddr, ":")-1]
}

func getHeaders(r *http.Request)Headers {
    headers := Headers{}
    for k, v := range r.Header {
        headers[k] = v[0]
    }
    return headers
}

func getArgs(r *http.Request)map[string]string {
    args := map[string]string{}
    for k, v := range r.URL.Query() {
        args[k] = v[0]
    }
    return args
}

func buildResponseDict(r *http.Request, items []string) ResponseDict {
    res := make(ResponseDict)
    for _, item := range items {
        switch item {
            case "headers": res[item] = getHeaders(r)
            case "url": res[item] = "http://" + r.Host + r.URL.String()
            case "args": res[item] = getArgs(r)
            case "user-agent": res[item] = r.UserAgent()
            case "origin": res[item] = getIp(r)
            case "gzipped": res[item] = true
            case "method": res[item] = r.Method
            case "form": res[item] = make(map[string]string) // TODO: implement
            case "data": res[item] = "" // TODO: implement
            case "files": res[item] = make(map[string]string) // TODO: implement
            case "json": res[item] = nil // TODO: implement
        }
    }
    return res
}

func respond(w http.ResponseWriter, content string, headers Headers) {
    var buf bytes.Buffer
    fmt.Fprintf(&buf, content)
    headers["Content-Length"] = strconv.Itoa(buf.Len())
    for k, v := range headers {
        w.Header().Set(k, v)
    }
    io.Copy(w, &buf)
}

func respondJson(w http.ResponseWriter, response ResponseDict) {
    response_text, err := json.MarshalIndent(response, "", "  ")
    if err != nil { log.Fatal(err) }
    var buf bytes.Buffer
    fmt.Fprintf(&buf, string(response_text))
    headers := make(Headers)
    headers["Content-Type"] = "application/json"
    respond(w, string(response_text), headers)
}

func checkMethod(w http.ResponseWriter, r *http.Request, method string) bool {
    if r.Method != method {
        w.WriteHeader(405)
        return false
    }
    return true
}

func homepageHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to gohttpbin!") // TODO: rip off httpbin.org :P
}

func ipHandler(w http.ResponseWriter, r *http.Request) {
    respondJson(w, buildResponseDict(r, []string{"origin"}))
}

func useragentHandler(w http.ResponseWriter, r *http.Request) {
    respondJson(w, buildResponseDict(r, []string{"user-agent"}))
}

func headersHandler(w http.ResponseWriter, r *http.Request) {
    respondJson(w, buildResponseDict(r, []string{"headers"}))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    if !checkMethod(w, r, "GET") { return }
    respondJson(w, buildResponseDict(r, []string{"headers", "url", "args", "origin"}))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
    if !checkMethod(w, r, "POST") { return }
    respondJson(w, buildResponseDict(r, []string{"url", "args", "form", "data", "origin", "headers", "files", "json"}))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
    if !checkMethod(w, r, "PUT") { return }
    respondJson(w, buildResponseDict(r, []string{"url", "args", "form", "data", "origin", "headers", "files", "json"}))
}

func patchHandler(w http.ResponseWriter, r *http.Request) {
    if !checkMethod(w, r, "PATCH") { return }
    respondJson(w, buildResponseDict(r, []string{"url", "args", "form", "data", "origin", "headers", "files", "json"}))
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
    if !checkMethod(w, r, "DELETE") { return }
    respondJson(w, buildResponseDict(r, []string{"url", "args", "data", "origin", "headers", "json"}))
}

func gzipHandler(w http.ResponseWriter, r *http.Request) {
    res := buildResponseDict(r, []string{"headers", "origin", "gzipped", "method"})
    response, err := json.MarshalIndent(res, "", "  ")
    if err != nil { log.Fatal(err) }
    buf := new(bytes.Buffer)
    writer := gzip.NewWriter(buf)
    fmt.Fprint(writer, string(response))
    writer.Close()
    str := buf.String()
    headers := Headers{
        "Content-Encoding": "gzip",
        "Content-Type": "application/json",
    }
    respond(w, str, headers)
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
    re := regexp.MustCompile(`/stream/(\d{1,9}).*`)
    match := re.FindStringSubmatch(r.URL.String())
    if len(match) == 0 {
        w.WriteHeader(404)
        return
    }
    res := buildResponseDict(r, []string{"url", "args", "headers", "origin"})

    w.Header().Set("Transfer-Encoding", "chunked")
    w.Header().Set("Content-Type", "application/json")

    count, err := strconv.Atoi(match[1])
    if err != nil { log.Fatal(err) }
    for i := 0; i < count; i++ {
        res["id"] = i
        response, err := json.Marshal(res)
        if err != nil { log.Fatal(err) }
        response = append(response, byte(10))

        w.Write(response)
        w.(http.Flusher).Flush()
    }
}

func delayHandler(w http.ResponseWriter, r *http.Request) {
    re := regexp.MustCompile(`/delay/(\d{1,2}).*`)
    match := re.FindStringSubmatch(r.URL.String())
    if len(match) == 0 {
        w.WriteHeader(404)
        return
    }
    length, err := strconv.Atoi(match[1])
    if err != nil { log.Fatal(err) }
    if length > 10 {
        length = 10
    }
    time.Sleep(time.Second*time.Duration(length))
    respondJson(w, buildResponseDict(r, []string{"url", "args", "form", "data", "origin", "headers", "files"}))
}

func main() {
    http.HandleFunc("/", homepageHandler)
    http.HandleFunc("/ip", ipHandler)
    http.HandleFunc("/user-agent", useragentHandler)
    http.HandleFunc("/headers", headersHandler)
    http.HandleFunc("/get", getHandler)
    http.HandleFunc("/post", postHandler)
    http.HandleFunc("/put", putHandler)
    http.HandleFunc("/patch", patchHandler)
    http.HandleFunc("/delete", deleteHandler)
    http.HandleFunc("/gzip", gzipHandler)
    http.HandleFunc("/stream/", streamHandler)
    http.HandleFunc("/delay/", delayHandler)

    fmt.Printf("Listening on port 8000...\n")
    err := http.ListenAndServe(":8000", nil)
    if err != nil { log.Fatal(err) }
}
