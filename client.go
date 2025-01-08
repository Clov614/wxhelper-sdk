// Package wxhelper_sdk
// @Author Clover
// @Data 2025/1/7 下午4:01:00
// @Desc 客户端&消息hook
package wxhelper_sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/eatmoreapple/env"
	"time"
	"wxhelper-sdk/inner"
	"wxhelper-sdk/logging"
)

const (
	ENVTcpAddr          = "TCP_ADDR"
	ENVWxApiBaseUrl     = "WX_API_BASE_URL"
	ENVTcpHookURL       = "WX_HOOK_URL"
	DefaultTcpAddr      = "19099"
	DefaultWxApiBaseUrl = "http://127.0.0.1:19088"
	DefaultTcpHookURL   = "127.0.0.1:19089"
)

var (
	ErrNotLogin = errors.New("not login")
)

type Client struct {
	listener  *TCPMessageListener
	ctx       context.Context
	stop      context.CancelFunc
	msgBuffer *MessageBuffer
	wxClient  *inner.WxClient
	isLogin   bool
}

// GetMsgPair 获取消息对
func (c *Client) GetMsgPair() (*MessagePairs, error) {
	if !c.isLogin {
		logging.Warn("客户端并未登录成功，请稍重试")
		return nil, ErrNotLogin
	}
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
	addr := env.Name(ENVTcpAddr).StringOrElse(DefaultTcpAddr)                   // "19099"
	WxApiBaseUrl := env.Name(ENVWxApiBaseUrl).StringOrElse(DefaultWxApiBaseUrl) // "http:// 127.0.0.1:19088"
	tcpHookURL := env.Name(ENVTcpHookURL).StringOrElse(DefaultTcpHookURL)       //  "127.0.0.1:19089"
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		listener:  NewTCPMessageListener(addr), // tcp server
		ctx:       ctx,
		stop:      cancel,
		msgBuffer: NewMessageBuffer(msgChanSize, time.Millisecond*100), // 消息缓冲区 <缓冲大小>, <获取消息对超时>
		wxClient:  inner.NewWxClient(WxApiBaseUrl, tcpHookURL),
	}
}

// Run 运行tcp监听 以及 请求tcp监听信息
func (c *Client) Run() {
	err := c.startListen()
	if err != nil {
		panic(err)
	}

	err = c.wxClient.HookSyncMsg(c.ctx)
	if err != nil {
		logging.Error(err.Error())
	}

	c.isLogin, err = c.wxClient.CheckLogin(c.ctx)
	if err != nil {
		logging.ErrorWithErr(err, "checkLogin error")
		c.stop()
		return
	}
}
