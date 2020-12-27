package consts

import (
	"fmt"

	"github.com/fatedier/frp/test/e2e/pkg/port"
)

const (
	TestString = "frp is a fast reverse proxy to help you expose a local server behind a NAT or firewall to the internet."
)

var (
	PortServerName  string
	PortClientAdmin string

	DefaultServerConfig = `
	[common]
	bind_port = {{ .%s }}
	log_level = trace
	`

	DefaultClientConfig = `
	[common]
	server_port = {{ .%s }}
	log_level = trace
	`
)

func init() {
	PortServerName = port.GenName("Server")
	PortClientAdmin = port.GenName("ClientAdmin")
	DefaultServerConfig = fmt.Sprintf(DefaultServerConfig, port.GenName("Server"))
	DefaultClientConfig = fmt.Sprintf(DefaultClientConfig, port.GenName("Server"))
}
