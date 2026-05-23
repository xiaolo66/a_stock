package a_stock

import (
	"github.com/bensema/gotdx"
)

// Client 封装 gotdx 主站行情连接。
type Client struct {
	tdx *gotdx.Client
}

// New 使用默认主站地址池创建客户端。
func New() *Client {
	hosts := gotdx.MainHostAddresses()
	primary := ""
	var pool []string
	if len(hosts) > 0 {
		primary = hosts[0]
		if len(hosts) > 1 {
			pool = hosts[1:]
		}
	}
	return NewWithOptions(
		gotdx.WithTCPAddress(primary),
		gotdx.WithTCPAddressPool(pool...),
		gotdx.WithAutoSelectFastest(true),
		gotdx.WithTimeoutSec(6),
	)
}

// NewWithOptions 使用自定义 gotdx 选项创建客户端。
func NewWithOptions(opts ...gotdx.Option) *Client {
	return &Client{tdx: gotdx.New(opts...)}
}

// Disconnect 断开行情连接。
func (c *Client) Disconnect() {
	if c == nil || c.tdx == nil {
		return
	}
	c.tdx.Disconnect()
}

// Raw 返回底层 gotdx 客户端，便于调用库内其他接口。
func (c *Client) Raw() *gotdx.Client {
	return c.tdx
}
