package qqwry

import (
	"encoding/binary"
	"net"
	"strings"

	"github.com/1Panel-dev/1Panel/cmd/server/qqwry"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	indexLen      = 7
	redirectMode1 = 0x01
	redirectMode2 = 0x02
)

var IpCommonDictionary []byte

type QQwry struct {
	Data   []byte
	Offset int64
}

func NewQQwry() (*QQwry, error) {
	IpCommonDictionary := qqwry.QQwryByte
	return &QQwry{Data: IpCommonDictionary}, nil
}

// readData 从文件中读取数据
func (q *QQwry) readData(num int, offset ...int64) (rs []byte) {
	if len(offset) > 0 {
		q.setOffset(offset[0])
	}
	nums := int64(num)
	end := q.Offset + nums
	dataNum := int64(len(q.Data))
	if q.Offset > dataNum {
		return nil
	}

	if end > dataNum {
		end = dataNum
	}
	rs = q.Data[q.Offset:end]
	q.Offset = end
	return
}

// setOffset 设置偏移量
func (q *QQwry) setOffset(offset int64) {
	q.Offset = offset
}

// Find ip地址查询对应归属地信息
func (q *QQwry) Find(ip string) (res ResultQQwry) {
	res = ResultQQwry{}
	res.IP = ip
	if strings.Count(ip, ".") != 3 {
		return res
	}
	offset := q.searchIndex(binary.BigEndian.Uint32(net.ParseIP(ip).To4()))
	if offset <= 0 {
		return
	}

	var area []byte
	mode := q.readMode(offset + 4)
	if mode == redirectMode1 {
		countryOffset := q.readUInt24()
		mode = q.readMode(countryOffset)
		if mode == redirectMode2 {
			c := q.readUInt24()
			area = q.readString(c)
		} else {
			area = q.readString(countryOffset)
		}
	} else if mode == redirectMode2 {
		countryOffset := q.readUInt24()
		area = q.readString(countryOffset)
	} else {
		area = q.readString(offset + 4)
	}

	enc := simplifiedchinese.GBK.NewDecoder()
	res.Area, _ = enc.String(string(area))

	return
}

type ResultQQwry struct {
	IP   string `json:"ip"`
	Area string `json:"area"`
}

// readMode 获取偏移值类型
func (q *QQwry) readMode(offset uint32) byte {
	mode := q.readData(1, int64(offset))
	return mode[0]
}

// readString 获取字符串
func (q *QQwry) readString(offset uint32) []byte {
	q.setOffset(int64(offset))
	data := make([]byte, 0, 30)
	for {
		buf := q.readData(1)
		if buf[0] == 0 {
			break
		}
		data = append(data, buf[0])
	}
	return data
}

// searchIndex 查找索引位置
func (q *QQwry) searchIndex(ip uint32) uint32 {
	header := q.readData(8, 0)

	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	for {
		mid := q.getMiddleOffset(start, end)
		buf := q.readData(indexLen, int64(mid))
		_ip := binary.LittleEndian.Uint32(buf[:4])

		if end-start == indexLen {
			offset := byteToUInt32(buf[4:])
			buf = q.readData(indexLen)
			if ip < binary.LittleEndian.Uint32(buf[:4]) {
				return offset
			}
			return 0
		}

		if _ip > ip {
			end = mid
		} else if _ip < ip {
			start = mid
		} else if _ip == ip {
			return byteToUInt32(buf[4:])
		}
	}
}

// readUInt24
func (q *QQwry) readUInt24() uint32 {
	buf := q.readData(3)
	return byteToUInt32(buf)
}

// getMiddleOffset
func (q *QQwry) getMiddleOffset(start uint32, end uint32) uint32 {
	records := ((end - start) / indexLen) >> 1
	return start + records*indexLen
}

// byteToUInt32 将 byte 转换为uint32
func byteToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}
