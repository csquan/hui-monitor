package types

type TransactionState int

const (
	TxReadyCollectState TransactionState = iota
	TxCollectingState
	TxCollectedState
)
