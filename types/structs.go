package types

import (
	"math/big"
	"time"
)

type Base struct {
	ID        uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

type TransactionTask struct {
	*Base        `xorm:"extends"`
	ID           uint64    `xorm:"f_id not null pk autoincr bigint(20)" gorm:"primary_key"`
	ParentID     uint64    `xorm:"f_parent_id"`
	UserID       string    `xorm:"f_uid"`
	UUID         int64     `xorm:"f_uuid"`
	RequestId    string    `xorm:"f_request_id"`
	Nonce        uint64    `xorm:"f_nonce"`
	GasPrice     string    `xorm:"f_gas_price"`
	GasLimit     string    `xorm:"f_gas_limit"`
	ChainId      int       `xorm:"f_chain_id"`
	From         string    `xorm:"f_from"`
	To           string    `xorm:"f_to"`
	ContractAddr string    `xorm:"f_contract_addr"`
	Receiver     string    `xorm:"f_receiver"`
	Amount       string    `xorm:"f_amount"`
	Value        string    `xorm:"f_value"`
	InputData    string    `xorm:"f_input_data"`
	SignHash     string    `xorm:"f_sign_hash"`
	TxHash       string    `xorm:"f_tx_hash"`
	State        int       `xorm:"f_state"`
	Tx_type      int       `xorm:"f_type"`
	Receipt      string    `xorm:"f_receipt"`
	Sig          string    `xorm:"f_sig"`
	Error        string    `xorm:"f_error"`
	Times        int       `xorm:"f_retry_times"`
	CreatedAt    time.Time `xorm:"created f_created_at"`
	UpdatedAt    time.Time `xorm:"updated f_updated_at"`
}

type CollectTxDB struct {
	*Base          `xorm:"extends"`
	Id             uint64 `xorm:"f_id"`
	Hash           string `xorm:"f_tx_hash"`
	Addr           string `xorm:"f_addr"`
	Sender         string `xorm:"f_sender"`
	Receiver       string `xorm:"f_receiver"`
	TokenCnt       string `xorm:"f_token_cnt"`
	TokenCntOrigin string `xorm:"f_token_cnt_origin"`
	LogIndex       int    `xorm:"f_log_index"`
	BlockState     uint8  `xorm:"f_block_state"`
	BlockNum       uint64 `xorm:"f_block_num"`
	BlockTime      uint64 `xorm:"f_block_time"`
	CollectState   int    `xorm:"f_collect_state"`
	Chain          string `xorm:"f_chain"`
}

type TxErc20 struct {
	Id             uint64 `xorm:"id"`
	Hash           string `xorm:"tx_hash"`
	Addr           string `xorm:"addr"`
	Sender         string `xorm:"sender"`
	Receiver       string `xorm:"receiver"`
	TokenCnt       string `xorm:"token_cnt"`
	TokenCntOrigin string `xorm:"token_cnt_origin"`
	LogIndex       int    `xorm:"log_index"`
	BlockState     uint8  `xorm:"block_state"`
	BlockNum       uint64 `xorm:"block_num"`
	BlockTime      uint64 `xorm:"block_time"`
}

func (c *CollectTxDB) Copy(tx *TxErc20) {
	c.Id = tx.Id
	c.Hash = tx.Hash
	c.Addr = tx.Addr
	c.Sender = tx.Sender
	c.Receiver = tx.Receiver

	c.TokenCnt = tx.TokenCnt
	c.TokenCntOrigin = tx.TokenCntOrigin
	c.LogIndex = tx.LogIndex
	c.BlockState = tx.BlockState
	c.BlockNum = tx.BlockNum

	c.BlockTime = tx.BlockTime
}

type Account struct {
	*Base        `xorm:"extends"`
	Id           uint64    `xorm:"f_id"`
	Addr         string    `xorm:"f_addr"`
	Balance      string    `xorm:"f_balance"`
	UpdatedAt    time.Time `xorm:"updated f_updated_at"`
	Lastcheck    string    `xorm:"f_lastcheck"`
	ContractAddr string    `xorm:"f_contractAddr"`
}

type Monitor struct {
	*Base  `xorm:"extends"`
	Id     uint64 `xorm:"f_id"`
	Addr   string `xorm:"f_addr"`
	Height uint64 `xorm:"f_height"`
}

type Token struct {
	*Base       `xorm:"extends"`
	Threshold   string `xorm:"f_currency"`
	Chain       string `xorm:"f_chain"`
	Symbol      string `xorm:"f_symbol"`
	Address     string `xorm:"f_address"`
	Decimal     int    `xorm:"f_decimal"`
	CrossSymbol string `xorm:"f_cross_symbol"`
}

func (t *Token) TableName() string {
	return "t_token"
}

func (t *Account) TableName() string {
	return "t_account"
}

func (t *Monitor) TableName() string {
	return "t_monitor"
}

func (t *TxErc20) TableName() string {
	return "tx_erc20"
}

func (t *CollectTxDB) TableName() string {
	return "t_src_tx"
}

func (t *TransactionTask) TableName() string {
	return "t_transaction_task"
}

type HttpRes struct {
	RequestId string `json:"requestId"`
	Hash      string `json:"hash"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Status    int    `json:"status"`
}

type Data1 struct {
	UID string `json:"uid" `
}

// HttpData success data
type HttpData struct {
	Code int `json:"code" example:"0"`
	Data Data1
}

type Data struct {
	UID string `json:"uid" `
}

type Balance_Erc20 struct {
	Id             string `xorm:"id"`
	Addr           string `xorm:"addr"`
	ContractAddr   string `xorm:"contract_addr"`
	Balance        string `xorm:"balance"`
	Height         string `xorm:"height"`
	Balance_Origin string `xorm:"balance_origin"`
}

type Tx struct {
	TxType               string
	From                 string
	To                   string
	Hash                 string
	Index                string
	Value                string
	Input                string
	Nonce                string
	GasPrice             string
	GasLimit             string
	GasUsed              string
	IsContract           string
	IsContractCreate     string
	BlockTime            string
	BlockNum             string
	BlockHash            string
	ExecStatus           string
	CreateTime           string
	BlockState           string
	MaxFeePerGas         string //交易费上限
	BaseFee              string
	MaxPriorityFeePerGas string //小费上限
	BurntFees            string //baseFee*gasused
}

type Erc20Transfer struct {
	TxHash          string
	Addr            string //合约地址
	Sender          string
	Receiver        string
	Tokens          *big.Int
	LogIndex        int
	SenderBalance   *big.Int
	ReceiverBalance *big.Int
}

type Erc20Info struct {
	Id                   string `xorm:"f_id"`
	Addr                 string `xorm:"f_addr"`
	Name                 string `xorm:"f_name"`
	Symbol               string `xorm:"f_symbol"`
	Decimals             string `xorm:"f_decimals"`
	Totoal_Supply        string `xorm:"f_total_supply"`
	Totoal_Supply_Origin string `xorm:"f_total_supply_origin"`
	Create_Time          string `xorm:"f_create_time"`
}

type SignData struct {
	UID     string
	Address string
	Hash    string
}

type SigData struct {
	Signature string "json:signature"
}

type CallBackData struct {
	RequestID string
	Hash      string
}
