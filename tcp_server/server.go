package tcpserver

import (
	"context"
	"fmt"
	"interface/tcp"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Config stores tcp server properties
type Config struct {
	Address  string        `yaml:"address"`
	MaxCount uint32        `yaml:"maxCount"`
	TimeOut  time.Duration `yaml:"timeOut"`
}

// ClientCounter Record the number of clients (current server)
var ClientCounter int32

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	var closeChan = make(chan struct{})
	var sigCh = make(chan os.Signal, 1)

	// 接受中断信号
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	// 创建监听器
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	fmt.Printf("bind: %s, start listening...", cfg.Address)
	ListenAndServe(listener, handler, closeChan)

	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeClan <-chan struct{}) {
	errChan := make(chan error, 1)
	defer close(errChan)

	go func() {
		select {
		case <-closeClan:
			fmt.Printf("get exit signal")
		case err := <-errChan:
			fmt.Printf("accept error: %s", err.Error())
		}
		fmt.Printf("shutting down.....")
		_ = listener.Close()
		_ = handler.Close()
	}()

	var ctx = context.Background()
	var waitDone sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				fmt.Printf("accept occurs temporary error: %v, retry in 5ms", err)
				time.Sleep(5 * time.Millisecond)
				continue
			}
			errChan <- err
			break
		}

		fmt.Printf("accept link")
		ClientCounter++
		waitDone.Add(1)

		go func() {
			// 逻辑结束后, client减1
			defer func() {
				waitDone.Done()
				atomic.AddInt32(&ClientCounter, -1)
			}()
			// 处理函数
			handler.Hanlde(ctx, conn)
		}()
	}
	waitDone.Wait()
}
