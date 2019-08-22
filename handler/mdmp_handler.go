package handler

import (
	"bufio"
	"context"
	"io"
	"net"
	"newgateway/common"
	"newgateway/config"
	"newgateway/constant"
	"newgateway/logger"
	"newgateway/model"
	"newgateway/mqtt"
	"strconv"
	"sync"
	"time"
)

var bufferSize = config.GetConfig().Server.BufferSize

type MDMPHandler struct {
	//用于保存所有有效的连接
	activeConn sync.Map

	//关闭状态标识位
	closing common.AtomicBool
}

//关闭handler, 并关闭所有活跃的connectivity
func (h *MDMPHandler) Close() error {
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

//异步hand MDMP消息, 并管理连接
func (h *MDMPHandler) Handle(ctx context.Context, conn net.Conn) {
	//如果handler正在关闭, 则关闭当前连接
	if h.closing.Get() {
		if conn != nil {
			conn.Close()
		}
		return
	}

	client := &mqtt.Client{
		Conn:          conn,
		Assure:        sync.Map{},
		SubscribeMap:  sync.Map{},
		Ticker:        config.GetConfig().Server.TickerInterval,
		Closing:       make(chan bool),
		Closed:        false,
		AssureClosing: make(chan bool),
		Buffer:        make([]byte, 1024*1024),
		IsBufferEmpty: true,
		BufferOffset:  0,
	}
	go client.Tick()

	//读
	reader := bufio.NewReader(conn)

	var (
		buff    = make([]byte, bufferSize*1024)
		n       = 0
		err     error
		connMsg *model.MQTTMessage
	)
	n, err = reader.Read(buff)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	connMsg, err = mqtt.ParseByteArray(buff[:n])
	if err != nil {
		logger.Error("error parsing connection string: ", buff[:n], err)
		return
	}
	logger.Debug("accept type: ", strconv.Itoa(connMsg.FixedHeader.PackageType), " accept message: ", buff[:n])

	//dealConnect
	if connMsg.FixedHeader.PackageType == constant.MQTT_MSG_TYPE_CONNECT && h.dealConnect(connMsg, client) {
		x := make(chan bool)
		dur := time.Duration(3*connMsg.VariableHeader.KeepAliveTimer/2) * time.Second
		timeout := time.NewTimer(dur)
		for {
			//监听超时, 异步读取数据
			go func() {
				timeout.Reset(dur)
				n, err = reader.Read(buff)
				logger.Debug("received bytes of ", n, " ", err)
				if n > 0 {
					client.DealByteArray(buff[:n])
					x <- true
					return
				}
				if err != nil {
					if err == io.EOF {
						logger.Info("connection closed by client")
					} else {
						client.Will(connMsg)
						logger.Error(err.Error())
					}
					client.Closing <- true
					return
				}
			}()
			select {
			case <-x: //正常收取消息
				continue
			case <-timeout.C: //超时
				logger.Warn("connection time out")
				client.Will(connMsg)
				h.activeConn.Delete(client)
				timeout.Stop()
				close(x)
				client.Close()
				return
			case <-client.Closing: //客户端关闭
				h.activeConn.Delete(client)
				timeout.Stop()
				close(x)
				client.Close()
				return
			}
		}
	}
	client.Close()
}

//创建连接
func (h *MDMPHandler) dealConnect(msg *model.MQTTMessage, cli *mqtt.Client) bool {
	//TODO 验证身份

	//保存连接
	h.activeConn.Store(cli, msg)

	//产生返回值
	res := &model.MQTTMessage{
		FixedHeader: &model.FixedHeader{
			PackageType:     constant.MQTT_MSG_TYPE_CONNECTACK,
			RemainingLength: 2,
		},
		VariableHeader: &model.VariableHeader{
			ConnectReturnCode: constant.MQTT_CONNECT_RETURN_CODE_ACCEPTED,
		},
	}
	go cli.Write(res)
	return true
}
