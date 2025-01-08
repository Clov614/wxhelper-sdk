// Package wxhelper_sdk
// @Author Clover
// @Data 2025/1/7 下午4:01:00
// @Desc
package wxhelper_sdk

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"wxhelper-sdk/logging"
)

// MessageHandler 消息处理者
type MessageHandler interface {
	HandleMessage(message *Message) error
}

// MessageHandlerFunc 以函数的方式实现MessageHandler
type MessageHandlerFunc func(message *Message) error

func (f MessageHandlerFunc) HandleMessage(message *Message) error {
	return f(message)
}

// ReaderMessageHandler 从reader中解析 Message 并处理
type ReaderMessageHandler struct {
	Reader         io.Reader
	MessageHandler MessageHandler
}

func (rmh *ReaderMessageHandler) Serve() error {
	var msg Message
	if err := json.NewDecoder(rmh.Reader).Decode(&msg); err != nil {
		return err
	}
	logging.Info("parse message successfully")
	logging.Debug("[MessageHandler]", map[string]interface{}{"msg": msg})
	err := rmh.MessageHandler.HandleMessage(&msg)
	if err != nil {
		return err
	}
	return nil
}

// MessageListener 消息监听者
type MessageListener interface {
	ListenAndServe(handler MessageHandler) error
}

// TCPMessageListener tcp实现
type TCPMessageListener struct {
	Addr string
}

// ListenAndServe 启动tcp服务并监听处理消息
func (tl *TCPMessageListener) ListenAndServe(ctx context.Context, messageHandler MessageHandler) error {
	listener, err := net.Listen("tcp", tl.Addr)
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go tl.processMessage(conn, messageHandler) // 处理每个接收到的消息
	}
}

// 处理每个tcp连接消息
func (tl *TCPMessageListener) processMessage(conn net.Conn, messageHandler MessageHandler) {
	defer func() { _ = conn.Close() }()
	defer func() { _, _ = conn.Write([]byte("200 OK")) }()
	handler := ReaderMessageHandler{Reader: conn, MessageHandler: messageHandler}
	_ = handler.Serve()
}

func NewTCPMessageListener(addr string) *TCPMessageListener {
	return &TCPMessageListener{Addr: ":" + addr}
}
