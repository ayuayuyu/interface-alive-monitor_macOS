package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
	"unsafe"
)

// 受信した生のバイト列をパースする関数
func ParseRouteMessage(data []byte) {
	// ヘッダサイズよりデータが短い場合は処理しない
	headerSize := int(unsafe.Sizeof(RtMsghdr{}))
	if len(data) < headerSize {
		return
	}

	// ヘッダ部分をパーズ（データ構造に変換）
	var hdr RtMsghdr
	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.LittleEndian, &hdr); err != nil {
		log.Println("binary.Read failed:", err)
		return
	}

	// メッセージタイプで処理を分岐
	switch hdr.RtmType {
	case syscall.RTM_ADD:
		// fmt.Println("RTM_ADD (1): 経路（ルート）の追加")
	case syscall.RTM_DELETE:
		// fmt.Println("RTM_DELETE (2): 経路の削除")
	case syscall.RTM_CHANGE:
		// fmt.Println("RTM_CHANGE (3): 経路情報の変更")
	case syscall.RTM_GET:
		// fmt.Println("RTM_GET (4): 経路情報の要求")
	case syscall.RTM_LOSING:
		// fmt.Println("RTM_LOSING (5): 経路の通信品質低下")
	case syscall.RTM_REDIRECT:
		// fmt.Println("RTM_REDIRECT (6): 経路変更の指示")
	case syscall.RTM_MISS:
		// fmt.Println("RTM_MISS (7): 経路検索の失敗（キャッシュミス）")
	case syscall.RTM_LOCK:
		// fmt.Println("RTM_LOCK (8): 経路情報の一部のロック")
	case syscall.RTM_RESOLVE:
		// fmt.Println("RTM_RESOLVE (11): リンク層アドレスの解決要求")
	case syscall.RTM_NEWADDR:
		fmt.Println("RTM_NEWADDR (12): インターフェースへの新しいアドレス追加")
		parseAddressMessage(data)
	case syscall.RTM_DELADDR:
		fmt.Println("RTM_DELADDR (13): インターフェースからのアドレス削除")
		parseAddressMessage(data)
	case syscall.RTM_IFINFO:
		fmt.Println("RTM_IFINFO (14): インターフェースの状態変化")
		parseIfMessage(data)
	case syscall.RTM_NEWMADDR:
		// fmt.Println("RTM_NEWMADDR (15): 新しいマルチキャストグループメンバーの追加")
	case syscall.RTM_DELMADDR:
		// fmt.Println("RTM_DELMADDR (16): マルチキャストグループメンバーの削除")
	default:
		fmt.Printf("未対応のメッセージタイプです: Type=%d (0x%x)\n", hdr.RtmType, hdr.RtmType)
	}
}

// インターフェースメッセージのパース
func parseIfMessage(data []byte) {
	headerSize := int(unsafe.Sizeof(IfMsghdr{}))
	if len(data) < headerSize {
		return
	}
	var ifm IfMsghdr
	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.LittleEndian, &ifm); err != nil {
		log.Println("binary.Read failed:", err)
		return
	}

	idx := ifm.IfmIndex
	ifName, _ := net.InterfaceByIndex(int(idx))

	fmt.Printf("Interface: %s (index=%d), flags=0x%x\n", ifName.Name, ifm.IfmIndex, ifm.IfmFlags)

	if ifm.IfmFlags&syscall.IFF_UP != 0 {
		fmt.Println("Status: UP")
	} else {
		fmt.Println("Status: DOWN")
	}
}

func parseAddressMessage(data []byte) {
	var ifa IfaMsghdr
	headerSize := int(unsafe.Sizeof(IfaMsghdr{}))
	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.LittleEndian, &ifa); err != nil {
		log.Println("binary.Read failed:", err)
		return
	}

	idx := ifa.IfamIndex
	ifName, _ := net.InterfaceByIndex(int(idx))
	fmt.Printf("Interface: %s (index=%d)\n", ifName.Name, idx)

	addrs := ifa.IfamAddrs
	p := data[headerSize:]

	for i := 0; i < RTAX_MAX; i++ {
		if addrs&(1<<uint(i)) == 0 {
			continue
		}

		saLen := int(p[0])
		family := int(p[1])

		// sockaddr の実データをパース
		_, ip := parseSockaddr(p)

		if i == RTAX_IFA && ip != "" {
			if ifa.IfamType == syscall.RTM_NEWADDR {
				fmt.Printf("  [%s] 追加された IP: %s\n", familyName(family), ip)
			} else if ifa.IfamType == syscall.RTM_DELADDR {
				fmt.Printf("  [%s] 削除された IP: %s\n", familyName(family), ip)
			}
		}

		// sockaddr は 4バイトアラインメント
		p = p[roundup(saLen, 4):]
		if len(p) == 0 {
			break
		}
	}
}

// sockaddr を取り出して IP を返す
func parseSockaddr(data []byte) (int, string) {
	if len(data) < 2 {
		return 0, ""
	}
	saLen := int(data[0])
	family := int(data[1])

	// fmt.Printf("saLen: %d , family: %d\n", saLen, family)
	// fmt.Printf("data: %v\n", data)

	switch family {
	case syscall.AF_INET:
		if saLen >= 8 && len(data) >= 8 {
			ip := net.IPv4(data[4], data[5], data[6], data[7])
			return saLen, ip.String()
		}
	case syscall.AF_INET6:
		if saLen >= 26 && len(data) >= 8+16 {
			ip := net.IP(data[8 : 8+16])
			return saLen, ip.String()
		}
	}
	return saLen, ""
}
