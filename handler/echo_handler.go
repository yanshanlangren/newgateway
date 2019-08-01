package handler

import (
	"bufio"
	"context"
	"io"
	"net"
	"newgateway/common"
	"newgateway/logger"
	"newgateway/mqtt"
	"sync"
)

type EchoHandler struct {
	// 保存所有工作状态client的集合(把map当set用)
	// 需使用并发安全的容器
	activeConn sync.Map

	// 和 tcp server 中作用相同的关闭状态标识位
	closing common.AtomicBool
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		//关闭handler,并拒绝新的连接进入
		conn.Close()
	}

	client := &mqtt.Client{
		Conn: conn,
	}
	h.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				h.activeConn.Delete(conn)
			} else {
				logger.Warn(err)
			}
			return
		}
		logger.Info(msg)

		// 发送数据前先置为waiting状态
		client.Waiting.Add(1)

		// 模拟关闭时未完成发送的情况
		//logger.Info("sleeping")
		//time.Sleep(10 * time.Second)

		b := []byte("response: "+msg)
		conn.Write(b)
		// 发送完毕, 结束waiting
		client.Waiting.Done()
	}
}

func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// TODO: concurrent wait
	// 尝试关闭所有的客户端
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		cli := key.(*mqtt.Client)
		cli.Close()
		return true
	})
	return nil
}
