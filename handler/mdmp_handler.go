package handler

import (
	"bufio"
	"context"
	"net"
	"newgateway/common"
	"newgateway/constant"
	"newgateway/logger"
	"newgateway/model"
	"newgateway/mqtt"
	"sync"
	"time"
)

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
		conn.Close()
		return
	}

	client := &mqtt.Client{
		Conn:          conn,
		Assure:        sync.Map{},
		SubscribeMap:  sync.Map{},
		Ticker:        10,
		Closing:       make(chan bool),
		Closed:        false,
		AssureClosing: make(chan bool),
	}
	go client.Tick()

	//读
	reader := bufio.NewReader(conn)
	connBuffer := make([]byte, 16*1024)
	connLen, err := reader.Read(connBuffer)
	if err != nil {
		logger.Fatal(err.Error())
		return
	}
	msgByteArr := connBuffer[:connLen]
	connMsgArr, err1 := mqtt.ParseMQTTMessage(msgByteArr)
	if err1 != nil {
		logger.Fatal(err1.Error())
		return
	}
	connMsg := connMsgArr[0]
	//logger.Info("accept type:" + strconv.Itoa(connMsg.FixedHeader.PackageType) + " \r\naccept message:" + string(msgByteArr))

	//dealConnect
	if connMsg.FixedHeader.PackageType == constant.MQTT_MSG_TYPE_CONNECT && h.dealConnect(connMsg, client) {
		var (
			buffer     [][]byte
			count      = 0
			bufferSize = 1000
		)
		for i := 0; i < bufferSize; i++ {
			buffer = append(buffer, make([]byte, 16*1024))
		}
		for {
			x := make(chan bool)
			//监听超时, 异步读取数据
			go func(num int) {
				//go func() {
				//buff := make([]byte, 16*1024)
				n, err := reader.Read(buffer[num])
				//n, err := reader.Read(buff)
				if err != nil {
					logger.Fatal(err.Error())
					client.Closing <- true
					return
				}
				go client.Deal(buffer[num][:n])
				//go client.Deal(buff[:n])
				x <- true
			}(count % bufferSize)
			//}()
			select {
			case <-x: //正常收取消息
				count++
				continue
			case <-time.After(time.Duration(3*connMsg.VariableHeader.KeepAliveTimer/2) * time.Second): //超时
				logger.Info("connection time out")
				h.activeConn.Delete(client)
				client.Close()
				return
			case <-client.Closing: //客户端关闭
				logger.Info("connection close")
				h.activeConn.Delete(client)
				client.Close()
				return
			}
		}
	}
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
