package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
	"github.com/googollee/go-socket.io"
	"database/sql"
	_ "github.com/lib/pq"
)

// Should look like path
const websocketRoom = "/chat"

func main() {

     	pg_user := os.Getenv("PG_USER")
	pg_db := os.Getenv("PG_DB")
	pg_host := os.Getenv("PG_HOST")
	pg_password := os.Getenv("PG_PASSWORD")
	//pg_port := os.Getenv("PG_PORT")


	db, err := sql.Open("postgres", "user=" +pg_user+ " host=" +pg_host+ " dbname=" + pg_db + " password=" + pg_password)
	if err != nil {
		log.Fatal(err)
	}

	lastMessages := []string{}
	var lmMutex sync.Mutex
	// Sets the number of maxium goroutines to the 2*numberCPU + 1
	runtime.GOMAXPROCS((runtime.NumCPU() * 2) + 1)

	// Configuring socket.io Server
	sio, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	sio.On("connection", func(so socketio.Socket) {
		var username string
		username = "User-" + so.Id()
		log.Println("on connection", username)
		so.Join(websocketRoom)

		lmMutex.Lock()
		for i, _ := range lastMessages {
			so.Emit("message", lastMessages[i])
		}
		lmMutex.Unlock()

		so.On("joined_message", func(message string) {
			username = message
			log.Println("joined_message", message)
			res := map[string]interface{}{
				"username": username,
				"dateTime": time.Now().UTC().Format(time.RFC3339Nano),
				"type":     "joined_message",
			}
			jsonRes, _ := json.Marshal(res)
			so.Emit("message", string(jsonRes))
			so.BroadcastTo(websocketRoom, "message", string(jsonRes))
		})
		so.On("send_message", func(message string) {
			log.Println("send_message from", username)
			//to db
			_, err := db.Exec("INSERT INTO messages(usuario, message) VALUES($1, $2)", username, message)
			if err != nil {
				log.Fatal(err)
			}

			res := map[string]interface{}{
				"username": username,
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
			so.Emit("message", string(jsonRes))
			so.BroadcastTo(websocketRoom, "message", string(jsonRes))
		})
		so.On("disconnection", func() {
			log.Println("on disconnect", username)
		})
	})
	sio.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	// Sets up the handlers and listen on port 8080
	http.Handle("/socket.io/", sio)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", http.FileServer(http.Dir("./templates/")))

	// Default to :8080 if not defined via environmental variable.
	var listen string = os.Getenv("LISTEN")

	if listen == "" {
		listen = ":8080"
	}

	log.Println("listening on", listen)
	http.ListenAndServe(listen, nil)
}
