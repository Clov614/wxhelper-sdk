// Package wxhelper_sdk
// @Author Clover
// @Data 2025/1/7 下午10:29:00
// @Desc
package wxhelper_sdk

import (
	"fmt"
	"os"
	"testing"
)

func TestClient_GetMsgPair(t *testing.T) {
	os.Setenv(ENVTcpAddr, "19089")
	os.Setenv(ENVWxApiBaseUrl, "http://127.0.0.1:19088")
	os.Setenv(ENVTcpHookURL, "192.168.31.149:19089")
	defer func() {
		os.Unsetenv(ENVTcpAddr)
		os.Unsetenv(ENVWxApiBaseUrl)
		os.Unsetenv(ENVTcpHookURL)
	}()
	client := NewClient(100)
	client.Run(true)

	msg, err := client.GetMsg()
	if err != nil {
		t.Error(err)
	}
	t.Log(fmt.Sprintf("msg: %v", msg))

}
