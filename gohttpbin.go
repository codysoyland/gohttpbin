package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var templates = template.Must(template.ParseFiles("index.html"))

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
	} else {
		templates.ExecuteTemplate(w, "index.html", "")
	}
}

func ipHandler(w http.ResponseWriter, r *http.Request) {
	RespondInfo(r, w, L{"origin"})
}

func useragentHandler(w http.ResponseWriter, r *http.Request) {
	RespondInfo(r, w, L{"user-agent"})
}

func headersHandler(w http.ResponseWriter, r *http.Request) {
	RespondInfo(r, w, L{"headers"})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckMethod(w, r, "GET") {
		return
	}
	RespondInfo(r, w, L{"headers", "url", "args", "origin"})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckMethod(w, r, "POST") {
		return
	}
	RespondInfo(r, w, L{"url", "args", "form", "data", "origin", "headers", "files", "json"})
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckMethod(w, r, "PUT") {
		return
	}
	RespondInfo(r, w, L{"url", "args", "form", "data", "origin", "headers", "files", "json"})
}

func patchHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckMethod(w, r, "PATCH") {
		return
	}
	RespondInfo(r, w, L{"url", "args", "form", "data", "origin", "headers", "files", "json"})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckMethod(w, r, "DELETE") {
		return
	}
	RespondInfo(r, w, L{"url", "args", "data", "origin", "headers", "json"})
}

func gzipHandler(w http.ResponseWriter, r *http.Request) {
	res := BuildResponseDict(r, L{"headers", "origin", "gzipped", "method"})
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
	res := BuildResponseDict(r, L{"url", "args", "headers", "origin"})

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
	RespondInfo(r, w, L{"url", "args", "form", "data", "origin", "headers", "files"})
}

func responseHeaderHandler(w http.ResponseWriter, r *http.Request) {
	args := GetArgs(r)
	for k, v := range args {
		w.Header().Set(k, v)
	}
	RespondJson(w, GetArgs(r))
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
	args := GetArgs(r)
	if url, ok := args["url"]; ok {
		http.Redirect(w, r, url, 302)
	} else {
		BadRequest(w, r)
	}
}

func cookiesHandler(w http.ResponseWriter, r *http.Request) {
	RespondInfo(r, w, L{"cookies"})
}

func setCookiesHandler(w http.ResponseWriter, r *http.Request) {
	cookies := GetArgs(r)
	for k, v := range cookies {
		http.SetCookie(w, &http.Cookie{Name: k, Value: v, Path: "/"})
	}
	http.Redirect(w, r, "/cookies", 302)
}

func deleteCookiesHandler(w http.ResponseWriter, r *http.Request) {
	cookies := GetArgs(r)
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
