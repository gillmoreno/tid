package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Super small signaling server for WebRTC bootstrap / NAT traversal.
// This is the ONLY centralized piece. It does NOT carry your data or code blobs.
// Peers exchange offers, answers, and ICE candidates here.
//
// Endpoints (all CORS open for easy testing):
//   POST /room/{room}/offer   { "sdp": "...", "type": "offer" }
//   GET  /room/{room}/answer
//   POST /room/{room}/answer  { "sdp": "...", "type": "answer" }
//   POST /room/{room}/candidate  { "candidate": "...", "sdpMid": "..." }
//   GET  /room/{room}/candidates
//
// For the thin test we also support simple address exchange as fallback:
//   POST /register   { "room": "...", "addr": "..." }
//   GET  /get?room=...

type signalMsg struct {
	SDP       string `json:"sdp,omitempty"`
	Type      string `json:"type,omitempty"`
	Candidate string `json:"candidate,omitempty"`
	SDPMid    string `json:"sdpMid,omitempty"`
}

type addrInfo struct {
	Addr string    `json:"addr"`
	Room string    `json:"room"`
	TS   time.Time `json:"ts"`
}

var (
	addrFlag = flag.String("addr", ":8081", "listen address for signaling")
	mu       sync.Mutex

	// WebRTC signaling per room
	offers     = make(map[string]signalMsg)
	answers    = make(map[string]signalMsg)
	candidates = make(map[string][]signalMsg)

	// Simple addr exchange (fallback / thin test)
	addrs = make(map[string]addrInfo)
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/register", registerAddr)
	mux.HandleFunc("/get", getAddrs)

	// WebRTC style
	mux.HandleFunc("/room/", roomHandler)

	// CORS for browser test client
	handler := cors(mux)

	fmt.Printf("Tiny signaling server on %s\n", *addrFlag)
	fmt.Println("For WebRTC test (recommended for phone):")
	fmt.Println("  Two browsers open test-client.html, point to this server, use same room.")
	fmt.Println("Expose with: ngrok http 8081   (then use the https ngrok url in the client)")
	log.Fatal(http.ListenAndServe(*addrFlag, handler))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the Origin if present (supports 'null' for file:// and exact localhost origin)
		// Fall back to * 
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, ngrok-skip-browser-warning")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	// Echo Origin to support 'null' (file open) and exact localhost origins
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, ngrok-skip-browser-warning")

	// Very naive path parsing: /room/{room}/offer etc.
	path := r.URL.Path
	parts := splitPath(path) // /room/ROOM/verb
	if len(parts) < 3 {
		http.Error(w, "bad path", 400)
		return
	}
	room := parts[1]
	verb := parts[2]

	// Log for debugging (only for interesting verbs)
	if verb == "answer" || verb == "candidates" || verb == "offer" {
		log.Printf("CORS set for /room/%s/%s , Origin: %s", room, verb, r.Header.Get("Origin"))
	}

	mu.Lock()
	defer mu.Unlock()

	switch {
	case r.Method == "POST" && verb == "offer":
		var m signalMsg
		json.NewDecoder(r.Body).Decode(&m)
		offers[room] = m
		w.WriteHeader(200)
		log.Printf("Room %s: offer received", room)

	case r.Method == "GET" && verb == "offer":
		if o, ok := offers[room]; ok {
			json.NewEncoder(w).Encode(o)
			// keep it for now so answerer can get it
		} else {
			w.WriteHeader(204)
		}

	case r.Method == "GET" && verb == "answer":
		if a, ok := answers[room]; ok {
			json.NewEncoder(w).Encode(a)
			delete(answers, room) // consume
		} else {
			w.WriteHeader(204)
		}

	case r.Method == "POST" && verb == "answer":
		var m signalMsg
		json.NewDecoder(r.Body).Decode(&m)
		answers[room] = m
		w.WriteHeader(200)
		log.Printf("Room %s: answer received", room)

	case r.Method == "POST" && verb == "candidate":
		var m signalMsg
		json.NewDecoder(r.Body).Decode(&m)
		candidates[room] = append(candidates[room], m)
		w.WriteHeader(200)

	case r.Method == "GET" && verb == "candidates":
		list := candidates[room]
		json.NewEncoder(w).Encode(list)

	default:
		http.Error(w, "not found", 404)
	}
}

func registerAddr(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, ngrok-skip-browser-warning")

	var info addrInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	info.TS = time.Now()
	key := info.Room + ":" + r.RemoteAddr

	mu.Lock()
	addrs[key] = info
	mu.Unlock()

	w.WriteHeader(200)
}

func getAddrs(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, ngrok-skip-browser-warning")

	room := r.URL.Query().Get("room")
	if room == "" {
		http.Error(w, "room required", 400)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var out []addrInfo
	for _, a := range addrs {
		if a.Room == room && time.Since(a.TS) < 3*time.Minute {
			out = append(out, a)
		}
	}
	json.NewEncoder(w).Encode(out)
}

func splitPath(p string) []string {
	var res []string
	start := 0
	for i := 0; i < len(p); i++ {
		if p[i] == '/' {
			if i > start {
				res = append(res, p[start:i])
			}
			start = i + 1
		}
	}
	if start < len(p) {
		res = append(res, p[start:])
	}
	return res
}