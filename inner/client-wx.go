// Package inner
// @Author Clover
// @Data 2025/1/8 下午2:52:00
// @Desc 对请求返回的原始消息进一步处理 wrapper
package inner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wxhelper-sdk/inner/models"
)

type WxClient struct {
	transport *Transport
}

var (
	ErrWxClientResp = errors.New("wx-client's response err")
	ErrJSONDecoder  = errors.New("JSON decoder error")
	colonSeparator  = ":"
)

func NewWxClient(WxApiBaseUrl string, tcpHookURL string) *WxClient {
	// api请求封装
	return &WxClient{transport: NewTransport(WxApiBaseUrl, tcpHookURL)}
}

func (c *WxClient) CheckLogin(ctx context.Context) (bool, error) {
	resp, err := c.transport.CheckLogin(ctx)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrWxClientResp, err)
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return false, fmt.Errorf("%w: %w", ErrJSONDecoder, err)
	}
	return r.Code == 1, nil
}

func (c *WxClient) GetUserInfo(ctx context.Context) (*models.Account, error) {
	resp, err := c.transport.GetUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWxClientResp, err)
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[*models.Account]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrJSONDecoder, err)
	}
	if r.Code != 1 {
		return nil, errors.New("get user info failed")
	}
	return r.Data, nil
}

func (c *WxClient) handleResponse(resp *http.Response, expectedCode int) error {
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("json decoder error: %w", err)
	}
	if r.Code != expectedCode {
		return fmt.Errorf("unexpected response code: %d, msg: %s", r.Code, r.Msg)
	}
	return nil
}

func (c *WxClient) SendText(ctx context.Context, to string, content string) error {
	resp, err := c.transport.SendText(ctx, to, content)
	if err != nil {
		return fmt.Errorf("send text failed: %w", err)
	}
	return c.handleResponse(resp, 0)
}

func (c *WxClient) GetContactList(ctx context.Context) (models.Members, error) {
	resp, err := c.transport.GetContactList(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[models.Members]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (c *WxClient) HTTPHookSyncMsg(ctx context.Context, url *url.URL, timeout time.Duration) error {
	opt := TransportHookSyncMsgOption{
		Url:        url.String(),
		EnableHttp: true,
		Timeout:    strconv.Itoa(int(timeout / time.Second)),
		Ip:         url.Hostname(),
		Port:       url.Port(),
	}
	resp, err := c.transport.HookSyncMsg(ctx, opt)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {

		return err
	}
	if r.Code != 0 {
		return errors.New("hook sync msg failed")
	}
	return nil
}

func (c *WxClient) HookSyncMsg(ctx context.Context) error {
	hookUrl := strings.Split(c.transport.TcpHookURL, colonSeparator)
	ip := hookUrl[0]
	port, err := strconv.Atoi(hookUrl[1])
	if err != nil {
		return fmt.Errorf("hook_port ilegal: %w", err)
	}
	opt := TransportHookSyncMsgOption{
		EnableHttp: false,
		Ip:         ip,
		Port:       strconv.Itoa(port),
	}
	_, err = c.transport.UnhookSyncMsg(ctx) // 每次上报hook先unhook
	if err != nil {
		return err
	}
	resp, err := c.transport.HookSyncMsg(ctx, opt)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code != 0 {
		return errors.New("hook sync msg failed")
	}
	return nil
}

func (c *WxClient) UnhookSyncMsg(ctx context.Context) error {
	resp, err := c.transport.UnhookSyncMsg(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code != 0 {
		return errors.New("unhook sync msg failed")
	}
	return nil
}

func (c *WxClient) SendImage(ctx context.Context, to string, imgPath string) error {
	resp, err := c.transport.SendImage(ctx, to, imgPath)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code != 1 {
		return errors.New(r.Msg)
	}
	return nil
}

func (c *WxClient) SendFile(ctx context.Context, to string, filePath string) error {
	resp, err := c.transport.SendFile(ctx, to, filePath)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code == 0 {
		return fmt.Errorf("send file failed with code %d", r.Code)
	}
	return nil
}

func (c *WxClient) GetChatRoomDetail(ctx context.Context, chatRoomId string) (*models.ChatRoomInfo, error) {
	resp, err := c.transport.GetChatRoomDetail(ctx, chatRoomId)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[models.ChatRoomInfo]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.Code != 1 {
		return nil, errors.New("get chat room detail failed")
	}
	return &r.Data, nil
}

func (c *WxClient) GetMemberFromChatRoom(ctx context.Context, chatRoomId string) (*models.GroupMember, error) {
	resp, err := c.transport.GetMemberFromChatRoom(ctx, chatRoomId)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[models.GroupMember]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.Code != 1 {
		return nil, errors.New("get chat room member failed")
	}
	return &r.Data, nil
}

func (c *WxClient) GetContactProfile(ctx context.Context, wxid string) (*models.Profile, error) {
	resp, err := c.transport.GetContactProfile(ctx, wxid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[models.Profile]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.Code < 0 {
		return nil, errors.New("get contact profile failed")
	}
	return &r.Data, nil
}

type SendAtTextOption struct {
	WxIds      []string
	ChatRoomID string
	Content    string
}

func (c *WxClient) SendAtText(ctx context.Context, opt SendAtTextOption) error {
	resp, err := c.transport.SendAtText(ctx, sendAtTextOption{
		WxIds:      strings.Join(opt.WxIds, ","),
		ChatRoomId: opt.ChatRoomID,
		Msg:        opt.Content,
	})
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code < 0 {
		return errors.New("send at text failed")
	}
	return nil
}

func (c *WxClient) AddMemberIntoChatRoom(ctx context.Context, chatRoomID string, memberIDs []string) error {
	resp, err := c.transport.AddMemberIntoChatRoom(ctx, chatRoomID, strings.Join(memberIDs, ","))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code != 1 {
		return errors.New("add member into chat room failed")
	}
	return nil
}

func (c *WxClient) InviteMemberToChatRoom(ctx context.Context, chatRoomID string, memberIDs []string) error {
	resp, err := c.transport.InviteMemberToChatRoom(ctx, chatRoomID, strings.Join(memberIDs, ","))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code != 1 {
		return errors.New("invite member into chat room failed")
	}
	return nil
}

func (c *WxClient) ForwardMsg(ctx context.Context, msgID, wxID string) error {
	resp, err := c.transport.ForwardMsg(ctx, msgID, wxID)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code != 1 {
		return errors.New("forward msg failed")
	}
	return nil
}

func (c *WxClient) QuitChatRoom(ctx context.Context, chatRoomId string) error {
	resp, err := c.transport.QuitChatRoom(ctx, chatRoomId)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var r result[any]
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Code <= 0 {
		return errors.New("quit chatroom failed")
	}
	return nil
}
