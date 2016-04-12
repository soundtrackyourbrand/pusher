package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"runtime"

	"github.com/soundtrackyourbrand/pusher/hub"
)

func main() {
	var (
		local    = flag.String("local", "", "serve as webserver, example: 0.0.0.0:8000")
		tcp      = flag.String("tcp", "", "serve as FCGI via TCP, example: 0.0.0.0:8000")
		unix     = flag.String("unix", "", "serve as FCGI via UNIX socket, example: /tmp/myprogram.sock")
		loglevel = flag.Int("loglevel", 0, "How much to log")
		numprocs = flag.Int("numprocs", runtime.NumCPU(), "How many processes to run")
	)
	flag.Parse()
	runtime.GOMAXPROCS(*numprocs)
	if *loglevel > 1 {
		fmt.Printf("Processes: %d\n", *numprocs)
	}

	mux := http.NewServeMux()
	// mux.Handle("/", http.FileServer(http.Dir("./js"))
	mux.Handle("/", http.FileServer(http.Dir("./")))
	mux.Handle("/ws", hub.NewServer().Loglevel(*loglevel))
	if *local != "" { // Run as a local web server
		err := http.ListenAndServe(*local, mux)
		if err != nil {
			log.Fatal(err)
		}
		if *loglevel > 0 {
			fmt.Println("Listening on local", *local)
		}
	} else if *tcp != "" { // Run as FCGI via TCP
		listener, err := net.Listen("tcp", *tcp)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Listening on tcp", *tcp)
		defer listener.Close()

		err = fcgi.Serve(listener, mux)
		if err != nil {
			log.Fatal(err)
		}
	} else if *unix != "" { // Run as FCGI via UNIX socket
		listener, err := net.Listen("unix", *unix)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Listening on unix", *unix)
		defer listener.Close()

		err = fcgi.Serve(listener, mux)
		if err != nil {
			log.Fatal(err)
		}
	} else { // Run as FCGI via standard I/O
		err := fcgi.Serve(nil, mux)
		if err != nil {
			log.Fatal(err)
		}
	}

}
