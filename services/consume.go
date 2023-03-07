package services

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/ethereum/hui-monitor/config"
	"github.com/ethereum/hui-monitor/types"
	"github.com/ethereum/hui-monitor/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	kafkasarama "github.com/suiguo/hwlib/kafka_sarama"
	"github.com/suiguo/hwlib/logger"
)

type ConsumeService struct {
	collectDb     types.IDB
	producer      kafkasarama.Producer
	clientAccount kafkasarama.Consumer
	clientTx      kafkasarama.Consumer
	config        *config.Config
}

func (c *ConsumeService) ProduceKafka() {
	reg := types.RegisterData{}
	reg.APPID = "8888"
	reg.Eth = "0x2b1cd262A97CDF0b76c19bF68F97C277492C37DF"
	b, err := json.Marshal(&reg)
	if err != nil {
		logrus.Fatal(err)
	}

	err = c.producer.PushMsg(c.config.KafkaInfo.TopicAccount, b)
	if err != nil {
		logrus.Fatal(err)
	}
}

func NewConsumeService(collectDb types.IDB, c *config.Config) *ConsumeService {
	reConf := func(old []kafkasarama.Config) []kafkasarama.Config {
		if c.KafkaInfo.Pass != "" {
			old = append(old, kafkasarama.WithSASLAuth(c.KafkaInfo.User, c.KafkaInfo.Pass, kafkasarama.SHA_256))
		}
		if c.KafkaInfo.Tls {
			old = append(old, kafkasarama.WithTls("", "", "", false))
		}
		return old
	}
	consumerConf := make([]kafkasarama.Config, 0)
	consumerConf = append(consumerConf,
		kafkasarama.WithConsumerAutoCommit(true),
		kafkasarama.WithConsumerOffsets(kafkasarama.OffsetOldest),
	)
	consumerConf = reConf(consumerConf)
	kafkaConsumerAccount, err := kafkasarama.NewSarConsumer(
		c.KafkaInfo.Url,
		c.KafkaInfo.GroupId,
		logger.NewStdLogger(),
		consumerConf...,
	)
	if err != nil || kafkaConsumerAccount == nil {
		logrus.Fatal(err)
	}
	kafkaConsumerTx, err := kafkasarama.NewSarConsumer(
		c.KafkaInfo.Url,
		c.KafkaInfo.GroupId,
		logger.NewStdLogger(),
		consumerConf...,
	)
	if err != nil || kafkaConsumerTx == nil {
		logrus.Fatal(err)
	}

	prodConf := []kafkasarama.Config{kafkasarama.WithProductAcks(sarama.WaitForAll)}
	prodConf = reConf(prodConf)
	kafkaProducer, err := kafkasarama.NewSarProducer(
		c.KafkaInfo.Url,
		true,
		logger.NewStdLogger(),
		prodConf...,
	)
	if err != nil || kafkaProducer == nil {
		logrus.Fatal(err)
	}
	return &ConsumeService{
		collectDb:     collectDb,
		config:        c,
		clientAccount: kafkaConsumerAccount,
		clientTx:      kafkaConsumerTx,
		producer:      kafkaProducer,
	}
}

func getMonitor(reg *types.RegisterData, chain string) (*types.Monitor, error) {
	monitor := types.Monitor{}
	monitor.Addr = reg.Eth
	monitor.Chain = chain
	monitor.Uid = reg.UID
	monitor.AppId = reg.APPID
	return &monitor, nil
}

func getTxMonitor(tx *types.TxData) (*types.TxMonitor, error) {
	TxMonitor := types.TxMonitor{}
	TxMonitor.Chain = tx.Chain
	TxMonitor.Hash = tx.Hash
	TxMonitor.OrderID = tx.OrderId
	return &TxMonitor, nil
}

func (c *ConsumeService) Run() (err error) {
	c.clientAccount.SubscribeTopics([]string{c.config.KafkaInfo.TopicAccount}, func(topic string, partition int32, offset int64, msg []byte) error {
		logrus.Info(topic, partition, offset, string(msg))
		reg := types.RegisterData{}

		err = json.Unmarshal(msg, &reg)
		if err != nil {
			logrus.Info(err)
		}

		err = utils.CommitWithSession(c.collectDb, func(s *xorm.Session) error {
			if reg.Eth != "" {
				monitor1, err := getMonitor(&reg, "hui")
				if err != nil {
					logrus.Error(err)
				}

				if err := c.collectDb.InsertMonitor(s, monitor1); err != nil { //插入monitor
					logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor1)
					return err
				}
				//monitor2, err := getMonitor(&reg, "eth")
				//if err != nil {
				//	logrus.Error(err)
				//}
				//if err := c.collectDb.InsertMonitor(s, monitor2); err != nil { //插入monitor
				//	logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor2)
				//	return err
				//}

				//monitor3, err := getMonitor(&reg, "bsc")
				//if err != nil {
				//	logrus.Error(err)
				//}
				//if err := c.collectDb.InsertMonitor(s, monitor3); err != nil { //插入monitor
				//	logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor3)
				//	return err
				//}
			}
			if reg.Btc != "" {
				//monitor, err := getMonitor(&reg, "btc")
				//if err != nil {
				//	logrus.Error(err)
				//}
				//
				//if err := c.collectDb.InsertMonitor(s, monitor); err != nil { //插入monitor
				//	logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor)
				//	return err
				//}
			}
			if reg.Trx != "" {
				monitor, err := getMonitor(&reg, "trx")
				if err != nil {
					logrus.Error(err)
				}

				if err := c.collectDb.InsertMonitor(s, monitor); err != nil { //插入monitor
					logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor)
					return err
				}
			}
			return nil
		})
		return nil
	})

	c.clientTx.SubscribeTopics([]string{c.config.KafkaInfo.TopicTx}, func(topic string, partition int32, offset int64, msg []byte) error {
		logrus.Info(topic, partition, offset, string(msg))
		TxData := types.TxData{}

		logrus.Info("receive msg and begin parse")
		err = json.Unmarshal(msg, &TxData)
		if err != nil {
			logrus.Info(err)
		}
		logrus.Info(TxData)

		TxMonitor, err := getTxMonitor(&TxData)
		if err != nil {
			logrus.Error(err)
		}

		err = utils.CommitWithSession(c.collectDb, func(s *xorm.Session) error {
			TxMonitor.GetReceiptTimes = 0
			TxMonitor.ReceiptState = -1
			if err := c.collectDb.InsertMonitorTx(s, TxMonitor); err != nil { //插入monitor
				logrus.Errorf("insert tx monitor task error:%v tasks:[%v]", err, TxData)
				return err
			}
			return nil
		})
		return nil
	})

	return nil
}

func (c *ConsumeService) Name() string {
	return "Consume"
}
