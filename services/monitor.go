package services

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/hui-monitor/config"
	"github.com/ethereum/hui-monitor/types"
	"github.com/ethereum/hui-monitor/utils"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"strings"
)

type MonitorService struct {
	collect_db types.IDB
	wallet_db  types.IDB

	config *config.Config
}

func NewMonitorService(collect_db types.IDB, c *config.Config) *MonitorService {
	return &MonitorService{
		collect_db: collect_db,
		config:     c,
	}
}

func (c *MonitorService) Run() (err error) {
	monitors, err := c.collect_db.GetMonitorInfo()

	if err != nil {
		return err
	}

	tokensArr, err := c.GetTokenInfo()
	if err != nil {
		logrus.Error(err)
	}
	logrus.Info(tokensArr)
	for _, monitor := range monitors {
		for _, tokens := range tokensArr {
			var infos []map[string]interface{}
			err = json.Unmarshal([]byte(*tokens), &infos)
			if err != nil {
				logrus.Error(err)
				continue
			}
			for _, token := range infos {
				balance, err := c.GetUserBalance(token["chain"].(string), monitor.Addr, token["symbol"].(string))
				if err != nil {
					logrus.Error(err)
				}
				logrus.Info(balance)
				assets := types.Asset{}
				err = c.HandleInsertCollect(&assets)
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	}

	return
}

func (c *MonitorService) GetUserBalance(chain string, addr string, symbol string) (string, error) {
	param := types.AssetInParam{
		Symbol:  symbol,
		Chain:   chain,
		Address: addr,
	}
	msg, err1 := json.Marshal(param)
	if err1 != nil {
		logrus.Error(err1)
		//return nil, err1
	}
	url := c.config.WalletInfo.URL + "/" + "getAsset"
	res, err1 := utils.Post(url, msg)
	if err1 != nil {
		logrus.Error(err1)
		//return nil, err1
	}
	fmt.Println(res)
	return res, nil
}

func (c *MonitorService) GetTokenInfo() ([]*string, error) {
	assetStrs := make([]*string, 0)
	url := c.config.WalletInfo.URL + "/" + "getSupportedMappedToken"
	res, err := utils.Get(url)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	coinArray := strings.Split(res[1:len(res)-1], ",")

	url = c.config.WalletInfo.URL + "/" + "getMappedToken"
	for _, coin := range coinArray {
		param := types.Coin{
			MappedSymbol: coin[1 : len(coin)-1],
		}
		msg, err1 := json.Marshal(param)
		if err1 != nil {
			logrus.Error(err1)
			return nil, err1
		}
		res, err1 := utils.Post(url, msg)
		if err1 != nil {
			logrus.Error(err1)
			return nil, err1
		}
		assetStrs = append(assetStrs, &res)
	}
	return assetStrs, nil
}
func (c *MonitorService) HandleInsertCollect(tx *types.Asset) error {
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

	return nil
}

func (c *MonitorService) Name() string {
	return "Monitor"
}
