package websocket

import (
    "log"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

const wsTimeout = 30 * time.Second

type WebSocketProxy struct {
    conn         *websocket.Conn
    pingTicker   *time.Ticker
    heartbeat    *time.Timer
    heartbeatMux sync.Mutex
}

func (wsp *WebSocketProxy) CheckHeartbeat() {
    wsp.heartbeatMux.Lock()
    defer wsp.heartbeatMux.Unlock()

    if wsp.heartbeat != nil {
        wsp.heartbeat.Stop()
    }

    wsp.heartbeat = time.AfterFunc(wsTimeout, func() {
        log.Println("Terminating WebSocket due to inactivity")
        wsp.pingTicker.Stop()
        wsp.conn.Close()
        wsp.conn = nil
    })
}

func (wsp *WebSocketProxy) Connect(token, deviceId string) (*websocket.Conn, error) {
    if wsp.conn != nil {
        return wsp.conn, nil
    }

    dialer := websocket.DefaultDialer
    conn, _, err := dialer.Dial("wss://api.orbitbhyve.com/v1/events", nil)
    if err != nil {
        return nil, err
    }
    wsp.conn = conn

    wsp.pingTicker = time.NewTicker(25 * time.Second)
    go func() {
        for range wsp.pingTicker.C {
            wsp.conn.WriteJSON(map[string]string{"event": "ping"})
        }
    }()

    wsp.conn.SetCloseHandler(func(code int, text string) error {
        log.Println("WebSocket Closed")
        wsp.pingTicker.Stop()
        return nil
    })

    wsp.conn.SetPingHandler(func(appData string) error {
        log.Println("Received ping from server")
        return nil
    })

    wsp.conn.SetPongHandler(func(appData string) error {
        log.Println("Received pong from server")
        wsp.CheckHeartbeat()
        return nil
    })

    wsp.CheckHeartbeat()
    wsp.conn.WriteJSON(map[string]string{
        "event":                "app_connection",
        "orbit_session_token":  token,
        "subscribe_device_id":  deviceId,
    })

    log.Println("WebSocket Connected")
    return wsp.conn, nil
}
