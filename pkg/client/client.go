package client

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/gillcaleb/orbit-bhyve-go-client/pkg/websocket"
)

type Client struct {
    token      string
    userId     string
    config     Config
    ws         *WebSocketProxy
    client     *http.Client
}

type Config struct {
	Endpoint string
    Email    string
    Password string
    DeviceId string
}

func NewClient(config Config) *Client {
    return &Client{
        config: config,
        ws:     &WebSocketProxy{},
        client: &http.Client{},
    }
}

func (c *Client) init() error {
    log.Println("Initializing Client")
    payload := map[string]interface{}{
        "session": map[string]string{
            "email":    c.config.Email,
            "password": c.config.Password,
        },
    }
    payloadBytes, _ := json.Marshal(payload)
    req, err := http.NewRequest(http.MethodPost, c.config.Endpoint+"/session", bytes.NewBuffer(payloadBytes))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    c.token = result["orbit_session_token"].(string)
    c.userId = result["user_id"].(string)
    log.Println("API Token:", c.token)
    log.Println("User ID:", c.userId)
    return nil
}

func (c *Client) sync() error {
    conn, err := c.ws.Connect(c.token, c.config.DeviceId)
    if err != nil {
        return err
    }
    return conn.WriteJSON(map[string]interface{}{
        "event":     "sync",
        "device_id": c.config.DeviceId,
    })
}

func (c *Client) startZone(zoneId, minutes int) error {
    conn, err := c.ws.Connect(c.token, c.config.DeviceId)
    if err != nil {
        return err
    }
    return conn.WriteJSON(map[string]interface{}{
        "event":     "change_mode",
        "mode":      "manual",
        "device_id": c.config.DeviceId,
        "timestamp": time.Now().Format(time.RFC3339),
        "stations": []map[string]interface{}{
            {"station": zoneId, "run_time": minutes},
        },
    })
}

func (c *Client) stopZone() error {
    conn, err := c.ws.Connect(c.token, c.config.DeviceId)
    if err != nil {
        return err
    }
    return conn.WriteJSON(map[string]interface{}{
        "event":     "change_mode",
        "mode":      "manual",
        "device_id": c.config.DeviceId,
        "timestamp": time.Now().Format(time.RFC3339),
        "stations":  []map[string]interface{}{},
    })
}

func (c *Client) modeOff() error {
    conn, err := c.ws.Connect(c.token, c.config.DeviceId)
    if err != nil {
        return err
    }
    return conn.WriteJSON(map[string]interface{}{
        "event":     "change_mode",
        "mode":      "off",
        "device_id": c.config.DeviceId,
        "timestamp": time.Now().Format(time.RFC3339),
    })
}
