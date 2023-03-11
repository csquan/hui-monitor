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
	collectDb types.IDB
	config    *config.Config
}

func NewMonitorService(collectDb types.IDB, c *config.Config) *MonitorService {
	return &MonitorService{
		collectDb: collectDb,
		config:    c,
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
func (c *MonitorService) GetHotWallet(str string) ([]string, error) {
	str = str[1 : len(str)-1]
	arr := strings.Split(str, ",")
	return arr, nil
}

func (c *MonitorService) Run() (err error) {
	monitors, err := c.collectDb.GetMonitorInfo()

	if err != nil {
		return err
	}

	//这里首先排除热钱包地址
	hot_str, err := utils.GetHotWallets(c.config.WalletInfo.URL)
	if err != nil {
		logrus.Error(err)
		return
	}

	hot_msg := gjson.Get(hot_str, "message")

	_, err = c.GetHotWallet(hot_msg.String())
	if err != nil {
		logrus.Error(err)
		return
	}
	//for _, hotAddr := range hotAddrs {
	//	if len(hotAddr) > 1 {
	//		hotAddr = hotAddr[1 : len(hotAddr)-1]
	//		for _, monitor := range monitors {
	//			if hotAddr == monitor.Addr {
	//				logrus.Info("开始删除地址：监控地址 :" + monitor.Addr + "匹配到的热钱包地址: " + hotAddr)
	//				c.collectDb.DelCollectTask(monitor.Addr, monitor.Chain)
	//				return //todo:
	//			}
	//		}
	//	}
	//}

	//获取所有支持的mappedToken名称
	tokensArr, err := c.GetTokenInfo()
	if err != nil {
		logrus.Error(err)
	}
	for _, monitor := range monitors {
		logrus.Info(monitor)
		for _, tokens := range tokensArr {
			logrus.Info(tokens)
			var infos []map[string]interface{}
			err = json.Unmarshal([]byte(*tokens), &infos)
			if err != nil {
				logrus.Error(err)
				continue
			}
			for _, token := range infos {
				logrus.Info("当前token: " + token["symbol"].(string))
				//src_tx 中是否有相同地址的交易，且归集状态为未完成
				exist, err := c.GetSrcTx(token["chain"].(string), monitor.Addr, token["symbol"].(string))
				if err != nil {
					logrus.Info("GetSrcTx出错")
					logrus.Error(err)
				}
				if exist == true { //相同地址的交易存在且归集状态为未完成，则这里就不处理
					logrus.Info("相同地址的交易存在且归集未完成" + monitor.Addr + "token: " + token["symbol"].(string))
					continue
				}
				logrus.Info("相同地址的交易但是归集已经完成，可以继续进行，addr:" + monitor.Addr + "token: " + token["symbol"].(string))
				//得到账户的资产
				AssetsStr, err := c.GetUserAssets(token["chain"].(string), monitor.Addr, token["symbol"].(string))
				if err != nil {
					logrus.Info("GetUserAssets 错误返回:")
					logrus.Error(err)
				}
				logrus.Info("资产返回，Asset:" + AssetsStr)
				errorstr := gjson.Get(AssetsStr, "error")
				if errorstr.String() != "" { //钱包这里返回应该规范下
					logrus.Info("gjson错误返回:")
					continue
				}
				balance := gjson.Get(AssetsStr, "balance")
				logrus.Info("余额:" + balance.String())
				if balance.String() == "0" {
					logrus.Info("balacne为0返回:")
					continue
				}
				assets := types.Asset{}
				err = json.Unmarshal([]byte(AssetsStr), &assets)
				if err != nil {
					logrus.Info("Unmarshal错误返回:")
					logrus.Error(err)
					continue
				}
				logrus.Info(assets.Status)
				logrus.Info("PendingWithdrawalBalance" + assets.PendingWithdrawalBalance)
				// 资产状态不是冻结且PendingWithdrawalBalanc为0
				if assets.Status == 0 && assets.PendingWithdrawalBalance == "" {
					logrus.Info("资产状态不冻结且PendingWithdrawalBalanc不为0:")
					srcTx := c.getCollectSrcTx(assets, monitor.Uid)

					//插入归集源交易--如果这里有 相同链 相同symbol 相同地址 余额的交易就不插入？
					err = c.HandleInsertCollect(&srcTx)
					if err != nil {
						logrus.Error(err)
					}
				} else {
					logrus.Info("资产状态是冻结或者PendingWithdrawalBalanc为0:")
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
	exist, err := c.collectDb.GetSrcTx(chain, addr, symbol)
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
	logrus.Info(res)
	coinArray := strings.Split(res[1:len(res)-1], ",")
	logrus.Info(coinArray)
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
	err := utils.CommitWithSession(c.collectDb, func(s *xorm.Session) error {
		if err := c.collectDb.InsertCollectTx(s, tx); err != nil {
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
