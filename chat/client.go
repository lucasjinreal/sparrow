package chat

import (
	"encoding/json"
	"time"

	"../std"

	"github.com/go-clog/clog"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Maximum send queue size
	maxQueueSize = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client wrap websocket client
//
type Client struct {
	id  uint64 // client id
	ids string // client id string
	//room string          // room id
	hub  *RoomHub        // room hub
	msgs std.Queue       // message queue
	conn *websocket.Conn // websocket connection
	quit chan struct{}
}

// NewClient create an new client instance
func NewClient(hub *RoomHub, conn *websocket.Conn) *Client {
	c := &Client{
		id:   std.GenUniqueID(),
		hub:  hub,
		conn: conn,
		msgs: std.NewSyncQueue(maxQueueSize),
		quit: make(chan struct{}, 2),
	}
	//c.ids = fmt.Sprintf("%d", c.id)

	// handler websocket read/write client in seperate goroutine
	go c.readPump()
	go c.writePump()
	return c
}

// PushMessage push message to client
func (c *Client) PushMessage(msg *Message) bool {
	return c.msgs.Add(msg)
}

// OnMessage handle client message
func (c *Client) OnMessage(msg *Message) {
	switch msg.Type {
	case T_MESSAGE:
		clog.Trace("client %v send message: %v.", msg.From, msg.Data)
		if msg.Room == "" {
			msg.Data = "no room specified"
			c.ids = msg.From
			c.PushMessage(msg)
		} else {
			c.hub.Broadcast(msg)
		}

	case T_ROOMS:
		reply := &Message{Type: T_ROOMS}
		clog.Trace("client %s get room list", msg.From)
		if res := c.hub.RoomList(); res != nil {
			if bs, err := json.Marshal(res); err == nil {
				reply.Data = string(bs)
				reply.From = msg.From
			} else {
				clog.Error(2, "marshal rooms (%v) failed: %v.", res, err)
			}
		}
		c.PushMessage(reply)

	case T_JOIN:
		reply := &Message{Type: T_JOIN}
		clog.Trace("client %s join room %s", msg.From, msg.Room)
		if c.hub.JoinRoom(c, msg.Room) {
			reply.Data = msg.Room
		}
		c.PushMessage(reply)

	case T_LEAVE:
		reply := &Message{Type: T_LEAVE}
		clog.Trace("client %s leave room %s", msg.From, msg.Room)
		if c.hub.LeaveRoom(c, msg.Room) {
			reply.Data = msg.Room
		}
		c.PushMessage(reply)

	case T_CREATE:
		reply := &Message{Type: T_CREATE}
		clog.Trace("client %s create room %s", msg.From, msg.Data)
		if rm := c.hub.NewRoom(msg.Data); rm != nil {
			if bs, err := json.Marshal(rm); err == nil {
				reply.Data = string(bs)
			} else {
				clog.Error(2, "marshal room %+v failed: %v.", rm, err)
			}
		}
		c.PushMessage(reply)

	default:
		clog.Trace("unknown message type %v from client %s.", msg.Type, c.ids)
	}
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		close(c.quit)
		clog.Info("client %d read routine end.", c.id)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				clog.Error(2, "client %v read message error: %v", c.id, err)
			} else {
				clog.Trace("client %v closed.", c.id)
			}
			return
		}

		c.ids = msg.From
		msg.Timestamp = std.GetNowMs()
		if c.hub.OnMessage(&msg); msg.Discard {
			clog.Info("discard cleint %s message %v.", c.id, msg)
			continue
		}

		c.OnMessage(&msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.msgs.Close()
		c.conn.Close()
		clog.Info("client %d write routine end.", c.id)
	}()

	var (
		err   error
		msg   *Message
		chMsg = c.msgs.Cout()
	)

	for {
		select {
		case v, ok := <-chMsg:
			if !ok && c.hub.IsClosed() {
				clog.Warn("websocket server closed.")
				return
			}
			if msg, ok = v.(*Message); !ok {
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = c.conn.WriteJSON(msg)

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					clog.Error(2, "client %d writer error: %v", c.id, err)
				} else {
					clog.Trace("client %v closed.", c.id)
				}
				return
			}
			if msg.Type == T_CLOSE {
				return
			}

			clog.Trace("send message %+v to client %v.", msg.Data, c.id)
			break

		case <-ticker.C:
			// heartbeat with client
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = c.conn.WriteMessage(websocket.PingMessage, nil)

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					clog.Error(2, "client %d heartbeat error: %v", c.id, err)
				} else {
					clog.Trace("client %v heartbeat closed.", c.id)
				}
				return
			}

		case <-c.quit:
			return
		}
	}
}
