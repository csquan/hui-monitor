package services

import (
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
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
	cli, err := kafka.GetDefaultKafka(kafka.ALLType, "192.168.31.243:9092", "monitor_group", kafka.Earliest, nil)
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

func (c *ConsumeService) Run() (err error) {
	//c.ProduceKafka()
	data := c.client.MessageChan()
	out_msg := <-data
	fmt.Println(string(out_msg.Value))

	return nil
}

func (c *ConsumeService) Name() string {
	return "Consume"
}
