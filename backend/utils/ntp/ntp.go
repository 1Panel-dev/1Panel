package ntp

import (
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

const ntpEpochOffset = 2208988800

type packet struct {
	Settings       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	RefTimeSec     uint32
	RefTimeFrac    uint32
	OrigTimeSec    uint32
	OrigTimeFrac   uint32
	RxTimeSec      uint32
	RxTimeFrac     uint32
	TxTimeSec      uint32
	TxTimeFrac     uint32
}

func GetRemoteTime() (time.Time, error) {
	conn, err := net.Dial("udp", "pool.ntp.org:123")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return time.Time{}, fmt.Errorf("failed to set deadline: %v", err)
	}

	req := &packet{Settings: 0x1B}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		return time.Time{}, fmt.Errorf("failed to set request: %v", err)
	}

	rsp := &packet{}
	if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
		return time.Time{}, fmt.Errorf("failed to read server response: %v", err)
	}

	secs := float64(rsp.TxTimeSec) - ntpEpochOffset
	nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32

	showtime := time.Unix(int64(secs), nanos)

	return showtime, nil
}

func UpdateSystemDate(dateTime string) error {
	system := runtime.GOOS
	if system == "linux" {
		if _, err := cmd.Execf(`date -s  "` + dateTime + `"`); err != nil {
			return fmt.Errorf("update system date failed, err: %v", err)
		}
		return nil
	}
	return fmt.Errorf("the current system architecture %v does not support synchronization", system)
}
