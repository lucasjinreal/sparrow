package main

import (
	"flag"
	"./chat"
	_ "net/http/pprof"

	"github.com/go-clog/clog"
	"github.com/gorilla/mux"
	"net/http"
	"gitlab.com/jinfagang/colorgo"
	"net"
	"fmt"
	"./common"
)

func GetLocalIPAddr()  {
	addr, err := net.InterfaceAddrs()
	common.CheckError(err)
	for _, add := range addr {
		if ipNet, ok := add.(*net.IPNet); ok && ! ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				fmt.Println(ipNet.IP.String())
			}
		}
	}
}

func init() {
	clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level:      clog.TRACE,
		BufferSize: 1024,
	})
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "http service address")
	flag.Parse()

	cg.PrintlnGreen("=> Starting sparrow, serves all the messages...")
	cg.PrintlnGreen("=> Now sparrow on serving.")
	// get local ip, this can be add as local server
	fmt.Println()
	cg.PrintlnBlue("=> For local group chat, using local address below (one of them):")
	GetLocalIPAddr()


	hub := chat.NewChatHub()
	hub.NewRoom("默认聊天组")

	hub.AddHandlers(
		func(msg *chat.Message, hub *chat.RoomHub) {
			if msg.Type == chat.T_MESSAGE {
				clog.Trace("Got a msg: %+v", msg)
			} else {
				clog.Info("cmd: %s, from: %s", msg.Type, msg.From)
			}
		},
	)

	r := mux.NewRouter()
	r.StrictSlash(false)
	r.HandleFunc("/", IndexHandle)
	r.HandleFunc("/ws", hub.ServeWebsocket)
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	//http.HandleFunc("/", serveHome)
	//http.HandleFunc("/ws", hub.ServeWebsocket)

	defer clog.Shutdown()
	clog.Info("%v", http.ListenAndServe(addr, r))
}
