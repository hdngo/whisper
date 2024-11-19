package handler

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/hdngo/whisper/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // NOTE: Allow all connections by default, change in production
	},
	Subprotocols: []string{"access_token"},
}

type ChatHandler struct {
	hub       *ws.Hub
	jwtSecret string
}

func NewChatHandler(hub *ws.Hub, jwtSecret string) *ChatHandler {
	return &ChatHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Instead of getting token from header, get it from Sec-WebSocket-Protocol
	protocols := r.Header.Get("Sec-WebSocket-Protocol")
	if protocols == "" {
		http.Error(w, "no token provided", http.StatusUnauthorized)
		return
	}

	// The protocol string will be in format: "access_token|<actual_token>"
	protocolParts := strings.Split(protocols, "|")
	if len(protocolParts) != 2 || protocolParts[0] != "access_token" {
		http.Error(w, "invalid protocol", http.StatusUnauthorized)
		return
	}

	token := protocolParts[1]

	// Parse and validate the token
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userID := int64(claims["user_id"].(float64))
	username := claims["username"].(string)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
		return
	}

	client := ws.NewClient(h.hub, conn, userID, username)
	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
