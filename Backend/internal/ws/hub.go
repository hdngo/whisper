package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/hdngo/whisper/internal/model"
	"github.com/hdngo/whisper/internal/repository"
)

type Hub struct {
	clients    sync.Map
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	msgRepo    *repository.MessageRepository
	mutex      sync.RWMutex
	done       chan struct{}
}

func NewHub(msgRepo *repository.MessageRepository) *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		msgRepo:    msgRepo,
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Hub recovered from panic: %v", r)
		}
	}()

	for {
		select {
		case client := <-h.Register:
			h.handleRegister(client)
		case client := <-h.Unregister:
			h.handleUnregister(client)
		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		case <-h.done:
			return
		}
	}
}

func (h *Hub) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	close(h.done)

	h.clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		client.conn.Close()
		return true
	})
}

func (h *Hub) handleRegister(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients.Store(client, true)

	wsMsg := &model.WSMessage{
		Type: model.MessageTypeJoin,
		Payload: map[string]string{
			"username": client.username,
		},
	}
	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		go h.broadcast(msgBytes)
	}

	go h.broadcastOnlineUsers()
}

func (h *Hub) handleUnregister(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients.LoadAndDelete(client); ok {
		close(client.send)

		wsMsg := &model.WSMessage{
			Type: model.MessageTypeLeave,
			Payload: map[string]string{
				"username": client.username,
			},
		}
		if msgBytes, err := json.Marshal(wsMsg); err == nil {
			go h.broadcast(msgBytes)
		}

		go h.broadcastOnlineUsers()
	}
}

func (h *Hub) handleBroadcast(message []byte) {
	var wsMsg model.WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("error unmarshalling message: %v", err)
		return
	}

	if wsMsg.Type == model.MessageTypeChat {
		go h.storeMessage(wsMsg)
	}

	go h.broadcast(message)
}

func (h *Hub) storeMessage(wsMsg model.WSMessage) {
	payload, ok := wsMsg.Payload.(map[string]interface{})
	if !ok {
		log.Printf("invalid message payload")
		return
	}

	msg := &model.Message{
		Content:  payload["content"].(string),
		UserID:   int64(payload["user_id"].(float64)),
		Username: payload["username"].(string),
	}

	if err := h.msgRepo.Create(context.Background(), msg); err != nil {
		log.Printf("error storing message: %v", err)
	}
}

func (h *Hub) broadcast(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	h.clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		select {
		case client.send <- message:
		default:
			h.clients.Delete(client)
			close(client.send)
		}
		return true
	})
}

func (h *Hub) broadcastOnlineUsers() {
	users := make(map[string]bool)

	h.clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		users[client.username] = true
		return true
	})

	uniqueUsers := make([]string, 0, len(users))
	for username := range users {
		uniqueUsers = append(uniqueUsers, username)
	}

	wsMsg := &model.WSMessage{
		Type:    model.MessageTypeUsers,
		Payload: uniqueUsers,
	}

	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		go h.broadcast(msgBytes)
	}
}
