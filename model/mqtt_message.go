package model

type MQTTMessage struct {
	FixedHeader    *FixedHeader
	VariableHeader *VariableHeader
	Payload        *Payload
}

type FixedHeader struct {
	PackageType     int            //数据包类型
	SpecificToken   *SpecificToken //不同类型数据包的具体类型
	RemainingLength int            //剩余长度
}

type SpecificToken struct {
	DUP    int //0，表示当前为第一次发送, 1表示非第一次发送
	Qos    int
	Retain int //消息要推送给当前订阅者且持久保存, 后面的订阅者会收到最新一条retain=1的消息
}

type VariableHeader struct {
	ProtocolName      string
	ProtocolVersion   int
	ConnectFlags      *ConnectFlags
	KeepAliveTimer    int
	ConnectReturnCode int
	TopicName         string
	MessageId         int //要求在一个特定方向（服务器发往客户端为一个方向，客户端发送到服务器端为另一个方向）的通信消息中必须唯一
}

type ConnectFlags struct {
	UserNameFlag int
	PasswordFlag int
	WillRetain   int
	WillQos      int
	WillFlag     int
	CleanSession int //0，表示如果订阅的客户机断线了，要保存为其要推送的消息（QoS为1和QoS为2），若其重新连接时，需将这些消息推送（若客户端长时间不连接，需要设置一个过期值）; 1，断线服务器即清理相关信息，重新连接上来之后，会再次订阅。
	Reserved     int
}

type Payload struct {
	ClientId          string
	WillTopic         string
	WillMessage       string
	UserName          string
	Password          string
	SubscribePayload  string
	SubscribeAckQos   int
	UnsubscribeTopics string
	Data              string
}
