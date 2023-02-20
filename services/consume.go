package services

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/ethereum/hui-monitor/config"
	"github.com/ethereum/hui-monitor/types"
	"github.com/ethereum/hui-monitor/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	kafka_sarama "github.com/suiguo/hwlib/kafka_sarama"
	"github.com/suiguo/hwlib/logger"
	"strings"
)

type ConsumeService struct {
	collect_db     types.IDB
	producer       kafka_sarama.Producer
	client_account kafka_sarama.Consumer
	client_tx      kafka_sarama.Consumer
	config         *config.Config
}

func (c *ConsumeService) ProduceKafka() {
	reg := types.RegisterData{}
	reg.APPID = "8888"
	reg.Eth = "0x2b1cd262A97CDF0b76c19bF68F97C277492C37DF"
	b, err := json.Marshal(&reg)
	if err != nil {
		logrus.Fatal(err)
	}

	err = c.producer.PushMsg(c.config.KafkaInfo.TopicAccount, string(b))
	if err != nil {
		logrus.Fatal(err)
	}
}

func NewConsumeService(collect_db types.IDB, c *config.Config) *ConsumeService {
	kafka_consumer_account, err := kafka_sarama.NewSarConsumer([]string{c.KafkaInfo.Url}, c.KafkaInfo.GroupId, logger.NewStdLogger(), kafka_sarama.WithConsumerAutoCommit(true), kafka_sarama.WithConsumerOffsets(kafka_sarama.OffsetOldest))
	if err != nil || kafka_consumer_account == nil {
		logrus.Fatal(err)
	}
	kafka_consumer_tx, err := kafka_sarama.NewSarConsumer([]string{c.KafkaInfo.Url}, c.KafkaInfo.GroupId, logger.NewStdLogger(), kafka_sarama.WithConsumerAutoCommit(true), kafka_sarama.WithConsumerOffsets(kafka_sarama.OffsetOldest))
	if err != nil || kafka_consumer_tx == nil {
		logrus.Fatal(err)
	}

	kafka_producer, err := kafka_sarama.NewSarProducer([]string{c.KafkaInfo.Url}, true, logger.NewStdLogger(), kafka_sarama.WithProductAcks(sarama.WaitForAll))
	if err != nil || kafka_producer == nil {
		logrus.Fatal(err)
	}
	return &ConsumeService{
		collect_db:     collect_db,
		config:         c,
		client_account: kafka_consumer_account,
		client_tx:      kafka_consumer_tx,
		producer:       kafka_producer,
	}
}

func getMonitor(reg *types.RegisterData, chain string) (*types.Monitor, error) {
	monitor := types.Monitor{}
	monitor.Addr = strings.ToLower(reg.Eth)
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
	c.client_account.SubscribeTopics([]string{c.config.KafkaInfo.TopicAccount}, func(topic string, partition int32, offset int64, msg []byte) error {
		logrus.Info(topic, partition, offset, string(msg))
		reg := types.RegisterData{}

		err = json.Unmarshal(msg, &reg)
		if err != nil {
			logrus.Info(err)
		}

		err = utils.CommitWithSession(c.collect_db, func(s *xorm.Session) error {
			if reg.Eth != "" {
				monitor1, err := getMonitor(&reg, "hui")
				if err != nil {
					logrus.Error(err)
				}

				if err := c.collect_db.InsertMonitor(s, monitor1); err != nil { //插入monitor
					logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor1)
					return err
				}

				monitor2, err := getMonitor(&reg, "eth")
				if err != nil {
					logrus.Error(err)
				}
				if err := c.collect_db.InsertMonitor(s, monitor2); err != nil { //插入monitor
					logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor2)
					return err
				}

				monitor3, err := getMonitor(&reg, "bsc")
				if err != nil {
					logrus.Error(err)
				}
				if err := c.collect_db.InsertMonitor(s, monitor3); err != nil { //插入monitor
					logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor3)
					return err
				}
			} else if reg.Btc != "" {
				monitor, err := getMonitor(&reg, "btc")
				if err != nil {
					logrus.Error(err)
				}

				if err := c.collect_db.InsertMonitor(s, monitor); err != nil { //插入monitor
					logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor)
					return err
				}
			} else if reg.Trx != "" {
				monitor, err := getMonitor(&reg, "trx")
				if err != nil {
					logrus.Error(err)
				}

				if err := c.collect_db.InsertMonitor(s, monitor); err != nil { //插入monitor
					logrus.Errorf("insert monitor task error:%v tasks:[%v]", err, monitor)
					return err
				}
			}
			return nil
		})
		return nil
	})

	c.client_tx.SubscribeTopics([]string{c.config.KafkaInfo.TopicTx}, func(topic string, partition int32, offset int64, msg []byte) error {
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

		err = utils.CommitWithSession(c.collect_db, func(s *xorm.Session) error {
			if err := c.collect_db.InsertMonitorTx(s, TxMonitor); err != nil { //插入monitor
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
