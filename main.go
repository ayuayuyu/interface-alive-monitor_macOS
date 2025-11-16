package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// カーネルとのsocket作成
	fd, err := syscall.Socket(syscall.AF_ROUTE, syscall.SOCK_RAW, 0)
	if err != nil {
		fmt.Printf("failed to create socket: %v\n", err)
		os.Exit(1)
	}
	// NOTE: defer で閉じるとシグナル ハンドラで明示的に Close したときに二重閉鎖になる
	// ため明示的に最後に Close する。

	fmt.Println("ネットワーク監視の開始...")

	// メッセージを流すチャネル（バッファ付き）とワーカープール
	messages := make(chan []byte, 100)
	var wg sync.WaitGroup
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for b := range messages {
				// 各ワーカーでパース処理を行う
				ParseRouteMessage(b)
			}
			fmt.Printf("worker %d: exit\n", id)
		}(i)
	}

	// シグナルで安全にシャットダウンするためのゴルーチン
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nシグナル検出、シャットダウン...")
		// 読み取りブロックを解除するためにソケットを閉じる
		syscall.Close(fd)
		// ワーカーに終了を伝える
		close(messages)
	}()

	buf := make([]byte, 2048)
	for {
		n, err := syscall.Read(fd, buf)
		if err != nil {
			// ソケットが閉じられた等で読み取りが失敗したらループを抜ける
			fmt.Printf("failed to read from socket: %v\n", err)
			break
		}
		if n == 0 {
			continue
		}
		// データはスライスの背後バッファを参照しているため
		// コピーしてチャネルに流す
		b := make([]byte, n)
		copy(b, buf[:n])
		select {
		case messages <- b:
		default:
			// キューが一杯の時は落とすという選択肢
			fmt.Println("message queue full, drop message")
		}
	}

	// ループを抜けたらワーカーの終了を待つ
	fmt.Println("main loop 終了, ワーカー待機...")
	wg.Wait()
	// 最後にソケットを閉じる（まだ開いていれば）
	syscall.Close(fd)
	fmt.Println("終了")
}
