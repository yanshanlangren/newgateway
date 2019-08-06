package main

import (
	"context"
	"fmt"
	"net"
	"newgateway/config"
	"newgateway/handler"
	"newgateway/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//handler := handler.NewEchoHandler()
	handler := &handler.MDMPHandler{
	}
	ListenAndServe(":"+config.GetConfig().Server.Port, handler)
}
func ListenAndServe(address string, handler handler.Handler) {
	// 绑定监听地址
	listener, err := net.Listen(config.GetConfig().Server.Network, address)
	if err != nil {
		logger.Error(fmt.Sprintf("listen err: %v", err))
	}

	// 监听中断信号
	var closing bool
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// 收到中断信号后开始关闭流程
			logger.Info("shuting down...")
			// 设置标志位为关闭中, 使用原子操作保证线程可见性
			closing = true
			// listener 关闭后 listener.Accept() 会立即返回错误
			listener.Close()
		}
	}()

	defer listener.Close()
	logger.Info(fmt.Sprintf("bind: %s, start listening...", address))

	ctx, _ := context.WithCancel(context.Background())
	for {
		// Accept 会一直阻塞直到有新的连接建立或者listen中断才会返回
		conn, err := listener.Accept()
		if err != nil {
			// 通常是由于listener被关闭无法继续监听导致的错误
			logger.Error(fmt.Sprintf("accept err: %v", err))
		}
		// 开启新的 goroutine 处理该连接
		logger.Info("accept link")
		go handler.Handle(ctx, conn)
	}
}
