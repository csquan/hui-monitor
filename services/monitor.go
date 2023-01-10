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
	collect_db    types.IDB
	hui_block_db  types.IDB
	eth_block_db  types.IDB
	bsc_block_db  types.IDB
	btc_block_db  types.IDB
	tron_block_db types.IDB

	config *config.Config
}

func NewMonitorService(collect_db types.IDB, hui_block_db types.IDB, eth_block_db types.IDB, bsc_block_db types.IDB,
	btc_block_db types.IDB, tron_block_db types.IDB, c *config.Config) *MonitorService {
	return &MonitorService{
		collect_db:    collect_db,
		hui_block_db:  hui_block_db,
		eth_block_db:  eth_block_db,
		bsc_block_db:  bsc_block_db,
		btc_block_db:  btc_block_db,
		tron_block_db: tron_block_db,
		config:        c,
	}
}

func (c *MonitorService) Run() (err error) {
	monitors, err := c.collect_db.GetMonitorInfo()

	if err != nil {
		return err
	}

	for _, monitor := range monitors {
		//根据链类型分别调用GetMonitorCollectTask
		hui_erc20_txs := make([]*types.TxErc20, 0)
		bsc_erc20_txs := make([]*types.TxErc20, 0)
		eth_erc20_txs := make([]*types.TxErc20, 0)

		targetAddr := monitor.Addr

		switch monitor.Chain {
		case "eth":
			hui_erc20_txs, err = c.hui_block_db.GetMonitorCollectTask(targetAddr, monitor.Height)
			if err != nil {
				logrus.Error(err)
			}

			bsc_erc20_txs, err = c.eth_block_db.GetMonitorCollectTask(targetAddr, monitor.Height)
			if err != nil {
				logrus.Error(err)
			}

			eth_erc20_txs, err = c.bsc_block_db.GetMonitorCollectTask(targetAddr, monitor.Height)
			if err != nil {
				logrus.Error(err)
			}
			break
		case "btc":
			break
		case "tron":
			break
		default:

		}

		if len(hui_erc20_txs) == 0 && len(bsc_erc20_txs) == 0 && len(eth_erc20_txs) == 0 {
			logrus.Infof("no tx of target addr.")
			continue
		}
		err = c.HandleInsertCollect(hui_erc20_txs, "hui", targetAddr)
		if err != nil {
			logrus.Error(err)
		}
		err = c.HandleInsertCollect(bsc_erc20_txs, "bsc", targetAddr)
		if err != nil {
			logrus.Error(err)
		}
		err = c.HandleInsertCollect(eth_erc20_txs, "eth", targetAddr)
		if err != nil {
			logrus.Error(err)
		}
	}

	return
}

func (c *MonitorService) HandleInsertCollect(txs []*types.TxErc20, chainName string, targetAddr string) error {
	for _, erc20_tx := range txs {
		collectTask := types.CollectTxDB{}
		collectTask.Copy(erc20_tx)
		collectTask.Chain = chainName

		collectTask.CollectState = int(types.TxReadyCollectState)

		err := utils.CommitWithSession(c.collect_db, func(s *xorm.Session) error {
			if err := c.collect_db.InsertCollectTx(s, &collectTask); err != nil { //插入归集交易表
				logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, collectTask)
				return err
			}
			//先看看monitor中有没有该地址，没有插入，有则更新
			count, err := c.collect_db.GetMonitorCountInfo(targetAddr)
			if err != nil {
				logrus.Errorf("get monitor info error:%v addr:[%v]", err, targetAddr)
				return err
			}
			if count > 0 {
				if err := c.collect_db.UpdateMonitor(collectTask.BlockNum, targetAddr); err != nil { //更新monitor
					logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, collectTask)
					return err
				}
			} else {
				monitor := types.Monitor{}
				monitor.Addr = targetAddr
				monitor.Height = collectTask.BlockNum

				if err := c.collect_db.InsertMonitor(s, &monitor); err != nil { //插入monitor
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
	return nil
}

func (c *MonitorService) Name() string {
	return "Monitor"
}
