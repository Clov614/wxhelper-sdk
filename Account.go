// Package wxhelper_sdk
// @Author Clover
// @Data 2025/1/7 下午3:59:00
// @Desc
package wxhelper_sdk

type Account struct {
	Account         string `json:"account"`
	City            string `json:"city"`
	Country         string `json:"country"`
	CurrentDataPath string `json:"currentDataPath"`
	DataSavePath    string `json:"dataSavePath"`
	DbKey           string `json:"dbKey"`
	HeadImage       string `json:"headImage"`
	Mobile          string `json:"mobile"`
	Name            string `json:"name"`
	Province        string `json:"province"`
	Signature       string `json:"signature"`
	Wxid            string `json:"wxid"`
	PrivateKey      string `json:"privateKey"`
	PublicKey       string `json:"publicKey"`
}
