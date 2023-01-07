package db

import (
	"fmt"
	"github.com/ethereum/HuiCollect/types"
)

func (m *Mysql) GetMonitorCountInfo(Addr string) (int, error) {
	count := 0
	sql := fmt.Sprintf("select count(*) from t_monitor where f_addr = \"%s\";", Addr)
	ok, err := m.engine.SQL(sql).Limit(1).Get(&count)
	if err != nil {
		return count, err
	}
	if !ok {
		return count, nil
	}

	return count, err
}

func (m *Mysql) GetMonitorHeightInfo(Addr string) (int, error) {
	height := 0
	sql := fmt.Sprintf("select * from t_monitor where f_addr = \"%s\";", Addr)
	ok, err := m.engine.SQL(sql).Limit(1).Get(&height)
	if err != nil {
		return height, err
	}
	if !ok {
		return height, nil
	}

	return height, err
}

func (m *Mysql) GetMonitorInfo() ([]*types.Monitor, error) {
	monitors := make([]*types.Monitor, 0)
	err := m.engine.Table("t_monitor").Find(&monitors)
	if err != nil {
		return nil, err
	}
	return monitors, err
}

func (m *Mysql) GetMonitorCollectTask(addr string, height uint64) ([]*types.TxErc20, error) {
	tasks := make([]*types.TxErc20, 0)
	err := m.engine.Table("tx_erc20").Where("receiver = ? and block_num > ?", addr, height).OrderBy("block_num").Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err
}
