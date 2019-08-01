module newgateway

go 1.12

require (
	git.internal.yunify.com/MDMP2/cloudevents v0.0.0-20190417033153-b26740c15a29
	github.com/Shopify/sarama v1.19.0
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/btree v1.0.0 // indirect
	github.com/imroc/biu v0.0.0-20170329141542-0376ce6761c0
	github.com/pierrec/lz4 v2.2.4+incompatible // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190706150252-9beb055b7962 // indirect
	google.golang.org/grpc v1.20.1 // indirect
	gopkg.in/yaml.v2 v2.2.2
	qingcloud.com/qing-cloud-mq v0.0.0-00010101000000-000000000000
)

replace qingcloud.com/qing-cloud-mq => ../../qing-cloud-mq/

replace git.internal.yunify.com/MDMP2/api => ../api/

replace git.internal.yunify.com/MDMP2/cloudevents => ../cloudevents/
