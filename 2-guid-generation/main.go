package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var encoding = base32.NewEncoding("0123456789ABCDEFGHIJKLMNOPQRSTUV").WithPadding(base32.NoPadding)

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "generate_ok",
			"id":   generateKSUID(),
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func generateKSUID() string {
	// Based on https://github.com/segmentio/ksuid
	timestamp := time.Now().Unix()
	payload := make([]byte, 16)
	_, err := rand.Read(payload)
	if err != nil {
		log.Fatal(err)
	}

	ksuid := make([]byte, 20)
	binary.BigEndian.PutUint32(ksuid[:4], uint32(timestamp))
	copy(ksuid[4:], payload)

	return encoding.EncodeToString(ksuid)
}
