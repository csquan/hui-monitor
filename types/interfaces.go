package types

import (
	"github.com/go-xorm/xorm"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock_db.go -package=mock
type IReader interface {
	GetMonitorInfo() ([]*Monitor, error)
	GetSrcTx(chain string, addr string, symbol string) (bool, error)
	GetUncollectedSrcTx() ([]*CollectSrcTx, error)
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine

	InsertMonitor(itf xorm.Interface, monitor *Monitor) (err error)
	InsertCollectTx(itf xorm.Interface, task *CollectSrcTx) (err error)
	InsertMonitorTx(itf xorm.Interface, monitor *TxMonitor) (err error)

	UpdateCollectTx(itf xorm.Interface, task *CollectSrcTx) (err error)
}

type IDB interface {
	IReader
	IWriter
}

type IAsyncService interface {
	Name() string
	Run() error
}
