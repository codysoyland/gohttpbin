package main

import (
	"bytes"
	"io"
	"strings"
	"net/http"
	"fmt"
	"strconv"
	"encoding/json"
	"log"
)

type Headers map[string]string
type ResponseDict map[string]interface{}
type L []string

func getHeaders(r *http.Request) Headers {
	headers := Headers{}
	for k, v := range r.Header {
		headers[k] = v[0]
	}
	return headers
}

func getCookies(r *http.Request) map[string]string {
	cookies := r.Cookies()
	cookie_map := make(map[string]string)
	for _, cookie := range cookies {
		cookie_map[cookie.Name] = cookie.Value
	}
	return cookie_map
}

func GetArgs(r *http.Request) map[string]string {
	args := map[string]string{}
	for k, v := range r.URL.Query() {
		args[k] = v[0]
	}
	return args
}

func getIp(r *http.Request) string {
	return r.RemoteAddr[0 : len(r.RemoteAddr)-strings.LastIndex(r.RemoteAddr, ":")-1]
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "400 bad request", http.StatusBadRequest)
}

func BuildResponseDict(r *http.Request, items L) ResponseDict {
	res := make(ResponseDict)
	for _, item := range items {
		switch item {
		case "headers":
			res[item] = getHeaders(r)
		case "url":
			res[item] = "http://" + r.Host + r.URL.String()
		case "args":
			res[item] = GetArgs(r)
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

func Respond(w http.ResponseWriter, content string, headers Headers) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, content)
	headers["Content-Length"] = strconv.Itoa(buf.Len())
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	io.Copy(w, &buf)
}

func RespondJson(w http.ResponseWriter, response interface{}) {
	response_text, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	headers := make(Headers)
	headers["Content-Type"] = "application/json"
	Respond(w, string(response_text), headers)
}

func CheckMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		w.WriteHeader(405)
		return false
	}
	return true
}

func RespondInfo(r *http.Request, w http.ResponseWriter, l L) {
	RespondJson(w, BuildResponseDict(r, l))
}
