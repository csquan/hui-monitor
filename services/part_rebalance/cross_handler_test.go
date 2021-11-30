package part_rebalance

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/config"
	"github.com/starslabhq/hermes-rebalance/db"
	"github.com/starslabhq/hermes-rebalance/log"
	"github.com/starslabhq/hermes-rebalance/types"
	"math/big"
	"net/http"
	"testing"
)

var (
	confFile string
)

func init() {
	flag.StringVar(&confFile, "conf", "config.yaml", "conf file")
}
func TestCreateTreansfer(t *testing.T) {
	flag.Parse()
	logrus.Info(confFile)
	conf, err := config.LoadConf("../../"+confFile)
	if err != nil {
		logrus.Errorf("load config error:%v", err)
		return
	}

	if conf.ProfPort != 0 {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", conf.ProfPort), nil)
			if err != nil {
				panic(fmt.Sprintf("start pprof server err:%v", err))
			}
		}()
	}

	//setup log print
	err = log.Init(conf.AppName, conf.LogConf)
	if err != nil {
		log.Fatal(err)
	}

	dbtest, err := db.NewMysql(&config.DataBaseConf{
		DB: "root:123456sj@tcp(127.0.0.1:3306)/rebalance?charset=utf8mb4",
	})
	var tasks []*types.TransactionTask
	ReceiveFromBridgeParams := []*types.ReceiveFromBridgeParam{
		&types.ReceiveFromBridgeParam{
			ChainId:   1,
			ChainName: "bsc",
			From:      "606288c605942f3c84a7794c0b3257b56487263c",
			To:        "a929022c9107643515f5c777ce9a910f0d1e490c",
			Amount:    new(big.Int).SetInt64(100),
			TaskID:    new(big.Int).SetUint64(1),
		},
	}
	params := &types.Params{
		ReceiveFromBridgeParams:ReceiveFromBridgeParams,
	}
	data, _ := json.Marshal(params)
	task := &types.PartReBalanceTask{
		Base: &types.Base{ID: 10},
		Params: string(data),
	}
	c := &crossHandler{db:dbtest, clientMap: conf.ClientMap}
	tasks, err = c.CreateReceiveFromBridgeTask(task)
	if err != nil {
		logrus.Errorf("CreateReceiveFromBridgeTask error:%v task:[%v]", err, task)
		return
	}
	if tasks, err = c.SetNonceAndGasPrice(tasks); err != nil { //包含http，放在事物外面
		logrus.Errorf("SetNonceAndGasPrice error:%v task:[%v]", err, task)
		return
	}
	c.db.SaveTxTasks(dbtest.GetSession(), tasks)
}