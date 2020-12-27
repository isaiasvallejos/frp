package basic

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatedier/frp/test/e2e/framework"
	"github.com/fatedier/frp/test/e2e/framework/consts"
	"github.com/fatedier/frp/test/e2e/pkg/port"

	. "github.com/onsi/ginkgo"
)

var connTimeout = 2 * time.Second

var _ = Describe("[Feature: Basic]", func() {
	f := framework.NewDefaultFramework()

	Describe("TCP && UDP", func() {
		types := []string{"tcp", "udp"}
		for _, t := range types {
			proxyType := t
			It(fmt.Sprintf("Expose a %s echo server", strings.ToUpper(proxyType)), func() {
				serverConf := consts.DefaultServerConfig
				clientConf := consts.DefaultClientConfig

				localPortName := ""
				protocol := "tcp"
				switch proxyType {
				case "tcp":
					localPortName = framework.TCPEchoServerPort
					protocol = "tcp"
				case "udp":
					localPortName = framework.UDPEchoServerPort
					protocol = "udp"
				}
				getProxyConf := func(proxyName string, portName string, extra string) string {
					return fmt.Sprintf(`
				[%s]
				type = %s
				local_port = {{ .%s }}
				remote_port = {{ .%s }}
				`+extra, proxyName, proxyType, localPortName, portName)
				}

				tests := []struct {
					proxyName   string
					portName    string
					extraConfig string
				}{
					{
						proxyName: "normal",
						portName:  port.GenName("Normal"),
					},
					{
						proxyName:   "with-encryption",
						portName:    port.GenName("WithEncryption"),
						extraConfig: "use_encryption = true",
					},
					{
						proxyName:   "with-compression",
						portName:    port.GenName("WithCompression"),
						extraConfig: "use_compression = true",
					},
					{
						proxyName: "with-encryption-and-compression",
						portName:  port.GenName("WithEncryptionAndCompression"),
						extraConfig: `
						use_encryption = true
						use_compression = true
						`,
					},
				}

				// build all client config
				for _, test := range tests {
					clientConf += getProxyConf(test.proxyName, test.portName, test.extraConfig) + "\n"
				}
				// run frps and frpc
				f.RunProcesses([]string{serverConf}, []string{clientConf})

				for _, test := range tests {
					framework.ExpectRequest(protocol, f.UsedPorts[test.portName],
						[]byte(consts.TestString), []byte(consts.TestString), connTimeout, test.proxyName)
				}
			})
		}
	})

	Describe("STCP && SUDP", func() {
		types := []string{"stcp", "sudp"}
		for _, t := range types {
			proxyType := t
			It(fmt.Sprintf("Expose echo server with %s", strings.ToUpper(proxyType)), func() {
				serverConf := consts.DefaultServerConfig
				clientServerConf := consts.DefaultClientConfig
				clientVisitorConf := consts.DefaultClientConfig

				localPortName := ""
				protocol := "tcp"
				switch proxyType {
				case "stcp":
					localPortName = framework.TCPEchoServerPort
					protocol = "tcp"
				case "sudp":
					localPortName = framework.UDPEchoServerPort
					protocol = "udp"
				}

				correctSK := "abc"
				wrongSK := "123"

				getProxyServerConf := func(proxyName string, extra string) string {
					return fmt.Sprintf(`
				[%s]
				type = %s
				role = server
				sk = %s
				local_port = {{ .%s }}
				`+extra, proxyName, proxyType, correctSK, localPortName)
				}
				getProxyVisitorConf := func(proxyName string, portName, visitorSK, extra string) string {
					return fmt.Sprintf(`
				[%s]
				type = %s
				role = visitor
				server_name = %s
				sk = %s
				bind_port = {{ .%s }}
				`+extra, proxyName, proxyType, proxyName, visitorSK, portName)
				}

				tests := []struct {
					proxyName    string
					bindPortName string
					visitorSK    string
					extraConfig  string
					expectError  bool
				}{
					{
						proxyName:    "normal",
						bindPortName: port.GenName("Normal"),
						visitorSK:    correctSK,
					},
					{
						proxyName:    "with-encryption",
						bindPortName: port.GenName("WithEncryption"),
						visitorSK:    correctSK,
						extraConfig:  "use_encryption = true",
					},
					{
						proxyName:    "with-compression",
						bindPortName: port.GenName("WithCompression"),
						visitorSK:    correctSK,
						extraConfig:  "use_compression = true",
					},
					{
						proxyName:    "with-encryption-and-compression",
						bindPortName: port.GenName("WithEncryptionAndCompression"),
						visitorSK:    correctSK,
						extraConfig: `
						use_encryption = true
						use_compression = true
						`,
					},
					{
						proxyName:    "with-error-sk",
						bindPortName: port.GenName("WithErrorSK"),
						visitorSK:    wrongSK,
						expectError:  true,
					},
				}

				// build all client config
				for _, test := range tests {
					clientServerConf += getProxyServerConf(test.proxyName, test.extraConfig) + "\n"
				}
				for _, test := range tests {
					clientVisitorConf += getProxyVisitorConf(test.proxyName, test.bindPortName, test.visitorSK, test.extraConfig) + "\n"
				}
				// run frps and frpc
				f.RunProcesses([]string{serverConf}, []string{clientServerConf, clientVisitorConf})

				for _, test := range tests {
					expectResp := []byte(consts.TestString)
					if test.expectError {
						framework.ExpectRequestError(protocol, f.UsedPorts[test.bindPortName],
							[]byte(consts.TestString), connTimeout, test.proxyName)
						continue
					}

					framework.ExpectRequest(protocol, f.UsedPorts[test.bindPortName],
						[]byte(consts.TestString), expectResp, connTimeout, test.proxyName)
				}
			})
		}
	})

	Describe("TCPMUX", func() {
		It("Type tcpmux", func() {
			serverConf := consts.DefaultServerConfig
			clientConf := consts.DefaultClientConfig

			tcpmuxHTTPConnectPortName := port.GenName("TCPMUX")
			serverConf += fmt.Sprintf(`
			tcpmux_httpconnect_port = {{ .%s }}
			`, tcpmuxHTTPConnectPortName)

			getProxyConf := func(proxyName string, extra string) string {
				return fmt.Sprintf(`
				[%s]
				type = tcpmux
				multiplexer = httpconnect
				local_port = {{ .%s }}
				custom_domains = %s
				`+extra, proxyName, framework.TCPEchoServerPort, proxyName)
			}

			tests := []struct {
				proxyName   string
				portName    string
				extraConfig string
			}{
				{
					proxyName: "normal",
				},
				{
					proxyName:   "with-encryption",
					extraConfig: "use_encryption = true",
				},
				{
					proxyName:   "with-compression",
					extraConfig: "use_compression = true",
				},
				{
					proxyName: "with-encryption-and-compression",
					extraConfig: `
						use_encryption = true
						use_compression = true
					`,
				},
			}

			// build all client config
			for _, test := range tests {
				clientConf += getProxyConf(test.proxyName, test.extraConfig) + "\n"
			}
			// run frps and frpc
			f.RunProcesses([]string{serverConf}, []string{clientConf})

			// Request without HTTP connect should get error
			framework.ExpectTCPRequestError(f.UsedPorts[tcpmuxHTTPConnectPortName],
				[]byte(consts.TestString), connTimeout, "request without HTTP connect")
			for _, test := range tests {
				// TODO
				framework.ExpectTCPRequestError(f.UsedPorts[tcpmuxHTTPConnectPortName],
					[]byte(consts.TestString), connTimeout, test.proxyName)
			}
		})
	})
})
