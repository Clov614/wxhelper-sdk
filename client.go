// Package wxhelper_sdk
// @Author Clover
// @Data 2025/1/7 下午4:01:00
// @Desc 客户端&消息hook
package wxhelper_sdk

import (
	"context"
	"fmt"
	"github.com/eatmoreapple/env"
	"time"
	"wxhelper-sdk/logging"
)

const (
	TCP_ADDR = "TCP_ADDR"
)

type Client struct {
	listener  *TCPMessageListener
	ctx       context.Context
	stop      context.CancelFunc
	msgBuffer *MessageBuffer
}

// GetMsgPair 获取消息对
func (c *Client) GetMsgPair() (*MessagePairs, error) {
	msgPair, err := c.msgBuffer.Get(c.ctx)
	if err != nil {
		return nil, err
	}
	return msgPair, nil
}

func (c *Client) startListen() error {
	var handler MessageHandlerFunc = func(message *Message) error {
		err := c.msgBuffer.Put(c.ctx, message)
		if err != nil {
			return fmt.Errorf("MessageHandler err: %w", err)
		}
		return nil
	}
	err := c.listener.ListenAndServe(c.ctx, handler)
	if err != nil {
		return fmt.Errorf("listener err: %w", err)
	}
	return nil
}

func NewClient(msgChanSize int) *Client {
	addr := env.Name(TCP_ADDR).StringOrElse("19099")
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		listener:  NewTCPMessageListener(addr),
		ctx:       ctx,
		stop:      cancel,
		msgBuffer: NewMessageBuffer(msgChanSize, time.Millisecond*100),
	}
}

// Run 运行
func (c *Client) Run() {
	err := c.startListen()
	// TODO test 发起 unHooksyncMsg && hooksyncMsg
	// TODO Get需要hook成功后才能使用 hookstate状态判断
	if err != nil {
		logging.Error(err.Error())
	}
}
