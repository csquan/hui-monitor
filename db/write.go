package db

import (
	"fmt"
	"time"

	"github.com/ethereum/hui-monitor/config"
	"github.com/ethereum/hui-monitor/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"xorm.io/core"
)

type Mysql struct {
	conf   *config.CollectDataBaseConf
	engine *xorm.Engine
}

func NewCollectMysql(conf *config.CollectDataBaseConf) (m *Mysql, err error) {
	Engine, err := xorm.NewEngine("mysql", conf.DB)
	if err != nil {
		logrus.Errorf("create engine error: %v", err)
		return
	}
	Engine.ShowSQL(false)
	Engine.Logger().SetLevel(core.LOG_DEBUG)
	location, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	Engine.SetTZLocation(location)
	Engine.SetTZDatabase(location)

	m = &Mysql{
		conf:   conf,
		engine: Engine,
	}
	return
}

func (m *Mysql) GetEngine() *xorm.Engine {
	return m.engine
}

func (m *Mysql) GetSession() *xorm.Session {
	return m.engine.NewSession()
}

func (m *Mysql) InsertMonitor(itf xorm.Interface, monitor *types.Monitor) (err error) {
	_, err = itf.Insert(monitor)
	if err != nil {
		logrus.Errorf("insert collect task error:%v, monitor:%v", err, monitor)
	}
	return
}

func (m *Mysql) InsertMonitorTx(itf xorm.Interface, monitorTx *types.TxMonitor) (err error) {
	_, err = itf.Insert(monitorTx)
	if err != nil {
		logrus.Errorf("insert monitor Tx task error:%v, monitorTx:%v", err, monitorTx)
	}
	return
}

func (m *Mysql) InsertCollectTx(itf xorm.Interface, task *types.CollectSrcTx) (err error) {
	_, err = itf.Insert(task)
	if err != nil {
		logrus.Errorf("insert collect task error:%v, tasks:%v", err, task)
	}
	return
}

func (m *Mysql) UpdateCollectTx(state int, id uint64) (err error) {
	sql := fmt.Sprintf("update t_src_tx set f_collect_state = %d  where f_id = %d", state, id)
	_, err = m.engine.Exec(sql)
	if err != nil {
		logrus.Errorf("update collect task error:%v, id:%v,", err, id)
	}
	return
}
