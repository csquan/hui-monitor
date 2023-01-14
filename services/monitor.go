package services

import (
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
)

type MonitorService struct {
	collect_db types.IDB
	wallet_db  types.IDB

	config *config.Config
}

func NewMonitorService(collect_db types.IDB, wallet_db types.IDB, c *config.Config) *MonitorService {
	return &MonitorService{
		collect_db: collect_db,
		wallet_db:  wallet_db,
		config:     c,
	}
}

func (c *MonitorService) Run() (err error) {
	monitors, err := c.collect_db.GetMonitorInfo()

	if err != nil {
		return err
	}

	threshold := 0

	for _, monitor := range monitors {
		assets, err := c.wallet_db.GetMonitorCollectTask(monitor.Addr, monitor.Chain, threshold)
		if err != nil {
			logrus.Error(err)
		}

		err = c.HandleInsertCollect(assets)
		if err != nil {
			logrus.Error(err)
		}
	}

	return
}

func (c *MonitorService) HandleInsertCollect(txs []*types.Asset) error {
	for _, tx := range txs {
		err := utils.CommitWithSession(c.collect_db, func(s *xorm.Session) error {
			if err := c.collect_db.InsertCollectTx(s, tx); err != nil { //插入归集交易表
				logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, tx)
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("insert colelct transactidon task error:%v", err)
		}
	}
	return nil
}

func (c *MonitorService) Name() string {
	return "Monitor"
}
