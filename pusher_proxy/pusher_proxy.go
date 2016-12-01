package main

import (
	"fmt"
	"github.com/soundtrackyourbrand/pusher/client"
	"github.com/soundtrackyourbrand/pusher/hub"
	"log"
	"net/http"
	"strings"
)

var token = "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODEyMDc4NjMsImlhdCI6MTQ4MDYwMzA2Mywic3ViIjoiVlhObGNpd3NNV3N4Ym00M1pqSnVjM2N2IiwidHlwIjoic3VwZXJ1c2VyIn0.hY5MOilbtyqoDDPoHDOBEcEwcThYpgbxuti2aZWXH6NrVd_WBXzmb2U4qxcNIIISdw4oY0UgJnWpZbdyqsvTYQ"
var proxy_endpoint = "/devices/RGV2aWNlLCwxbWptcjU3aDVhOC8./rpcclient/rpcserver/webserver"
var input_endpoint = "/devices/RGV2aWNlLCwxbWptcjU3aDVhOC8./rpcserver/rpcclient/"

// {"Type":"Message","URI":"/devices/RGV2aWNlLCwxbWptcjU3aDVhOC8./rpcclient/rpcserver/webserver","Data":[{"uri":"/logfilter.html","channel":"f4309837-4087-1000-0bea-b1a94acfb63d"}],"Id":"H1e77lWU21pBddNhPy4c-Q==:4"}

type HtmlRequest struct {
	Uri     string `json:"uri"`
	Channel string `json:"channe"`
}

func startWebserver(client *client.Client) {
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
		proxy_req := HtmlRequest{Uri: "/logfilter.html", Channel: "xxx"}
		client.Authorize(input_endpoint+proxy_req.Channel, token)
		client.Subscribe(input_endpoint + proxy_req.Channel)
		client.Send(proxy_endpoint, proxy_req)
		entry := client.Next(hub.TypeMessage)
		fmt.Fprintf(w, "%#v", entry) // send data to client side

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

	go startWebserver(client)

	for {
		//entry := client.Next("ALL")
		//fmt.Println("%#v\n", entry)
	}
}
