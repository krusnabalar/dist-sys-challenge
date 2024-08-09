package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var encoding = base32.NewEncoding("0123456789ABCDEFGHIJKLMNOPQRSTUV").WithPadding(base32.NoPadding)

type server struct {
	node         *maelstrom.Node
	nodeID       []byte
	counter      uint32
	counterMutex sync.Mutex
}

func main() {
	n := maelstrom.NewNode()
	s := &server{node: n, nodeID: make([]byte, 4)}
	rand.Read(s.nodeID)

	n.Handle("generate", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "generate_ok",
			"id":   s.generateKSUID(),
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s *server) generateKSUID() string {
	// Based on https://github.com/segmentio/ksuid
	ksuid := make([]byte, 28)
	binary.BigEndian.PutUint32(ksuid[:4], uint32(time.Now().Unix()))
	s.counterMutex.Lock()
	binary.BigEndian.PutUint32(ksuid[4:8], s.counter)
	s.counter++
	s.counterMutex.Unlock()
	copy(ksuid[8:12], s.nodeID)
	rand.Read(ksuid[12:])
	return encoding.EncodeToString(ksuid)
}
