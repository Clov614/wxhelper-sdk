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
	ErrEmpty      = errors.New("the message is empty")
	ErrIlegal     = errors.New("the message is illegal")
	ErrNoMessage  = errors.New("the message is empty")
	ErrBufferFull = errors.New("the message buffer is full")
)

type MessagePairs struct {
	Msgs []*Message
}

// IsLegal 判断消息对是否合法
func (mp *MessagePairs) IsLegal() (bool, error) {
	if len(mp.Msgs) == 0 {
		return false, ErrEmpty
	}
	if len(mp.Msgs) != 2 {
		return false, ErrIlegal
	}
	return true, nil
}

func NewMessagePairs(msgs ...*Message) *MessagePairs {
	msgList := make([]*Message, len(msgs))
	for i, msg := range msgs {
		msgList[i] = msg
	}
	return &MessagePairs{msgList}
}

type MessageBuffer struct {
	msgCH       chan *Message      // 原始消息输入通道
	pairsCH     chan *MessagePairs // 成对消息输出通道
	unpairedMsg *Message           // 缓存的未配对消息
	timeout     time.Duration      // 等待配对的最大时间
	timer       *time.Timer
}

// NewMessageBuffer 创建消息缓冲区
func NewMessageBuffer(bufferSize int, pairTimeOut time.Duration) *MessageBuffer {
	mb := &MessageBuffer{
		msgCH:   make(chan *Message, bufferSize),
		pairsCH: make(chan *MessagePairs, bufferSize/2), // 成对输出，容量为原始通道的一半
		timeout: pairTimeOut,
	}
	go mb.processMessages()
	return mb
}

// processMessages 内部处理消息，将其配对后发送到 pairsCH
func (mb *MessageBuffer) processMessages() {
	for {
		// 阻塞式读取消息
		msg, ok := <-mb.msgCH
		if !ok {
			// 通道关闭时退出循环
			return
		}
		if mb.unpairedMsg == nil {
			// 如果没有未配对的消息，存入 unpairedMsg，并启动定时器等待第二条消息
			mb.unpairedMsg = msg
			mb.timer = time.AfterFunc(mb.timeout, mb.handleTimeout) // 使用定时器
		} else {
			// 已有未配对的消息，与当前消息配对后发送
			pair := NewMessagePairs(mb.unpairedMsg, msg)
			mb.pairsCH <- pair
			mb.unpairedMsg = nil
			if mb.timer != nil {
				mb.timer.Stop() // 如果配对成功，取消定时器
			}
		}
	}
}

// handleTimeout 处理超时逻辑
func (mb *MessageBuffer) handleTimeout() {
	if mb.unpairedMsg != nil {
		logging.Warn("message timeout, discarding unpaired message")
		mb.unpairedMsg = nil // 丢弃单条消息
	}
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
func (mb *MessageBuffer) Get(ctx context.Context) (*MessagePairs, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case pair := <-mb.pairsCH:
		logging.Info("retrieved message pair from buffer")
		return pair, nil
	}
}
