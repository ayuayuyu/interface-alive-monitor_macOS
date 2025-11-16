package main

import "syscall"

// Cの rt_msghdr 構造体をGoで再現
type RtMsghdr struct {
	RtmMsglen  uint16
	RtmVersion uint8
	RtmType    uint8
	RtmIndex   uint16
	RtmFlags   int32
	RtmAddrs   int32
	RtmPid     int32
	RtmSeq     int32
}

type IfMsghdr struct {
	IfmMsglen  uint16
	IfmVersion uint8
	IfmType    uint8
	IfmAddrs   int32
	IfmFlags   int32
	IfmIndex   uint16
	_          [2]byte // padding
	// struct if_data は省略
}

// Cの ifa_msghdr 構造体をGoで再現
type IfaMsghdr struct {
	IfamMsglen  uint16
	IfamVersion uint8
	IfamType    uint8
	IfamAddrs   int32
	IfamFlags   int32
	IfamIndex   uint16
	_           [2]byte // padding
	IfamMetric  int32
}

// Cの ifma_msghdr 構造体をGoで再現
type IfmaMsghdr struct {
	IfmamType  uint8
	IfmamAddrs int32
	IfmamFlags int32
	IfmamIndex uint16
	_          [2]byte // padding
}

// Cの sockaddr 構造体の先頭部分 (Goで型判別に利用)
type Sockaddr struct {
	Len    uint8
	Family uint8
}

// アドレスマスクの定数 (net/route.h より)
const (
	RTAX_DST     = 0 // 宛先アドレス
	RTAX_GATEWAY = 1 // ゲートウェイアドレス
	RTAX_NETMASK = 2 // ネットマスク
	RTAX_GENMASK = 3
	RTAX_IFP     = 4 // インターフェース名
	RTAX_IFA     = 5 // インターフェースのIPアドレス
	RTAX_AUTHOR  = 6
	RTAX_BRD     = 7 // ブロードキャストアドレス
	RTAX_MAX     = 8
)

func familyName(f int) string {
	switch f {
	case syscall.AF_INET:
		return "IPv4"
	case syscall.AF_INET6:
		return "IPv6"
	}
	return "Unknown"
}

func roundup(a, size int) int {
	return (a + size - 1) & ^(size - 1)
}
