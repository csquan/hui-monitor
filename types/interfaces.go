package types

import (
	"github.com/go-xorm/xorm"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock_db.go -package=mock
type IReader interface {
	GetMonitorInfo() ([]*Monitor, error)

	//根据链，符号，地址 查询归集源交易
	GetSrcTx(chain string, addr string, symbol string) (bool, error)
	//查询所有未归集完成的源交易
	GetUncollectedSrcTx() ([]*CollectSrcTx, error)
}

type IWriter interface {
	GetSession() *xorm.Session
	GetEngine() *xorm.Engine

	//插入监控地址
	InsertMonitor(itf xorm.Interface, monitor *Monitor) (err error)
	//插入归集源交易
	InsertCollectTx(itf xorm.Interface, task *CollectSrcTx) (err error)

	//插入nikki的监控交易，爬快匹配上使用，与归集无关
	InsertMonitorTx(itf xorm.Interface, monitor *TxMonitor) (err error)

	//更新归集交易状态
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
