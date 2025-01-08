package wxhelper_sdk

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMessageBuffer_Put(t *testing.T) {
	// 测试缓冲区大小为 2
	mb := NewMessageBuffer(2)

	// 定义一个消息
	msg := &Message{
		Content:    "Test message",
		CreateTime: int(time.Now().Unix()),
		FromUser:   "user1",
		ToUser:     "user2",
	}

	// 设置一个上下文（不取消）
	ctx := context.Background()

	// 测试将消息放入缓冲区
	err := mb.Put(ctx, msg)
	assert.Nil(t, err, "expected no error when putting a message into buffer")

	// 测试缓冲区已满，尝试添加第三条消息
	msg2 := &Message{
		Content:    "Second test message",
		CreateTime: int(time.Now().Unix()),
		FromUser:   "user3",
		ToUser:     "user4",
	}

	err = mb.Put(ctx, msg2)
	assert.Nil(t, err, "expected no error when putting another message into buffer")

	// 测试当缓冲区满时，第三条消息应被拒绝
	msg3 := &Message{
		Content:    "Third test message",
		CreateTime: int(time.Now().Unix()),
		FromUser:   "user5",
		ToUser:     "user6",
	}

	err = mb.Put(ctx, msg3)
	assert.Equal(t, ErrBufferFull, err, "expected ErrBufferFull when buffer is full")
}

func TestMessageBuffer_Get(t *testing.T) {
	// 测试缓冲区大小为 1
	mb := NewMessageBuffer(1)

	// 定义一个消息
	msg := &Message{
		Content:    "Test message",
		CreateTime: int(time.Now().Unix()),
		FromUser:   "user1",
		ToUser:     "user2",
	}

	// 设置一个上下文（不取消）
	ctx := context.Background()

	// 先将消息放入缓冲区
	err := mb.Put(ctx, msg)
	assert.Nil(t, err, "expected no error when putting a message into buffer")

	// 测试从缓冲区获取消息
	retrievedMsg, err := mb.Get(ctx)
	assert.Nil(t, err, "expected no error when getting a message from buffer")
	assert.Equal(t, msg, retrievedMsg, "retrieved message should be the same as the one put into the buffer")
}

func TestMessageBuffer_ContextCancel(t *testing.T) {
	// 测试缓冲区大小为 1
	mb := NewMessageBuffer(1)

	// 定义一个消息
	msg := &Message{
		Content:    "Test message",
		CreateTime: int(time.Now().Unix()),
		FromUser:   "user1",
		ToUser:     "user2",
	}

	// 设置一个带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 将消息放入缓冲区
	err := mb.Put(ctx, msg)
	assert.Nil(t, err, "expected no error when putting a message into buffer")

	// 取消上下文
	cancel()

	// 测试当上下文被取消时，调用 Get 会返回 ctx.Err()
	_, err = mb.Get(ctx)
	assert.Equal(t, context.Canceled, err, "expected context.Canceled error when context is cancelled")
}

func TestMessageBuffer_BufferFullRetry(t *testing.T) {
	// 测试缓冲区大小为 1
	mb := NewMessageBuffer(1)

	// 定义一个消息
	msg := &Message{
		Content:    "Test message",
		CreateTime: int(time.Now().Unix()),
		FromUser:   "user1",
		ToUser:     "user2",
	}

	// 设置一个上下文（不取消）
	ctx := context.Background()

	// 第一次将消息放入缓冲区
	err := mb.Put(ctx, msg)
	assert.Nil(t, err, "expected no error when putting a message into buffer")

	// 测试缓冲区满时，第二次调用应尝试重试
	// 由于缓冲区大小为 1，所以此时第二条消息会被阻塞并进行重试
	msg2 := &Message{
		Content:    "Second test message",
		CreateTime: int(time.Now().Unix()),
		FromUser:   "user3",
		ToUser:     "user4",
	}

	// 这里需要稍微等待一下以确保重试机制被触发
	err = mb.Put(ctx, msg2)
	assert.Equal(t, err, ErrBufferFull, "expected ErrBufferFull when buffer is full")
}
