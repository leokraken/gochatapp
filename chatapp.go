package main

import (
	"encoding/json"
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func main() {
	lastMessages := []string{}
	var lmMutex sync.Mutex
	// Sets the number of maxium goroutines to the 2*numberCPU + 1
	runtime.GOMAXPROCS((runtime.NumCPU() * 2) + 1)

	// Configuring socket.io Server
	sio := socketio.NewSocketIOServer(&socketio.Config{})

	sio.Of("/chat").On("connect", func(ns *socketio.NameSpace) {
		log.Println("connected:", ns.Id(), " in channel ", ns.Endpoint())
		for i, _ := range lastMessages {
			ns.Emit("message", lastMessages[i])
		}
	})

	sio.Of("/chat").On("joined_message", func(ns *socketio.NameSpace, message string) {
		ns.Session.Values["username"] = message
		res := map[string]interface{}{
			"username": message,
			"dateTime": time.Now().UTC().Format(time.RFC3339Nano),
			"type":     "joined_message",
		}
		jsonRes, _ := json.Marshal(res)
		sio.In("/chat").Broadcast("message", string(jsonRes))
	})

	sio.Of("/chat").On("send_message", func(ns *socketio.NameSpace, message string) {
		res := map[string]interface{}{
			"username": ns.Session.Values["username"],
			"message":  message,
			"dateTime": time.Now().UTC().Format(time.RFC3339),
			"type":     "message",
		}
		jsonRes, _ := json.Marshal(res)
		lmMutex.Lock()
		if len(lastMessages) == 100 {
			lastMessages = lastMessages[1:100]
		}
		lastMessages = append(lastMessages, string(jsonRes))
		lmMutex.Unlock()
		sio.In("/chat").Broadcast("message", string(jsonRes))
	})

	// Sets up the handlers and listen on port 8080
	sio.Handle("/", http.FileServer(http.Dir("./templates/")))
	sio.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	log.Println("listening on port 8080")
	http.ListenAndServe(":8080", sio)
}
