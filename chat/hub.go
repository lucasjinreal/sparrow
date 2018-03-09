package chat

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/go-clog/clog"
	"github.com/gorilla/websocket"
)

// RoomHub chat room controller
type RoomHub struct {
	rooms    sync.Map         // room list
	clients  sync.Map         // clients not join any room yet
	handlers []MessageHandler // onmessage handler
	quit     chan struct{}
}

// NewChatHub create an new chat room
func NewChatHub() *RoomHub {
	return &RoomHub{
		quit:     make(chan struct{}, 1),
		handlers: make([]MessageHandler, 0),
	}
}

func (h *RoomHub) run() {
}

// NewRoom create room or return exist one
func (h *RoomHub) NewRoom(name string) *RoomInfo {
	if v, ok := h.rooms.Load(name); ok {
		r, _ := v.(*room)
		return &r.RoomInfo
	}

	r := newRoom(name, h)
	h.rooms.Store(r.ID, r)
	return &r.RoomInfo
}

// LoadRooms load rooms from database
func (h *RoomHub) LoadRooms() {
	// Query From DB
}

// GetRoom return given room information
func (h *RoomHub) GetRoom(roomID string) *RoomInfo {
	if v, ok := h.rooms.Load(roomID); ok {
		r, _ := v.(*room)
		return &r.RoomInfo
	}
	return nil
}

// DeleteRoom delete room from room hub
func (h *RoomHub) DeleteRoom(roomID string) *RoomInfo {
	if v, ok := h.rooms.Load(roomID); ok {
		h.rooms.Delete(roomID)
		r, _ := v.(*room)
		return &r.RoomInfo
	}
	return nil
}

// JoinRoom client join room
func (h *RoomHub) JoinRoom(c *Client, roomID string) bool {
	if v, ok := h.rooms.Load(roomID); ok {
		r, _ := v.(*room)
		select {
		case r.online <- c:
			h.RemoveClient(c)
			return true
		case <-h.quit:
			return false
		}
	}
	return false
}

// LeaveRoom leave rooom
func (h *RoomHub) LeaveRoom(c *Client, roomID string) bool {
	if v, ok := h.rooms.Load(roomID); ok {
		r, _ := v.(*room)
		select {
		case r.offline <- c:
			return true
		case <-h.quit:
			return false
		}
	}
	return false
}

// IsClosed check if room closed
func (h *RoomHub) IsClosed() bool {
	select {
	case <-h.quit:
		return true
	default:
		return false
	}
}

// RoomList get room list
func (h *RoomHub) RoomList() []*RoomInfo {
	res := make([]*RoomInfo, 0)
	h.rooms.Range(func(key, value interface{}) bool {
		if r, ok := value.(*room); ok {
			res = append(res, &r.RoomInfo)
		}
		return true
	})
	return res
}

// Broadcast message to all online clients
func (h *RoomHub) Broadcast(msg *Message) {
	if v, ok := h.rooms.Load(msg.Room); ok {
		if room, ok := v.(*room); ok {
			room.Broadcast(msg)
		}
	}
}

// AddHandlers add on message handler
func (h *RoomHub) AddHandlers(handlers ...MessageHandler) {
	if len(handlers) == 0 {
		return
	}
	for _, handler := range handlers {
		if handler != nil {
			h.handlers = append(h.handlers, handler)
		}
	}
}

// OnMessage handler list
func (h *RoomHub) OnMessage(msg *Message) {
	defer func() {
		if err := recover(); err != nil {
			clog.Error(2, "Panic to handle message (%v), error %v", msg, err)
		}
	}()
	for _, handler := range h.handlers {
		if handler(msg, h); msg.Discard {
			break
		}
	}
}

// RemoveClient remove client from unjoin room list
func (h *RoomHub) RemoveClient(c *Client) {
	if _, ok := h.clients.Load(c.id); ok {
		h.clients.Delete(c.id)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}
		if u.Host != r.Host {
			clog.Warn("websocket origin not equal: %s - %s", u.Host, r.Host)
		}
		return true
	},
}

// ServeWebsocket websocket connect handler
func (h *RoomHub) ServeWebsocket(w http.ResponseWriter, r *http.Request) {
	//serveChatHandler(h, w, r)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		clog.Error(2, "websocket chat connection error: %v", err)
		return
	}

	c := NewClient(h, conn)
	h.clients.Store(c.id, c)

	clog.Trace("new websocket client %v", c.id)
}
