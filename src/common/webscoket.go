package common

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	URL        string          // WebSocket server's URL
	Conn       *websocket.Conn // WebSocket Connection
	done       chan struct{}   // Flag to indicate that the connection is closed
	sendQueue  chan []byte     // Queue for sending messages
	Result     []byte          // Result of the last message
	ResultChan chan []byte     // Channel for receiving messages
}

func NewWebSocketClient(urlStr string) *WebSocketClient {
	return &WebSocketClient{
		URL:        urlStr,
		done:       make(chan struct{}),
		sendQueue:  make(chan []byte, 100),
		ResultChan: make(chan []byte, 100),
	}
}

func (c *WebSocketClient) Activate() error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 2 * time.Second // Timeout for handshake

	log.Printf("connecting WebSocket server: %s, HandshakeTimeout:%s\n", c.URL, dialer.HandshakeTimeout)
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.Conn = conn

	go c.receiveLoop()

	return nil
}

// create a new connection
func (c *WebSocketClient) Connect() error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 2 * time.Second // Timeout for handshake

	log.Printf("connected WebSocket server:: %s, HandshakeTimeout:%s\n", c.URL, dialer.HandshakeTimeout)

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.Conn = conn

	go c.sendLoop()
	go c.receiveLoop()

	return nil
}

// close the connection
func (c *WebSocketClient) Close() {
	if c.Conn == nil {
		return
	}
	close(c.done)
	c.Conn.Close()

}

// send message to ws server
func (c *WebSocketClient) Send(message []byte) error {
	c.sendQueue <- message
	return nil
}

func (c *WebSocketClient) sendLoop() {

	for {
		select {
		case <-c.done:
			return
		case message := <-c.sendQueue:
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("WriteMessage error:", err)
				return
			}
		}
	}
}

// receiveLoop 循环接收消息从 WebSocket 服务器
func (c *WebSocketClient) Receive() ([]byte, error) {
	_, message, err := c.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return message, nil
}
func (c *WebSocketClient) receiveLoop() {

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.Conn.ReadMessage()

			if err != nil {
				log.Printf("%v Connection Closed", c.URL)
				return

			}
			c.Result = message
			c.ResultChan <- message
		}
	}
}
