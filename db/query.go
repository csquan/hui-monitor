package db

import (
	"github.com/ethereum/HuiCollect/types"
)

func (m *Mysql) GetMonitorInfo() ([]*types.Monitor, error) {
	monitors := make([]*types.Monitor, 0)
	err := m.engine.Table("t_monitor").Find(&monitors)
	if err != nil {
		return nil, err
	}
	return monitors, err
}

func (m *Mysql) GetMonitorCollectTask(addr string, chain string, threshold int) ([]*types.Asset, error) {
	tasks := make([]*types.Asset, 0)
	err := m.engine.Table("asset").Where("address = ? and chain = ? and balance > ?", addr, chain, threshold).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}
