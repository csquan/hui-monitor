package types

import (
	"github.com/go-xorm/xorm"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock_db.go -package=mock
type IReader interface {
	//查询指定地址的监控表中的记录数
	GetMonitorCountInfo(Addr string) (int, error)
	//查询指定地址的监控表中的高度
	GetMonitorHeightInfo(Addr string) (int, error)
	//查询监控
	GetMonitorCollectTask(addr string, height uint64) ([]*TxErc20, error)

	GetMonitorInfo() ([]*Monitor, error)
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine

	InsertMonitor(itf xorm.Interface, monitor *Monitor) (err error)
	UpdateMonitor(height uint64, addr string) error

	SaveMonitorTask(itf xorm.Interface, monitor *Monitor) (err error)
	RemoveMonitorTask(addr string) error

	InsertCollectTx(itf xorm.Interface, task *CollectTxDB) (err error)
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
