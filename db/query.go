package db

import (
	"github.com/ethereum/hui-monitor/types"
)

func (m *Mysql) GetMonitorInfo() ([]*types.Monitor, error) {
	monitors := make([]*types.Monitor, 0)
	err := m.engine.Table("t_monitor").Find(&monitors)
	if err != nil {
		return nil, err
	}
	return monitors, err
}

func (m *Mysql) GetSrcTx(chain string, addr string, symbol string) (bool, error) {
	txs := make([]*types.CollectSrcTx, 0)
	err := m.engine.Table("t_src_tx").Where("f_chain=? and f_symbol =? and f_address=?", chain, symbol, addr).Find(&txs)
	if err != nil {
		return false, err
	}
	return len(txs) > 0, err
}

func (m *Mysql) GetUncollectedSrcTx() ([]*types.CollectSrcTx, error) {
	txs := make([]*types.CollectSrcTx, 0)
	err := m.engine.Table("t_src_tx").Where("f_collect_state !=2").Find(&txs)
	if err != nil {
		return nil, err
	}
	return txs, err
}
