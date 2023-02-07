package services

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/suiguo/hwlib/kafka"
)

type ConsumeService struct {
	collect_db types.IDB
	client     kafka.KafaClient
	config     *config.Config
}

func (c *ConsumeService) ProduceKafka() {
	err := c.client.Produce("registrar-user-created", &kafka.KafkaMsg{
		Msg: []byte("hello word"),
	})
	if err != nil {
		fmt.Println(err)
	}
}

func NewConsumeService(collect_db types.IDB, c *config.Config) *ConsumeService {
	cli, err := kafka.GetDefaultKafka(kafka.ALLType, "kafka:9092", "monitor_group", kafka.Earliest, nil)
	if err != nil {
		fmt.Println(err)
	}
	err = cli.Subscribe("registrar-user-created")
	if err != nil {
		fmt.Println(err)
	}
	return &ConsumeService{
		collect_db: collect_db,
		config:     c,
		client:     cli,
	}
}

func getMonitor(reg *types.RegisterData, chain string) (*types.Monitor, error) {
	monitor := types.Monitor{}
	monitor.Addr = reg.Eth
	monitor.Chain = "hui"
	monitor.Uid = reg.UID
	monitor.AppId = reg.APPID
	return &monitor, nil
}

func (c *ConsumeService) Run() (err error) {
	//c.ProduceKafka()
	data := c.client.MessageChan()
	out_msg := <-data
	reg_data := string(out_msg.Value)
	fmt.Println(reg_data)

	reg := types.RegisterData{}

	err = json.Unmarshal([]byte(reg_data), &reg)
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
}

func (c *ConsumeService) Name() string {
	return "Consume"
}
