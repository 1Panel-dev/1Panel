package v1

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/net"
)

func TestCopu(t *testing.T) {
	fmt.Println(net.IOCounters(false))
	fmt.Println(net.IOCounters(true))
}

// [{"name":"all","bytesSent":15498384367,"bytesRecv":18415197440,"packetsSent":11608058,"packetsRecv":12976591,"errin":0,"errout":21,"dropin":0,"dropout":1181,"fifoin":0,"fifoout":0}] <nil>
// [{"name":"lo0","bytesSent":12947010452,"bytesRecv":12947010452,"packetsSent":6519135,"packetsRecv":6519135,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"gif0","bytesSent":0,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"stf0","bytesSent":0,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"ap1","bytesSent":0,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"en0","bytesSent":2531548007,"bytesRecv":5443547837,"packetsSent":5019082,"packetsRecv":6374381,"errin":0,"errout":0,"dropin":0,"dropout":1065,"fifoin":0,"fifoout":0} {"name":"en1","bytesSent":0,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"en2","bytesSent":0,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"bridge0","bytesSent":0,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"awdl0","bytesSent":103717,"bytesRecv":157244,"packetsSent":412,"packetsRecv":460,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"llw0","bytesSent":0,"bytesRecv":0,"packetsSent":0,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"utun0","bytesSent":21810,"bytesRecv":0,"packetsSent":121,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"utun1","bytesSent":21810,"bytesRecv":0,"packetsSent":121,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"utun2","bytesSent":21810,"bytesRecv":0,"packetsSent":121,"packetsRecv":0,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0} {"name":"en5","bytesSent":6093969,"bytesRecv":6264468,"packetsSent":47985,"packetsRecv":48051,"errin":0,"errout":21,"dropin":0,"dropout":116,"fifoin":0,"fifoout":0} {"name":"utun3","bytesSent":13576933,"bytesRecv":18231580,"packetsSent":21129,"packetsRecv":34612,"errin":0,"errout":0,"dropin":0,"dropout":0,"fifoin":0,"fifoout":0}] <nil>
