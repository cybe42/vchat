package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"

	"golang.org/x/net/websocket"
)

func padOrTrim(bb []byte, size int) []byte {
	l := len(bb)
	if l == size {
		return bb
	}
	if l > size {
		return bb[l-size:]
	}
	tmp := make([]byte, size)
	copy(tmp[size-l:], bb)
	return tmp
}

func encrypt(key, text []byte) ([]byte, error) {
	c, err := aes.NewCipher(padOrTrim(key, 32))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, text, nil), nil
}

func decrypt(key, text []byte) ([]byte, error) {
	c, err := aes.NewCipher(padOrTrim(key, 32))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(text) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, text := text[:nonceSize], text[nonceSize:]
	return gcm.Open(nil, nonce, text, nil)
}

type Msg struct {
	Name    string
	Message string
}

type Client struct {
	Name    string
	IP      string
	ws      *websocket.Conn
	wg      *sync.WaitGroup
	channel string
}

func (c *Client) SetChannel(channel string) {
	c.channel = channel
}

func (c *Client) GetChannel() string {
	return c.channel
}

func (c *Client) Connect() error {
	var wg sync.WaitGroup
	wg.Add(1)
	c.wg = &wg
	if c.channel == "" {
		c.channel = "chat"
	}

	u, err := url.Parse(c.IP)
	if err != nil {
		return err
	}
	origin, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}
	origin = "http://" + origin
	ws, err := websocket.Dial(c.IP, "", origin)
	if err != nil {
		return err
	}
	c.ws = ws
	if _, err := ws.Write([]byte(c.Name)); err != nil {
		return err
	}
	return nil
}

func (c *Client) Listen(callback func(Msg, error)) error {
	go func() {
		defer c.wg.Done()
		for {
			var msg = make([]byte, 1024) // should be 1024
			n, err := c.ws.Read(msg)
			if err != nil {
				callback(Msg{}, err)
				return
			}
			var rawstr string = string(msg[:n])
			var message Msg
			json.Unmarshal([]byte(rawstr), &message)
			b64decoded, err := b64.StdEncoding.DecodeString(message.Message)
			if err != nil {
				continue
			} else {
				decrypted, err := decrypt([]byte(c.channel), []byte(b64decoded))
				if err != nil {
					continue
				} else {
					callback(Msg{Name: message.Name, Message: string(decrypted)}, nil)
				}
			}
		}
	}()
	c.wg.Wait()
	return nil
}

func (c *Client) Send(msg string, channel string) error {
	encrypted, err := encrypt([]byte(channel), []byte(msg))
	b64encrypted := b64.StdEncoding.EncodeToString(encrypted)
	if err != nil {
		return err
	}
	if _, err := c.ws.Write([]byte(b64encrypted)); err != nil {
		return err
	}
	return nil
}
