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
	"strings"
)

type MonitorService struct {
	collect_db types.IDB

	config *config.Config
}

func NewMonitorService(collect_db types.IDB, c *config.Config) *MonitorService {
	return &MonitorService{
		collect_db: collect_db,
		config:     c,
	}
}
func (c *MonitorService) getCollectSrcTx(asset types.Asset, uid string) types.CollectSrcTx {
	srcTx := types.CollectSrcTx{
		Chain:        asset.Chain,
		Symbol:       asset.Symbol,
		Address:      asset.Address,
		Uid:          uid,
		Balance:      asset.Balance,
		Status:       asset.Status,
		OwnerType:    asset.OwnerType,
		CollectState: 0,
	}
	return srcTx

}
func (c *MonitorService) Run() (err error) {
	monitors, err := c.collect_db.GetMonitorInfo()

	if err != nil {
		return err
	}

	//获取所有支持的mappedToken名称
	tokensArr, err := c.GetTokenInfo()
	if err != nil {
		logrus.Error(err)
	}
	for _, monitor := range monitors {
		for _, tokens := range tokensArr {
			var infos []map[string]interface{}
			err = json.Unmarshal([]byte(*tokens), &infos)
			if err != nil {
				logrus.Error(err)
				continue
			}
			for _, token := range infos {
				//src_tx 中是否有相同地址的交易，且归集状态为未完成
				exist, err := c.GetSrcTx(token["chain"].(string), monitor.Addr, token["symbol"].(string))
				if err != nil {

				}
				if exist == true { //相同地址的交易存在且归集状态为未完成，则这里就不处理
					continue
				}
				//得到账户的资产
				AssetsStr, err := c.GetUserAssets(token["chain"].(string), monitor.Addr, token["symbol"].(string))
				if err != nil {
					logrus.Error(err)
				}
				errorstr := gjson.Get(AssetsStr, "error")
				if errorstr.String() != "" { //钱包这里返回应该规范下
					continue
				}
				assets := types.Asset{}
				err = json.Unmarshal([]byte(AssetsStr), &assets)
				if err != nil {
					logrus.Error(err)
					continue
				}
				// 资产状态不是冻结且PendingWithdrawalBalanc为0
				if assets.Status == 0 && assets.PendingWithdrawalBalance == "" {
					srcTx := c.getCollectSrcTx(assets, monitor.Uid)

					//插入归集源交易
					err = c.HandleInsertCollect(&srcTx)
					if err != nil {
						logrus.Error(err)
					}
				}
			}
		}
	}

	return
}

func (c *MonitorService) GetUserAssets(chain string, addr string, symbol string) (string, error) {
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
	return res, nil
}

func (c *MonitorService) GetSrcTx(chain string, addr string, symbol string) (bool, error) {
	exist, err := c.collect_db.GetSrcTx(chain, addr, symbol)
	return exist, err
}

func (c *MonitorService) GetTokenInfo() ([]*string, error) {
	assetStrs := make([]*string, 0)
	url := c.config.WalletInfo.URL + "/" + "getSupportedMappedToken"
	res, err := utils.Get(url)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if res == "[]" {
		return nil, nil
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

// 插入归集源交易表
func (c *MonitorService) HandleInsertCollect(tx *types.CollectSrcTx) error {
	err := utils.CommitWithSession(c.collect_db, func(s *xorm.Session) error {
		if err := c.collect_db.InsertCollectTx(s, tx); err != nil {
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
