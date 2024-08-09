package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type server struct {
	messageSet map[int]struct{}
	msgMutex   sync.RWMutex
}

func main() {
	s := &server{messageSet: make(map[int]struct{})}
	n := maelstrom.NewNode()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		message := int(body["message"].(float64))
		s.msgMutex.Lock()
		if _, exists := s.messageSet[message]; exists {
			s.msgMutex.Unlock()
			return nil
		}
		s.messageSet[message] = struct{}{}
		s.msgMutex.Unlock()

		neighbors := n.NodeIDs()
		for _, neighbor := range neighbors {
			if neighbor != msg.Src && neighbor != n.ID() {
				if err := n.Send(neighbor, body); err != nil {
					log.Fatal(err)
				}
			}
		}
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		s.msgMutex.RLock()
		messages := make([]int, len(s.messageSet))
		for message := range s.messageSet {
			messages = append(messages, message)
		}
		s.msgMutex.RUnlock()
		return n.Reply(msg, map[string]any{"type": "read_ok", "messages": messages})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
