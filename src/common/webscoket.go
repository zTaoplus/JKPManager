package common

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient 结构体表示 WebSocket 客户端
type WebSocketClient struct {
	URL        string          // WebSocket 服务器的 URL
	Conn       *websocket.Conn // WebSocket 连接
	done       chan struct{}   // 通知连接已关闭的通道
	sendQueue  chan []byte     // 待发送消息的队列
	Result     []byte
	ResultChan chan []byte
}

// NewWebSocketClient 创建一个新的 WebSocket 客户端
func NewWebSocketClient(urlStr string) *WebSocketClient {
	return &WebSocketClient{
		URL:        urlStr,
		done:       make(chan struct{}),
		sendQueue:  make(chan []byte, 100), // 缓冲区大小可根据需要调整
		ResultChan: make(chan []byte, 100),
	}
}

func (c *WebSocketClient) Activate() error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 1 * time.Second // Timeout for handshake

	log.Printf("connected WebSocket 服务器: %s,HandshakeTimeout:%s\n", c.URL, dialer.HandshakeTimeout)
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.Conn = conn

	go c.receiveLoop()

	return nil
}

// Connect 连接到 WebSocket 服务器
func (c *WebSocketClient) Connect() error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 1 * time.Second // Timeout for handshake
	// dialer.TLSHandshakeTimeout = 5 * time.Second // Timeout for TLS handshake
	// dialer.WriteTimeout = 5 * time.Second        // Timeout for write operations
	// dialer.ReadTimeout = 5 * time.Second         // Timeout for read operations

	log.Printf("连接到 WebSocket 服务器: %s,HandshakeTimeout:%s\n", c.URL, dialer.HandshakeTimeout)
	//TODO(ZT): 在第一个连接建立的时候 是会收到3条消息的
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	// TODO: if err send delete to eg?
	c.Conn = conn

	go c.sendLoop()
	go c.receiveLoop()

	return nil
}

// Close 关闭 WebSocket 连接
func (c *WebSocketClient) Close() {
	if c.Conn == nil {
		return
	}
	close(c.done)
	// close(c.ResultChan)
	c.Conn.Close()

}

// 发送消息到 WebSocket 服务器
func (c *WebSocketClient) Send(message []byte) error {
	c.sendQueue <- message
	return nil
}

// sendLoop 循环发送消息到 WebSocket 服务器
func (c *WebSocketClient) sendLoop() {

	for {
		select {
		case <-c.done:
			return
		case message := <-c.sendQueue:
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				// log.Println("发送消息时发生错误:", err)
				return
			}
		}
	}
}

// receiveLoop 循环接收消息从 WebSocket 服务器
func (c *WebSocketClient) Receive() ([]byte, error) {
	_, message, err := c.Conn.ReadMessage()
	if err != nil {
		log.Println("接收消息时发生错误:", err)
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
				log.Printf("%v 连接关闭", c.URL)
				return

			}
			c.Result = message
			c.ResultChan <- message
		}
	}
}
