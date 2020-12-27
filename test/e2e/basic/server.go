package basic

import (
	"fmt"

	"github.com/fatedier/frp/test/e2e/framework"
	"github.com/fatedier/frp/test/e2e/framework/consts"
	"github.com/fatedier/frp/test/e2e/pkg/port"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("[Feature: Server Manager]", func() {
	f := framework.NewDefaultFramework()

	It("Ports Whitelist", func() {
		serverConf := consts.DefaultServerConfig
		clientConf := consts.DefaultClientConfig

		serverConf += `
			allow_ports = 10000-20000,20002,30000-50000
		`

		tcpPortName := port.GenName("TCP", port.WithRangePorts(10000, 20000))
		udpPortName := port.GenName("UDP", port.WithRangePorts(30000, 50000))
		clientConf += fmt.Sprintf(`
			[tcp-allowded-in-range]
			type = tcp
			local_port = {{ .%s }}
			remote_port = {{ .%s }}
			`, framework.TCPEchoServerPort, tcpPortName)
		clientConf += fmt.Sprintf(`
			[tcp-port-not-allowed]
			type = tcp
			local_port = {{ .%s }}
			remote_port = 20001
			`, framework.TCPEchoServerPort)
		clientConf += fmt.Sprintf(`
			[tcp-port-unavailable]
			type = tcp
			local_port = {{ .%s }}
			remote_port = {{ .%s }}
			`, framework.TCPEchoServerPort, consts.PortServerName)
		clientConf += fmt.Sprintf(`
			[udp-allowed-in-range]
			type = udp
			local_port = {{ .%s }}
			remote_port = {{ .%s }}
			`, framework.UDPEchoServerPort, udpPortName)
		clientConf += fmt.Sprintf(`
			[udp-port-not-allowed]
			type = udp
			local_port = {{ .%s }}
			remote_port = 20003
			`, framework.UDPEchoServerPort)

		f.RunProcesses([]string{serverConf}, []string{clientConf})

		// TCP
		// Allowed in range
		framework.ExpectTCPRequest(f.UsedPorts[tcpPortName], []byte(consts.TestString), []byte(consts.TestString), connTimeout)
		// Not Allowed
		framework.ExpectTCPRequestError(20001, []byte(consts.TestString), connTimeout)
		// Unavailable, already bind by frps
		framework.ExpectTCPRequestError(f.UsedPorts[consts.PortServerName], []byte(consts.TestString), connTimeout)

		// UDP
		// Allowed in range
		framework.ExpectUDPRequest(f.UsedPorts[udpPortName], []byte(consts.TestString), []byte(consts.TestString), connTimeout)
		// Not Allowed
		framework.ExpectUDPRequestError(20003, []byte(consts.TestString), connTimeout)
	})
})
