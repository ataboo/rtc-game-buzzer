package room

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MaxMessageSize = 2048
	ReadWait       = 3 * time.Second
	WriteWait      = 3 * time.Second
	PongWait       = 10 * time.Second
	PingPeriod     = 5 * time.Second
)

type WSMessage struct {
	Content string
	Sender  *Player
}

type Game struct {
	Host      *Player
	Players   []*Player
	Locked    bool
	msgChan   chan WSMessage
	leaveChan chan *Player
}

type Player struct {
	IsHost    bool
	Name      string
	Conn      *websocket.Conn
	WriteChan chan WSMessage
}

func (p *Player) readPump(leaveChan chan<- *Player, reqChan chan<- WSMessage) {
	defer func() {
		leaveChan <- p
		p.Conn.Close()
	}()

	p.Conn.SetReadLimit(MaxMessageSize)
	p.Conn.SetReadDeadline(time.Now().Add(PongWait))
	p.Conn.SetPongHandler(func(string) error { p.Conn.SetReadDeadline(time.Now().Add(PongWait)); return nil })
	for {
		req := WSMessage{}
		err := p.Conn.ReadJSON(req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Unexpected close")
			}
			fmt.Printf("Client %s Read err: %s", p.Name, err.Error())
			return
		}
		req.Sender = p

		reqChan <- req
	}
}

func (p *Player) writePump(leaveChan chan<- *Player) {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		leaveChan <- p
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			p.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := p.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case res, ok := <-p.WriteChan:
			if !ok {
				p.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			p.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := p.Conn.WriteJSON(res); err != nil {
				return
			}
		}
	}
}

func (g *Game) Start() {

}

func (g *Game) Stop() {

}

func (g *Game) AddPlayer(player *Player) error {
	for _, p := range g.Players {
		if p.Name == player.Name {
			return fmt.Errorf("name duplicate")
		}
	}

	if len(g.Players) == 0 {
		g.Host = player
	}
	g.Players = append(g.Players, player)

	go player.readPump(g.leaveChan, g.msgChan)
	go player.writePump(g.leaveChan)

	return nil
}

// func NewWSClient(conn *websocket.Conn, id int, name string) *WSClient {
// 	return &WSClient{
// 		conn:      conn,
// 		writeChan: make(chan *msg.WSResponse),
// 		ClientID:  id,
// 		Name:      name,
// 	}
// }

// func (c *WSClient) Start(leaveChan chan<- *WSClient, reqChan chan<- *msg.WSRequest) {
// 	go c.readPump(leaveChan, reqChan)
// 	go c.writePump(leaveChan)
// }

// func (c *WSClient) writePump(leaveChan chan<- *WSClient) {
// 	ticker := time.NewTicker(PingPeriod)
// 	defer func() {
// 		ticker.Stop()
// 		c.conn.Close()
// 	}()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
// 			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
// 				return
// 			}
// 		case res, ok := <-c.writeChan:
// 			if !ok {
// 				c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
// 				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
// 				return
// 			}

// 			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
// 			if err := c.conn.WriteJSON(res); err != nil {
// 				return
// 			}
// 		}
// 	}
// }

// func (c *WSClient) WriteResponse(res *msg.WSResponse) bool {
// 	select {
// 	case c.writeChan <- res:
// 		return true
// 	default:
// 		close(c.writeChan)
// 		return false
// 	}
// }

// func WriteResponse(conn *websocket.Conn, res msg.WSResponse) error {
// 	conn.SetWriteDeadline(time.Now().Add(WriteWait))
// 	return conn.WriteJSON(res)
// }
