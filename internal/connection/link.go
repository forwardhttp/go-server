package connection

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Link struct {
	Hash        string `json:"hash"`
	mux         *sync.Mutex
	conn        *websocket.Conn
	message     chan []byte
	lastMessage time.Time

	// Connected can be set to true and ignored for now
	// The idea is to cache requests, up to like 10 or 20
	// for a client and if the same connection is reopened
	// within a TBD amount of time, we can replay those requests to
	// the client
	connected bool
}

func (l *Link) URIsFromHash(broker *url.URL) (*url.URL, *url.URL) {

	var openURI = new(url.URL)
	*openURI = *broker
	openURI.Path = fmt.Sprintf("open/%s", l.Hash)

	var requestURI = new(url.URL)
	*requestURI = *broker
	requestURI.Path = l.Hash

	return openURI, requestURI

}
