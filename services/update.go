package services

import (
	"encoding/json"
	"github.com/ethereum/hui-monitor/config"
	"github.com/ethereum/hui-monitor/types"
	"github.com/ethereum/hui-monitor/utils"
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
		AppId:   "",
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
		status := gjson.Get(str, "status")
		logrus.Info("查询归集状态：tx.OrderId:" + tx.OrderId + "status:" + status.String())
		if status.Int() == 21 || status.Int() == 30 || status.Int() == 50 || status.Int() == 60 || status.Int() == 100 { //都认为成功，更新状态
			logrus.Info("更新该笔状态为完成")
			tx.CollectState = int(types.TxCollectedState)
			c.HandleUpdateCollect(tx)
		}
	}

	return nil
}

// 更新归集源交易表状态
func (c *UpdateService) HandleUpdateCollect(tx *types.CollectSrcTx) error {
	if err := c.collect_db.UpdateCollectTx(tx.CollectState, tx.ID); err != nil {
		logrus.Errorf("update colelct transaction task error:%v tasks:[%v]", err, tx)
		return err
	}
	return nil
}

func (c *UpdateService) Name() string {
	return "Update"
}
