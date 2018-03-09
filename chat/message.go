package chat

const (
	T_JOIN    = "JOIN"    // c -> s, client join room
	T_LEAVE   = "LEAVE"   // c -> s, client leave room
	T_CREATE  = "CREATE"  // c -> s, client room
	T_CLOSE   = "CLOSE"   // s -> c, room closed
	T_ROOMS   = "ROOMS"   // c -> s, get room list
	T_MESSAGE = "MESSAGE" // c <-> s, messge
)

// Message receive/send to websocket client
//
type Message struct {
	ID        uint64 `json:"id,omitempty"`        // message id
	Type      string `json:"type,omitempty"`      // message type
	From      string `json:"from,omitempty"`      // message from client id
	Room      string `json:"room,omitempty"`      // which room this message sends to
	Timestamp int64  `json:"timestamp,omitempty"` // message timestamp
	Data      string `json:"data,omitempty"`      // message data
	Discard   bool   `json:"-"`                   // discard this message, set by handler
}

// The MessageHandler type is an adapter to allow the use of
// ordinary functions as OnMessage handlers.
// return false, if don't want furture process
type MessageHandler func(*Message, *RoomHub)

// OnMessage handler
func (fn MessageHandler) OnMessage(msg *Message, hub *RoomHub) {
	fn(msg, hub)
}
