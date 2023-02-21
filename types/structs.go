package types

import (
	"time"
)

type Base struct {
	ID        uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

// 资产表 --getAsset的返回
type Asset struct {
	Base
	Chain                    string
	Symbol                   string
	Address                  string
	Status                   int
	OwnerType                int
	Balance                  string
	PendingWithdrawalBalance string
	UsedFee                  string
	FundedFee                string
}

type Monitor struct {
	*Base `xorm:"extends"`
	Uid   string `xorm:"f_uid"`
	AppId string `xorm:"f_appid"`
	Chain string `xorm:"f_chain"`
	Addr  string `xorm:"f_addr"`
}

// nikki的hash监控kfaka
type TxMonitor struct {
	*Base   `xorm:"extends"`
	Hash    string `xorm:"f_hash"`
	Chain   string `xorm:"f_chain"`
	OrderID string `xorm:"f_order_id"`
	Push    string `xorm:"f_push"`
}

func (t *TxMonitor) TableName() string {
	return "t_monitor_hash"
}

func (t *Monitor) TableName() string {
	return "t_monitor"
}

type RegisterData struct {
	UID   string `json:"uid"`
	APPID string `json:"app_id"`
	Eth   string `json:"eth"`
	Btc   string `json:"btc"`
	Trx   string `json:"trx"`
}

type TxData struct {
	Chain          string `json:"chain"`
	Hash           string `json:"hash"`
	TxHeight       uint64 `json:"tx_height"`        //交易所在高度
	CurChainHeight uint64 `json:"cur_chain_height"` //当前高度
	OrderId        string `json:"order_id"`         //回调地址
}

type CollectSrcTx struct {
	*Base        `xorm:"extends"`
	Chain        string `xorm:"f_chain"`
	Symbol       string `xorm:"f_symbol"`
	Address      string `xorm:"f_address"`
	Uid          string `xorm:"f_uid"`
	Balance      string `xorm:"f_balance"`
	Status       int    `xorm:"f_status"`
	OwnerType    int    `xorm:"f_ownerType"`
	CollectState int    `xorm:"f_collect_state"`
	OrderId      string `xorm:"f_order_id"`
}

func (t *CollectSrcTx) TableName() string {
	return "t_src_tx"
}

type Coin struct {
	MappedSymbol string `json:"mapped_symbol"`
}

type AssetInParam struct {
	Symbol  string `json:"symbol"`
	Chain   string `json:"chain"`
	Address string `json:"address"`
}

type WithdrawParam struct {
	AppId   string `json:"app_id"`
	OrderId string `json:"order_id"`
}
