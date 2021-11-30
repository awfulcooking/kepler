Kepler
===

A little websocket TCP proxy built to let [Amfora](https://github.com/makeworld-the-better-one/amfora) talk to Gemini servers when running in a browser.

Usage
---

```sh
$ git clone https://github.com/awfulcooking/kepler
$ cd kepler
$ go build
```

```
Usage of ./kepler:
  -addr string
    	http service address (default "localhost:1969")
  -anyOrigin
    	allow connections from any HTTP origin. (makes this an open proxy usable from any website!)
  -remotePort int
    	TCP port that clients will proxy to (default 1965)
  -timeout duration
    	maximum time allowed for requests in flight (default 10s)
```

Web clients should dial wss://\<kepler-addr>/\<hostname> to open a proxied socket to \<hostname>:\<remotePort>

License
---

MIT License.
