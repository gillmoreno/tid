package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	// We inline the framing to exactly match the rooms relay protocol.
	// In real code we will import or copy from the rooms/relay/protocol.go
)

const (
	msgUpdate     byte = 0x00 // append to room log
	msgCheckpoint byte = 0x01 // replace log with this single frame
	msgSyncEnd    byte = 0xfe // relay → client: backlog replay complete
)

// frame wraps a payload with the 1-byte type (same as rooms relay)
func frame(typ byte, payload []byte) []byte {
	out := make([]byte, 1+len(payload))
	out[0] = typ
	copy(out[1:], payload)
	return out
}

func readFrame(r io.Reader) (byte, []byte, error) {
	// Read 1-byte type
	typBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, typBuf); err != nil {
		return 0, nil, err
	}
	typ := typBuf[0]

	// For this thin test we use a simple 4-byte length prefix after the type.
	// (Real rooms relay is length-agnostic on wire for the blob, but for test we add length.)
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return 0, nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}
	return typ, payload, nil
}

func writeFrame(w io.Writer, typ byte, payload []byte) error {
	if err := binary.Write(w, binary.BigEndian, typ); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint32(len(payload))); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

var (
	listenAddr  = flag.String("listen", ":4501", "TCP address to listen on for direct P2P")
	peerAddr    = flag.String("peer", "", "direct peer address for testing (skips signaling)")
	signaling   = flag.String("signaling", "", "optional signaling server WS URL for real NAT test")
	roomID      = flag.String("room", "test-room", "room id")
	sendTest    = flag.Bool("send", false, "send a test update after connecting")

	activeConn net.Conn
	connMu     sync.Mutex
)

func main() {
	flag.Parse()

	fmt.Printf("=== Local P2P Relay (thin test layer) ===\n")
	fmt.Printf("room=%s\n", *roomID)
	if *signaling != "" {
		fmt.Printf("signaling=%s (will use for bootstrap)\n", *signaling)
	} else if *peerAddr != "" {
		fmt.Printf("direct peer=%s\n", *peerAddr)
	} else {
		fmt.Printf("listening for direct connections\n")
	}

	if *listenAddr != "" {
		go func() {
			if err := startListener(*listenAddr); err != nil {
				log.Printf("listener error: %v", err)
			}
		}()
	}

	if *peerAddr != "" {
		go func() {
			if err := connectDirect(*peerAddr); err != nil {
				log.Printf("direct connect error: %v", err)
			}
		}()
	}

	if *signaling != "" {
		go func() {
			if err := connectViaSignaling(*signaling); err != nil {
				log.Printf("signaling error: %v", err)
			}
		}()
	}

	if *sendTest {
		go func() {
			time.Sleep(2 * time.Second)
			// In real version this would come from the meta-app / custom code
			testPayload := []byte("test-expense:coffee:4.50")
			// For now just log — we will wire send later
			fmt.Printf("[TEST] would send update: %s\n", testPayload)
		}()
	}

	fmt.Println("\nCommands:")
	fmt.Println("  type 'send' + Enter to send a real update over the P2P connection")
	fmt.Println("  Ctrl-C or Enter to quit")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "send" {
			connMu.Lock()
			c := activeConn
			connMu.Unlock()
			if c != nil {
				payload := []byte("test-expense:coffee:4.50:" + time.Now().Format(time.RFC3339))
				if err := writeFrame(c, msgUpdate, payload); err != nil {
					fmt.Printf("[local] send error: %v\n", err)
				} else {
					fmt.Printf("[local] sent update: %s\n", payload)
				}
			} else {
				fmt.Println("[local] no active peer connection yet. Wait for 'connected' message.")
			}
		}
	}
}

func startListener(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("listening on %s for direct P2P", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleConnection(conn, "inbound-direct")
	}
}

func connectDirect(addr string) error {
	log.Printf("trying direct connect to %s", addr)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	handleConnection(conn, "outbound-direct")
	return nil
}

func connectViaSignaling(sigURL string) error {
	sigURL = strings.TrimSuffix(sigURL, "/")
	log.Printf("Using signaling at %s for discovery...", sigURL)

	// Register our "address" (for thin test we just use a fake browser-style id or our listen port)
	myID := "go-" + *roomID + "-" + fmt.Sprintf("%d", time.Now().Unix())
	register := map[string]string{
		"room": *roomID,
		"addr": myID,
	}
	b, _ := json.Marshal(register)

	// Register
	resp, err := http.Post(sigURL+"/register", "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("register to signaling: %w", err)
	}
	resp.Body.Close()
	log.Printf("Registered with signaling as %s", myID)

	// Poll for other peers
	for i := 0; i < 15; i++ { // try for ~30s
		time.Sleep(2 * time.Second)
		r, err := http.Get(sigURL + "/get?room=" + url.QueryEscape(*roomID))
		if err != nil {
			continue
		}
		var list []struct {
			Addr string `json:"addr"`
			Room string `json:"room"`
		}
		json.NewDecoder(r.Body).Decode(&list)
		r.Body.Close()

		for _, p := range list {
			if p.Addr != myID {
				log.Printf("Found peer via signaling: %s – attempting direct connect...", p.Addr)
				// For thin test we just log. In a real version we would try to connect using the addr
				// or switch to WebRTC using additional signaling endpoints.
				// For now the browser test-client does full WebRTC.
				log.Printf("Peer info from signaling: %+v (use the browser client for full WebRTC test)", p)
			}
		}
	}
	return nil
}

func handleConnection(conn net.Conn, direction string) {
	defer conn.Close()

	connMu.Lock()
	activeConn = conn
	connMu.Unlock()

	defer func() {
		connMu.Lock()
		if activeConn == conn {
			activeConn = nil
		}
		connMu.Unlock()
	}()

	log.Printf("[%s] connected: %s", direction, conn.RemoteAddr())
	fmt.Printf("\n>>> Connected to peer! Type 'send' + Enter here to transmit a real framed update.\n")

	// Demo loop: read frames, log them, optionally reply
	for {
		typ, payload, err := readFrame(conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] readFrame error: %v", direction, err)
			}
			return
		}
		log.Printf("[%s] got frame type=0x%x payload=%q", direction, typ, payload)

		if typ == msgUpdate {
			fmt.Printf("\n>>> RECEIVED UPDATE from peer: %s\n", payload)
			// reply with a checkpoint (demo of sync)
			reply := []byte("ack:" + string(payload))
			if err := writeFrame(conn, msgCheckpoint, reply); err != nil {
				log.Printf("write error: %v", err)
				return
			}
		} else if typ == msgCheckpoint {
			fmt.Printf("\n>>> RECEIVED CHECKPOINT from peer: %s\n", payload)
		}
	}
}