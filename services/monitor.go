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
	db       types.IDB
	block_db types.IDB
	config   *config.Config
}

func NewMonitorService(db types.IDB, block_db types.IDB, c *config.Config) *MonitorService {
	return &MonitorService{
		db:       db,
		block_db: block_db,
		config:   c,
	}
}

func (c *MonitorService) Run() (err error) {
	monitors, err := c.db.GetMonitorInfo()

	if err != nil {
		return err
	}

	for _, monitor := range monitors {
		targetAddr := monitor.Addr //"0x206beddf4f9fc55a116890bb74c6b79999b14eb1"

		erc20_txs, err := c.block_db.GetMonitorCollectTask(targetAddr, monitor.Height)
		if err != nil {
			logrus.Error(err)
			continue
		}

		if len(erc20_txs) == 0 {
			logrus.Infof("no tx of target addr.")
			continue
		}

		for _, erc20_tx := range erc20_txs {
			collectTask := types.CollectTxDB{}
			collectTask.Copy(erc20_tx)

			collectTask.CollectState = int(types.TxReadyCollectState)

			err := utils.CommitWithSession(c.db, func(s *xorm.Session) error {
				if err := c.db.InsertCollectTx(s, &collectTask); err != nil { //插入归集交易表
					logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, collectTask)
					return err
				}
				//先看看monitor中有没有该地址，没有插入，有则更新
				count, err := c.db.GetMonitorCountInfo(targetAddr)
				if err != nil {
					logrus.Errorf("get monitor info error:%v addr:[%v]", err, targetAddr)
					return err
				}
				if count > 0 {
					if err := c.db.UpdateMonitor(collectTask.BlockNum, targetAddr); err != nil { //更新monitor
						logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, collectTask)
						return err
					}
				} else {
					monitor := types.Monitor{}
					monitor.Addr = targetAddr
					monitor.Height = collectTask.BlockNum

					if err := c.db.InsertMonitor(s, &monitor); err != nil { //插入monitor
						logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, collectTask)
						return err
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("insert colelct sub transactidon task error:%v", err)
			}
		}
	}

	return
}

func (c MonitorService) Name() string {
	return "Monitor"
}
