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
	conf       *config.CollectDataBaseConf
	huiEngine  *xorm.Engine
	tronEngine *xorm.Engine
}

func NewCollectMysql(conf *config.CollectDataBaseConf) (m *Mysql, err error) {
	huiEngine, err := xorm.NewEngine("mysql", conf.HuiDB)
	if err != nil {
		logrus.Errorf("create engine error: %v", err)
		return
	}
	huiEngine.ShowSQL(false)
	huiEngine.Logger().SetLevel(core.LOG_DEBUG)
	location, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}
	huiEngine.SetTZLocation(location)
	huiEngine.SetTZDatabase(location)

	tronEngine, err := xorm.NewEngine("mysql", conf.TronDB)
	if err != nil {
		logrus.Errorf("create engine error: %v", err)
		return
	}
	tronEngine.ShowSQL(false)
	tronEngine.Logger().SetLevel(core.LOG_DEBUG)
	if err != nil {
		return nil, err
	}
	tronEngine.SetTZLocation(location)
	tronEngine.SetTZDatabase(location)

	m = &Mysql{
		conf:       conf,
		huiEngine:  huiEngine,
		tronEngine: tronEngine,
	}
	return
}

func (m *Mysql) GetEngine() *xorm.Engine {
	return m.huiEngine
}

func (m *Mysql) GetSession() *xorm.Session {
	return m.huiEngine.NewSession()
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
	_, err = m.huiEngine.Exec(sql)
	if err != nil {
		logrus.Errorf("update collect task error:%v, id:%v,", err, id)
	}
	return
}
