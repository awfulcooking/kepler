package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

var addr = flag.String("addr", "localhost:1969", "http service address")
var timeout = flag.Duration("timeout", 10*time.Second, "maximum time allowed for requests in flight")
var remotePort = flag.Int("remotePort", 1965, "TCP port that clients will proxy to")

// Whether to disable HTTP origin checking. Note that non-browser clients can spoof their origin,
// but it's a good protection if you want to restrict which websites can use your proxy.
var anyOrigin = flag.Bool("anyOrigin", false, "allow connections from any HTTP origin. (makes this an open proxy usable from any website!)")

// When not using -anyOrigin, allowed origins must be passed as trailing arguments to kepler
var allowedOrigins []string

func main() {
	flag.Parse()
	allowedOrigins = flag.Args()

	if !*anyOrigin && len(allowedOrigins) == 0 {
		os.Stderr.WriteString("must pass allowed HTTP origins as arguments, e.g:\n")
		os.Stderr.WriteString("\t" + os.Args[0] + " -addr <addr> '*.mydomain.com' other.site\n")
		os.Stderr.WriteString("or, use -anyOrigin.\n")
		os.Exit(64)
	}

	http.HandleFunc("/", serve)

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func serve(w http.ResponseWriter, r *http.Request) {
	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: *anyOrigin,
		OriginPatterns: allowedOrigins,
	})

	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), *timeout)
	defer cancel()

	client := websocket.NetConn(ctx, wsConn, websocket.MessageBinary)

	destination, err := url.Parse("gemini://" + r.URL.Path[1:])
	if err != nil {
		println("bad url in path")
		return
	}

	if destination.Port() == "" {
		destination.Host += ":1965"
	}
	if destination.Port() != "1965" {
		println("bad port requested")
	}

	println("destination is", destination.Host, destination.String())

	remote, err := net.Dial("tcp", destination.Host)
	if err != nil {
		println("can't dial upstream", err)
		return
	}

	proxy(client, remote)

	wsConn.Close(websocket.StatusNormalClosure, "")
}

func proxy(client net.Conn, remote net.Conn) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go _copy(wg, remote, client, "to remote")
	go _copy(wg, client, remote, "to client")
	wg.Wait()
}

func _copy(wg *sync.WaitGroup, dest net.Conn, src net.Conn, dir string) (int64, error) {
	n, err := io.Copy(dest, src)
	dest.Close()

	fmt.Println("copied", n, "bytes", dir, err)

	wg.Done()
	return n, err
}

