package main

import (
	"fmt"
	"github.com/nu7hatch/gouuid"
	"github.com/soundtrackyourbrand/pusher/client"
	"log"
	"net/http"
	"strings"
	"time"
)

var token = "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODEyMDc4NjMsImlhdCI6MTQ4MDYwMzA2Mywic3ViIjoiVlhObGNpd3NNV3N4Ym00M1pqSnVjM2N2IiwidHlwIjoic3VwZXJ1c2VyIn0.hY5MOilbtyqoDDPoHDOBEcEwcThYpgbxuti2aZWXH6NrVd_WBXzmb2U4qxcNIIISdw4oY0UgJnWpZbdyqsvTYQ"
var proxy_endpoint = "/devices/RGV2aWNlLCwxbWptcjU3aDVhOC8./rpcclient/rpcserver/webserver"
var input_endpoint = "/devices/RGV2aWNlLCwxbWptcjU3aDVhOC8./rpcserver/rpcclient/"

// {"Type":"Message","URI":"/devices/RGV2aWNlLCwxbWptcjU3aDVhOC8./rpcclient/rpcserver/webserver","Data":[{"uri":"/logfilter.html","channel":"f4309837-4087-1000-0bea-b1a94acfb63d"}],"Id":"H1e77lWU21pBddNhPy4c-Q==:4"}

type HtmlResponse struct {
	Body        string
	Code        int
	ContentType string
}

type HtmlRequest struct {
	Uri     string `json:"uri"`
	Channel string `json:"channel"`
	pipe    chan HtmlResponse
}

func startWebserver(send_req func(*HtmlRequest)) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()       // parse arguments, you have to call this by yourself
		fmt.Println(r.Form) // print form information in server side
		fmt.Println("path", r.URL.Path)
		fmt.Println("scheme", r.URL.Scheme)
		fmt.Println(r.Form["url_long"])
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}

		// Subscribe to a new channel
		channel, _ := uuid.NewV4()
		proxy_req := &HtmlRequest{Uri: r.URL.Path, Channel: channel.String()}
		send_req(proxy_req)
		select {
		case res, _ := <-proxy_req.pipe:
			fmt.Fprintf(w, res.Body)
			w.WriteHeader(res.Code)
			if len(res.ContentType) > 0 {
				w.Header().Set("Content-Type", res.ContentType)
			} else {
				w.Header().Set("Content-Type", "text/html")
			}
		case <-time.After(10 * time.Second):
			fmt.Fprintf(w, "Timeout accessing client")
			w.WriteHeader(500)
		}
	}) // set router
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	client := &client.Client{}

	client.Connect("http://localhost/", "wss://event.api.soundtrackyourbrand.com/")
	client.Authorize(proxy_endpoint, token)
	fmt.Println("Starting webserver :9090")

	dest := map[string]*HtmlRequest{}
	go startWebserver(func(proxy_req *HtmlRequest) {
		uri := input_endpoint + proxy_req.Channel
		proxy_req.pipe = make(chan HtmlResponse)
		client.Authorize(uri, token)
		client.Subscribe(uri)
		dest[uri] = proxy_req
		client.Send(proxy_endpoint, []interface{}{proxy_req})
	})

	for {
		entry := client.Next()
		data, s1 := entry.Data.(map[string]interface{})
		if s1 {
			body, _ := data["body"].(string)
			status, _ := data["status"].(float64)
			content_type, _ := data["content_type"].(string)
			response := HtmlResponse{body, int(status), content_type}
			dest[entry.URI].pipe <- response
			client.Unsubscribe(entry.URI)
			delete(dest, entry.URI)
		}
	}
}
