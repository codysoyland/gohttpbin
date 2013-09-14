package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Headers map[string]string
type ResponseDict map[string]interface{}
type L []string

func getIp(r *http.Request) string {
	return r.RemoteAddr[0 : len(r.RemoteAddr)-strings.LastIndex(r.RemoteAddr, ":")-1]
}

func getHeaders(r *http.Request) Headers {
	headers := Headers{}
	for k, v := range r.Header {
		headers[k] = v[0]
	}
	return headers
}

func getArgs(r *http.Request) map[string]string {
	args := map[string]string{}
	for k, v := range r.URL.Query() {
		args[k] = v[0]
	}
	return args
}

func getCookies(r *http.Request) map[string]string {
	cookies := r.Cookies()
	cookie_map := make(map[string]string)
	for _, cookie := range cookies {
		cookie_map[cookie.Name] = cookie.Value
	}
	return cookie_map
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "400 bad request", http.StatusBadRequest)
}

func buildResponseDict(r *http.Request, items L) ResponseDict {
	res := make(ResponseDict)
	for _, item := range items {
		switch item {
		case "headers":
			res[item] = getHeaders(r)
		case "url":
			res[item] = "http://" + r.Host + r.URL.String()
		case "args":
			res[item] = getArgs(r)
		case "user-agent":
			res[item] = r.UserAgent()
		case "origin":
			res[item] = getIp(r)
		case "gzipped":
			res[item] = true
		case "method":
			res[item] = r.Method
		case "form":
			res[item] = make(map[string]string) // TODO: implement
		case "data":
			res[item] = "" // TODO: implement
		case "files":
			res[item] = make(map[string]string) // TODO: implement
		case "json":
			res[item] = nil // TODO: implement
		case "cookies":
			res[item] = getCookies(r)
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

func respondJson(w http.ResponseWriter, response interface{}) {
	response_text, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
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

var templates = template.Must(template.ParseFiles("index.html"))

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
	} else {
		templates.ExecuteTemplate(w, "index.html", "")
	}
}

func ipHandler(w http.ResponseWriter, r *http.Request) {
	respondJson(w, buildResponseDict(r, L{"origin"}))
}

func useragentHandler(w http.ResponseWriter, r *http.Request) {
	respondJson(w, buildResponseDict(r, L{"user-agent"}))
}

func headersHandler(w http.ResponseWriter, r *http.Request) {
	respondJson(w, buildResponseDict(r, L{"headers"}))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "GET") {
		return
	}
	respondJson(w, buildResponseDict(r, L{"headers", "url", "args", "origin"}))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "POST") {
		return
	}
	respondJson(w, buildResponseDict(r, L{"url", "args", "form", "data", "origin", "headers", "files", "json"}))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "PUT") {
		return
	}
	respondJson(w, buildResponseDict(r, L{"url", "args", "form", "data", "origin", "headers", "files", "json"}))
}

func patchHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "PATCH") {
		return
	}
	respondJson(w, buildResponseDict(r, L{"url", "args", "form", "data", "origin", "headers", "files", "json"}))
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "DELETE") {
		return
	}
	respondJson(w, buildResponseDict(r, L{"url", "args", "data", "origin", "headers", "json"}))
}

func gzipHandler(w http.ResponseWriter, r *http.Request) {
	res := buildResponseDict(r, L{"headers", "origin", "gzipped", "method"})
	response, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	writer := gzip.NewWriter(w)
	writer.Write(response)
	writer.Close()
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`/status/(\d{1,3})$`)
	match := re.FindStringSubmatch(r.URL.String())
	if len(match) == 0 {
		http.NotFound(w, r)
		return
	}
	status, err := strconv.ParseInt(match[1], 10, 16)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(int(status))
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`/stream/(\d{1,9}).*`)
	match := re.FindStringSubmatch(r.URL.String())
	if len(match) == 0 {
		http.NotFound(w, r)
		return
	}
	res := buildResponseDict(r, L{"url", "args", "headers", "origin"})

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "application/json")

	count, err := strconv.Atoi(match[1])
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < count; i++ {
		res["id"] = i
		response, err := json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}
		response = append(response, byte(10))

		w.Write(response)
		w.(http.Flusher).Flush()
	}
}

func delayHandler(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`/delay/(\d{1,2}).*`)
	match := re.FindStringSubmatch(r.URL.String())
	if len(match) == 0 {
		http.NotFound(w, r)
		return
	}
	length, err := strconv.Atoi(match[1])
	if err != nil {
		log.Fatal(err)
	}
	if length > 10 {
		length = 10
	}
	time.Sleep(time.Second * time.Duration(length))
	respondJson(w, buildResponseDict(r, L{"url", "args", "form", "data", "origin", "headers", "files"}))
}

func responseHeaderHandler(w http.ResponseWriter, r *http.Request) {
	args := getArgs(r)
	for k, v := range args {
		w.Header().Set(k, v)
	}
	respondJson(w, getArgs(r))
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`/([\w-]+)/(\d{1,3}).*`)
	match := re.FindStringSubmatch(r.URL.String())
	if len(match) == 0 {
		http.NotFound(w, r)
		return
	}
	num, err := strconv.Atoi(match[2])
	if err != nil {
		log.Fatal(err)
	}
	var url string
	if num <= 1 {
		url = "/get"
	} else {
		url = "/" + match[1] + "/" + strconv.Itoa(num-1)
	}
	http.Redirect(w, r, url, 302)
}

func redirectToHandler(w http.ResponseWriter, r *http.Request) {
	args := getArgs(r)
	if url, ok := args["url"]; ok {
		http.Redirect(w, r, url, 302)
	} else {
		BadRequest(w, r)
	}
}

func cookiesHandler(w http.ResponseWriter, r *http.Request) {
	respondJson(w, buildResponseDict(r, L{"cookies"}))
}

func setCookiesHandler(w http.ResponseWriter, r *http.Request) {
	cookies := getArgs(r)
	for k, v := range cookies {
		http.SetCookie(w, &http.Cookie{Name: k, Value: v, Path: "/"})
	}
	http.Redirect(w, r, "/cookies", 302)
}

func deleteCookiesHandler(w http.ResponseWriter, r *http.Request) {
	cookies := getArgs(r)
	for k, _ := range cookies {
		http.SetCookie(w, &http.Cookie{Name: k, Expires: time.Unix(1, 0), Path: "/", MaxAge: -1})
	}
	http.Redirect(w, r, "/cookies", 302)
}

func basicAuthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

func digestAuthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

func denyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

func cacheHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
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
	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/response-headers", responseHeaderHandler)
	http.HandleFunc("/redirect/", redirectHandler)
	http.HandleFunc("/redirect-to", redirectToHandler)
	http.HandleFunc("/relative-redirect/", redirectHandler) // TODO: redirect handler is already relative. Decide how this should work.
	http.HandleFunc("/cookies", cookiesHandler)
	http.HandleFunc("/cookies/set", setCookiesHandler)
	http.HandleFunc("/cookies/delete", deleteCookiesHandler)
	http.HandleFunc("/basic-auth/", basicAuthHandler)
	http.HandleFunc("/digest-auth/", digestAuthHandler)
	http.HandleFunc("/stream/", streamHandler)
	http.HandleFunc("/delay/", delayHandler)
	http.HandleFunc("/html", htmlHandler)
	http.HandleFunc("/robots.txt", robotsHandler)
	http.HandleFunc("/deny", denyHandler)
	http.HandleFunc("/cache", cacheHandler)


	fmt.Printf("Listening on port 8000...\n")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
