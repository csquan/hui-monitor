package db

import (
	"time"

	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"xorm.io/core"
)

type Mysql struct {
	conf   *config.DataBaseConf
	engine *xorm.Engine
}

func NewMysql(conf *config.DataBaseConf) (m *Mysql, err error) {
	//"test:123@/test?charset=utf8"
	engine, err := xorm.NewEngine("mysql", conf.DB)
	if err != nil {
		logrus.Errorf("create engine error: %v", err)
		return
	}
	engine.ShowSQL(false)
	engine.Logger().SetLevel(core.LOG_DEBUG)
	location, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	engine.SetTZLocation(location)
	engine.SetTZDatabase(location)
	m = &Mysql{
		conf:   conf,
		engine: engine,
	}
	return
}

func NewBlockMysql(conf *config.MonitorConf) (m *Mysql, err error) {
	//"test:123@/test?charset=utf8"
	engine, err := xorm.NewEngine("mysql", conf.DB)
	if err != nil {
		logrus.Errorf("create engine error: %v", err)
		return
	}
	engine.ShowSQL(false)
	engine.Logger().SetLevel(core.LOG_DEBUG)
	location, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	engine.SetTZLocation(location)
	engine.SetTZDatabase(location)
	m = &Mysql{
		engine: engine,
	}
	return
}

func (m *Mysql) GetEngine() *xorm.Engine {
	return m.engine
}

func (m *Mysql) GetSession() *xorm.Session {
	return m.engine.NewSession()
}

func (m *Mysql) SaveMonitorTask(itf xorm.Interface, monitor *types.Monitor) (err error) {
	_, err = itf.Insert(monitor)
	if err != nil {
		logrus.Errorf("insert monitor task error:%v, tasks:%v", err, monitor)
	}
	return
}

func (m *Mysql) RemoveMonitorTask(addr string) error {
	_, err := m.engine.Exec("delete t_monitor where f_addr = ?", addr)
	return err
}

func (m *Mysql) InsertMonitor(itf xorm.Interface, monitor *types.Monitor) (err error) {
	_, err = itf.Insert(monitor)
	if err != nil {
		logrus.Errorf("insert collect task error:%v, monitor:%v", err, monitor)
	}
	return
}

func (m *Mysql) UpdateMonitor(height uint64, addr string) error {
	_, err := m.engine.Exec("update t_monitor set f_height = ? where f_addr = ?", height, addr)
	return err
}

func (m *Mysql) InsertCollectTx(itf xorm.Interface, task *types.CollectTxDB) (err error) {
	_, err = itf.Insert(task)
	if err != nil {
		logrus.Errorf("insert collect task error:%v, tasks:%v", err, task)
	}
	return
}
