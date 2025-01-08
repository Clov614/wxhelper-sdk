// Package wxhelper_sdk
// @Author Clover
// @Data 2025/1/7 下午7:52:00
// @Desc
package wxhelper_sdk_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wxhelper-sdk"
)

func TestMessageBuffer_NormalPairing(t *testing.T) {
	bufferSize := 10
	timeout := time.Millisecond * 500
	mb := wxhelper_sdk.NewMessageBuffer(bufferSize, timeout)

	// 模拟两个正常的消息
	msg1 := &wxhelper_sdk.Message{Content: "Message 1", MsgId: 1}
	msg2 := &wxhelper_sdk.Message{Content: "Message 2", MsgId: 2}

	ctx := context.Background()

	// 添加两条消息
	if err := mb.Put(ctx, msg1); err != nil {
		t.Fatalf("failed to put message 1: %v", err)
	}
	if err := mb.Put(ctx, msg2); err != nil {
		t.Fatalf("failed to put message 2: %v", err)
	}

	// 获取成对消息
	pair, err := mb.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get message pair: %v", err)
	}

	if len(pair.Msgs) != 2 {
		t.Fatalf("expected 2 messages in pair, got %d", len(pair.Msgs))
	}
	if pair.Msgs[0].MsgId != msg1.MsgId || pair.Msgs[1].MsgId != msg2.MsgId {
		t.Fatalf("message pair content mismatch")
	}
}

func TestMessageBuffer_TimeoutDiscard(t *testing.T) {
	bufferSize := 10
	timeout := time.Millisecond * 100
	mb := wxhelper_sdk.NewMessageBuffer(bufferSize, timeout)

	// 模拟一条未配对的消息
	msg := &wxhelper_sdk.Message{Content: "Unpaired Message", MsgId: 1}

	ctx := context.Background()

	// 添加消息
	if err := mb.Put(ctx, msg); err != nil {
		t.Fatalf("failed to put message: %v", err)
	}

	// 等待超过超时时间
	time.Sleep(timeout + time.Millisecond*50)

	// 尝试获取配对消息
	timer := time.NewTimer(time.Millisecond * 10)
	go func() {
		<-timer.C
		msg1 := &wxhelper_sdk.Message{Content: "Unpaired Message", MsgId: 1}
		msg2 := &wxhelper_sdk.Message{Content: "Unpaired Message", MsgId: 1}
		_ = mb.Put(ctx, msg1)
		_ = mb.Put(ctx, msg2)
	}()
	get, err := mb.Get(ctx)
	if (*get).Msgs[0] != msg || err != nil {
		t.Log("unpaired message correctly discarded after timeout") // 应该没有配对消息

	} else {
		t.Fatalf("unpaired message incorrectly discarded after timeout")
	}
	//if b, _ := get.IsLegal(); b {
	//	t.Log("get:", *get)
	//}
}

func TestMessageBuffer_BufferFull(t *testing.T) {
	var n int = 20 // 测试次数
	for _ = range n {
		testMessageBufferFull(t)
	}
}

func testMessageBufferFull(t *testing.T) {
	bufferSize := 1
	timeout := time.Second * 20
	mb := wxhelper_sdk.NewMessageBuffer(bufferSize, timeout)

	ctx := context.Background()

	// 填满缓冲区
	for i := 0; i < 3*bufferSize; i++ {
		msg := &wxhelper_sdk.Message{Content: "Message", MsgId: int64(i)}
		if err := mb.Put(ctx, msg); err != nil {
			t.Fatalf("failed to put message: %v", err)
		}
	}

	// 再添加一条消息，应触发错误
	msg := &wxhelper_sdk.Message{Content: "Extra Message", MsgId: 999}
	err := mb.Put(ctx, msg)
	if err == nil {
		t.Fatalf("expected error when putting message to full buffer, got nil")
	}
}

func TestMessageBuffer_GetWithContextTimeout(t *testing.T) {
	bufferSize := 10
	timeout := time.Millisecond * 500
	mb := wxhelper_sdk.NewMessageBuffer(bufferSize, timeout)

	// 设置一个超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// 获取消息对，应该因超时返回错误
	_, err := mb.Get(ctx)
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
}

func TestMessagePairs_IsLegal(t *testing.T) {
	// 测试合法配对
	msg1 := &wxhelper_sdk.Message{Content: "Message 1"}
	msg2 := &wxhelper_sdk.Message{Content: "Message 2"}
	pair := wxhelper_sdk.NewMessagePairs(msg1, msg2)

	if ok, err := pair.IsLegal(); !ok || err != nil {
		t.Fatalf("expected legal pair, got error: %v", err)
	}

	// 测试单条消息（非法）
	pair = wxhelper_sdk.NewMessagePairs(msg1)
	if ok, err := pair.IsLegal(); ok || !errors.Is(err, wxhelper_sdk.ErrIlegal) {
		t.Fatalf("expected illegal pair, got: %v", err)
	}

	// 测试空消息
	pair = wxhelper_sdk.NewMessagePairs()
	if ok, err := pair.IsLegal(); ok || !errors.Is(err, wxhelper_sdk.ErrEmpty) {
		t.Fatalf("expected empty error, got: %v", err)
	}
}
