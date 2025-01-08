// Package wxhelper_sdk
// @Author Clover
// @Data 2025/1/7 下午3:59:00
// @Desc
package wxhelper_sdk

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wxhelper-sdk/logging"
)

type Message struct {
	Content            string `json:"content"`
	CreateTime         int    `json:"createTime"`
	DisplayFullContent string `json:"displayFullContent"`
	FromUser           string `json:"fromUser"`
	MsgId              int64  `json:"msgId"`
	MsgSequence        int    `json:"msgSequence"`
	Pid                int    `json:"pid"`
	Signature          string `json:"signature"`
	ToUser             string `json:"toUser"`
	Type               int    `json:"type"`
	Base64Img          string `json:"base64Img,omitempty"`

	account *Account
}

var (
	ErrBufferFull = errors.New("the message buffer is full")
)

type MessageBuffer struct {
	msgCH chan *Message // 原始消息输入通道
}

// NewMessageBuffer 创建消息缓冲区 <缓冲大小>
func NewMessageBuffer(bufferSize int) *MessageBuffer {
	mb := &MessageBuffer{
		msgCH: make(chan *Message, bufferSize),
	}
	return mb
}

// Put 向缓冲区中添加消息
func (mb *MessageBuffer) Put(ctx context.Context, msg *Message) error {
	retries := 3
	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case mb.msgCH <- msg:
			logging.Info("put message to buffer")
			return nil
		default:
			logging.Warn("message buffer is full, retrying", map[string]interface{}{fmt.Sprintf("%d", i+1): retries})
		}

		// Optional: add a small delay before retrying to prevent busy-waiting
		time.Sleep(time.Millisecond * 100)
	}
	return ErrBufferFull
}

// Get 获取一组成对的消息（阻塞等待）
func (mb *MessageBuffer) Get(ctx context.Context) (*Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case pair := <-mb.msgCH:
		logging.Info("retrieved message pair from buffer")
		return pair, nil
	}
}
