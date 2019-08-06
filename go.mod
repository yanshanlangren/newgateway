module newgateway

go 1.12

require (
	git.internal.yunify.com/MDMP2/cloudevents v0.0.0-20190417033153-b26740c15a29
	github.com/Shopify/sarama v1.19.0
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/btree v1.0.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/pierrec/lz4 v2.2.4+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/rcrowley/go-metrics v0.0.0-20190706150252-9beb055b7962 // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.4.0
	github.com/tebeka/strftime v0.1.3 // indirect
	google.golang.org/grpc v1.20.1 // indirect
	gopkg.in/yaml.v2 v2.2.2
	qingcloud.com/qing-cloud-mq v0.0.0-00010101000000-000000000000
)

replace qingcloud.com/qing-cloud-mq => ../../qing-cloud-mq/

replace git.internal.yunify.com/MDMP2/api => ../api/

replace git.internal.yunify.com/MDMP2/cloudevents => ../cloudevents/
