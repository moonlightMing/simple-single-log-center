package wsserver

import (
    "github.com/gorilla/websocket"
    "net/http"
    "sync"
)

var (
    wsUpgrade = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        // 允许跨域
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
)

type wsMessage struct {
    MessageType int
    Data        []byte
}

type WsConnection struct {
    WsConn    *websocket.Conn // 底层WebSocket
    InChan    chan *wsMessage // 读队列
    OutChan   chan *wsMessage // 写队列
    mutex     sync.Mutex      // 避免重复关闭管道
    IsClosed  bool            // 是否处于关闭状态
    CloseChan chan struct{}   // 关闭通知
}

func NewWsConnection(w http.ResponseWriter, r *http.Request) (*WsConnection, error) {
    wsSocket, err := wsUpgrade.Upgrade(w, r, nil)
    if err != nil {
        return nil, err
    }

    return &WsConnection{
        WsConn:    wsSocket,
        InChan:    make(chan *wsMessage, 10),
        OutChan:   make(chan *wsMessage, 500),
        CloseChan: make(chan struct{}),
        IsClosed:  false,
    }, nil
}

func (wsConn *WsConnection) WsHandler() {
    // 读协程
    go wsConn.wsReadLoop()
    // 写协程
    go wsConn.wsWriteLoop()
}

func (wsConn *WsConnection) WsRead() (*wsMessage) {
    wsMsg := <-wsConn.InChan
    return wsMsg
}

func (wsConn *WsConnection) WsWrite(msgType int, data []byte) error {
    wsConn.OutChan <- &wsMessage{MessageType: msgType, Data: data}
    return nil
}

func (wsConn *WsConnection) wsReadLoop() {
    for !wsConn.IsClosed {
        // 读一个message
        msgType, data, err := wsConn.WsConn.ReadMessage()
        if err != nil {
            wsConn.CloseChan <- struct{}{}
            break
        }

        if msgType == websocket.CloseMessage {
            //log.Println("the ws close")
            wsConn.CloseChan <- struct{}{}
            break
        }

        req := &wsMessage{
            MessageType: msgType,
            Data:        data,
        }

        wsConn.InChan <- req
    }
}

func (wsConn *WsConnection) wsWriteLoop() {
    for !wsConn.IsClosed {
        msg, ok := <-wsConn.OutChan
        if !ok {
            break
        }
        if err := wsConn.WsConn.WriteMessage(msg.MessageType, msg.Data); err != nil {
            wsConn.CloseChan <- struct{}{}
        }
    }
}

func (wsConn *WsConnection) Close() {
    wsConn.mutex.Lock()
    defer wsConn.mutex.Unlock()
    if !wsConn.IsClosed {
        wsConn.IsClosed = true
        //close(wsConn.InChan)
        //close(wsConn.OutChan)
        //close(wsConn.CloseChan)
        wsConn.WsConn.Close()
    }
}
