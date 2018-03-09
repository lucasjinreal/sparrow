package chat

import (
	"../std"
	"sync/atomic"
	"time"

	"github.com/go-clog/clog"
)

const (
	ch32 = 32
)

// RoomInfo basic room information
//
type RoomInfo struct {
	ID      string    `json:"id,omitempty"`         // Room ID
	Name    string    `json:"name,omitempty"`       // Room name
	Desc    string    `json:"text,omitempty"`       // Room description
	Avatar  string    `json:"avatar,omitempty"`     // Room avatar
	Active  bool      `json:"active,omitempty"`     // Room is still active or not
	CCount  int32     `json:"clintCount,omitempty"` // Online client count
	MCount  int32     `json:"msgCount,omitempty"`   // Room history message count
	Updated time.Time `json:"updated,omitempty"`    // Latest message timestamp
}

type room struct {
	RoomInfo
	hub       *RoomHub
	online    chan *Client       // clients to be online
	offline   chan *Client       // clients to be offline
	clients   map[uint64]*Client // all online clients
	broadcast std.Queue          // message to broadcast
	quit      chan struct{}
}

func (r *room) run() {
	defer func() {
		r.broadcast.Close()
		clog.Info("Room %s closed.", r.RoomInfo.ID)
	}()

	var (
		msg   *Message
		chmsg = r.broadcast.Cout()
	)
	for {
		select {
		case c := <-r.online:
			r.clients[c.id] = c
			r.CCount++
			break

		case c := <-r.offline:
			if _, ok := r.clients[c.id]; ok {
				delete(r.clients, c.id)
				c.msgs.Close()
				r.CCount--
			}
			break

		case itm, ok := <-chmsg:
			if !ok {
				return
			}
			if msg, ok = itm.(*Message); !ok {
				break
			}
			for _, c := range r.clients {
				c.PushMessage(msg)
			}
			break

		case <-r.quit:
			msg = &Message{
				Timestamp: std.GetNowMs(),
				Room:      r.RoomInfo.ID,
				Type:      T_CLOSE,
			}
			for _, c := range r.clients {
				if !c.PushMessage(msg) {
					delete(r.clients, c.id)
					r.CCount--
				}
			}
			return

		case <-r.hub.quit:
			return
		}
	}
}

func newRoom(name string, h *RoomHub) *room {
	r := &room{
		RoomInfo: RoomInfo{
			ID:     std.GenUIDs(),
			Name:   name,
			Desc:   name,
			Active: true,
		},
		hub:       h,
		quit:      make(chan struct{}, 1),
		online:    make(chan *Client, ch32),
		offline:   make(chan *Client, ch32),
		clients:   make(map[uint64]*Client),
		broadcast: std.NewSyncQueue(maxQueueSize),
	}
	go r.run()
	return r
}

func (r *room) Broadcast(msg *Message) {
	if msg != nil {
		r.Updated = time.Now()
		r.broadcast.Add(msg)
		atomic.AddInt32(&r.MCount, 1)
	}
}

func (r *room) Close() {
	close(r.quit)
	r.Active = false
}
