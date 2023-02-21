package services

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/hui-monitor/config"
	"github.com/ethereum/hui-monitor/types"
	"github.com/ethereum/hui-monitor/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type UpdateService struct {
	collect_db types.IDB

	config *config.Config
}

func NewUpdateService(collect_db types.IDB, c *config.Config) *UpdateService {
	return &UpdateService{
		collect_db: collect_db,
		config:     c,
	}
}

func (c *UpdateService) GetCollectState(orderId string) (string, error) {
	param := types.WithdrawParam{
		OrderId: orderId,
	}
	msg, err1 := json.Marshal(param)
	if err1 != nil {
		logrus.Error(err1)
		//return nil, err1
	}
	url := c.config.WalletInfo.URL + "/" + "getWithdrawalOrderByOrderID"
	res, err1 := utils.Post(url, msg)
	if err1 != nil {
		logrus.Error(err1)
		//return nil, err1
	}
	return res, nil
}

// 查询src_tx中的orderId--srcTx中未完成状态的appid，根据这个去查状态更新
func (c *UpdateService) Run() (err error) {
	txs, err := c.collect_db.GetUncollectedSrcTx()
	if err != nil {
		return err
	}
	for _, tx := range txs {
		str, err := c.GetCollectState(tx.OrderId)
		if err != nil {
			logrus.Error(err)
			return err
		}
		state := gjson.Get(str, "status")
		logrus.Info(state.String())
	}

	return nil
}

// 更新归集源交易表状态
func (c *UpdateService) HandleUpdateCollect(tx *types.CollectSrcTx) error {
	err := utils.CommitWithSession(c.collect_db, func(s *xorm.Session) error {
		if err := c.collect_db.UpdateCollectTx(s, tx); err != nil {
			logrus.Errorf("insert colelct transaction task error:%v tasks:[%v]", err, tx)
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("insert colelct transactidon task error:%v", err)
	}

	return nil
}

func (c *UpdateService) Name() string {
	return "Update"
}
