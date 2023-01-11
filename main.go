package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/db"
	"github.com/ethereum/HuiCollect/log"
	"github.com/ethereum/HuiCollect/services"
	"github.com/sirupsen/logrus"
)

var (
	confFile string
)

func init() {
	flag.StringVar(&confFile, "conf", "config.yaml", "conf file")
	flag.StringVar(&config.Env, "env", "dev", "env")
}

func main() {
	flag.Parse()
	logrus.Info(confFile)

	conf, err := config.LoadConf(confFile)
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
	err = log.Init(conf.AppName, conf.LogConf, conf.Env)
	if err != nil {
		log.Fatal(err)
	}

	leaseAlive()
	defer removeFile()
	logrus.Info("Hui-Collect started")

	//listen kill signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	//setup db connection
	collect_dbConnection, err := db.NewCollectMysql(&conf.CollectDataBase)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	hui_block_dbConnection, err := db.NewHuiBlockMysql(&conf.MonitorHui)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	eth_block_dbConnection, err := db.NewEthBlockMysql(&conf.MonitorEth)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	bsc_block_dbConnection, err := db.NewBscBlockMysql(&conf.MonitorBsc)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	btc_block_dbConnection, err := db.NewBtcBlockMysql(&conf.MonitorBtc)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	tron_block_dbConnection, err := db.NewTronBlockMysql(&conf.MonitorTron)
	if err != nil {
		logrus.Fatalf("connect to dbConnection error:%v", err)
	}

	//setup scheduler
	scheduler, err := services.NewServiceScheduler(conf, collect_dbConnection,
		hui_block_dbConnection, eth_block_dbConnection, bsc_block_dbConnection, btc_block_dbConnection, tron_block_dbConnection, sigCh)
	if err != nil {
		return
	}
	scheduler.Start()
}

var fName = `/tmp/hui.lock`

func removeFile() {
	_ = os.Remove(fName)
}

func leaseAlive() {
	f, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("create alive file err:%v", err))
	}
	now := time.Now().Unix()
	_, _ = fmt.Fprintf(f, "%d", now)
}
