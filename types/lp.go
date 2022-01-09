package types

//LP接口参数
type LPResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Ts   int64  `json:"ts"`
	Data *Data  `json:"data"`
}

type SingleStrategy struct {
	VaultAddress    string `json:"vaultAddress"`
	Amount          string `json:"amount"`
	StrategyAddress string `json:"strategyAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenAddress    string `json:"tokenAddress"`
	Platform        string `json:"platform"`
	Chain           string `json:"chain"`
	ChainId         int    `json:"chainId"`
	Decimal         int    `json:"decimal"`
}

type Data struct {
	Thresholds            []*Threshold         `json:"threshold"`
	VaultInfoList         []*VaultInfo         `json:"vaultInfoList"`
	LiquidityProviderList []*LiquidityProvider `json:"liquidityProviderList"`
	SingleList            []*SingleStrategy    `json:"singleList"`
}
type Threshold struct {
	TokenAddress    string `json:"tokenAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	Chain           string `json:"chain"`
	ChainId         int    `json:"chainId"`
	ThresholdAmount string `json:"thresholdAmount"`
	Decimal         int    `json:"decimal"`
}
type LiquidityProvider struct {
	Chain    string `json:"chain"`
	ChainId  int    `json:"chainId"`
	LpSymbol string `json:"lpSymbol"`
	//LpAmount       string    `json:"lpAmount"`
	LpTokenAddress string    `json:"lpTokenAddress"`
	LpPlatform     string    `json:"lpPlatform"`
	LpInfoList     []*LpInfo `json:"lpInfoList"`
}
type LpInfo struct {
	LpIndex           int    `json:"lpIndex"`
	LpAmount          string `json:"lpAmount"`
	BaseTokenAddress  string `json:"baseTokenAddress"`
	QuoteTokenAddress string `json:"quoteTokenAddress"`
	BaseTokenSymbol   string `json:"baseTokenSymbol"`
	QuoteTokenSymbol  string `json:"quoteTokenSymbol"`
	BaseTokenAmount   string `json:"baseTokenAmount"`
	QuoteTokenAmount  string `json:"quoteTokenAmount"`
	StrategyAddress   string `json:"strategyAddress"`
}

type Strategy struct {
	Addr        string `json:"strategyAddress"`
	TokenSymbol string `json:"tokenSymbol"`
}
type VaultInfo struct {
	TokenSymbol  string                            `json:"tokenSymbol"`
	Chain        string                            `json:"chain"`
	Currency     string                            `json:"currency"`
	ActiveAmount map[string]*ControllerInfo        `json:"activeAmount"`
	Strategies   map[string]map[string][]*Strategy `json:"strategies"`
}

type ControllerInfo struct {
	ActiveAmount      string `json:"activeAmount"`
	ControllerAddress string `json:"vaultAddress"`
	ClaimedReward     string `json:"claimedReward"`
	VaultAmount       string `json:"vaultAmount"`
	SoloAmount        string `json:"soloAmount"`
	//下面几个字段不是从json解析出来的
	//Chain string
	//ChainID int
	//BridgeAddress string
}

//平无常接口参数

type ImpermanectLostReq struct {
	BizNo  string   `json:"bizNo"`
	LpList []*LpReq `json:"lpList"`
}

type LpReq struct {
	Chain          string       `json:"chain"`
	LpTokenAddress string       `json:"lpTokenAddress"`
	LpAmount       string       `json:"lpAmount"`
	TokenList      []*TokenInfo `json:"tokenList"`
	//Token0OriginAmount string `json:"token0OriginAmount"`
	//Token1OriginAmount string `json:"token1OriginAmount"`
}
type TokenInfo struct {
	TokenAddress      string `json:"tokenAddress"`
	TokenOriginAmount string `json:"tokenOriginAmount"`
}

type NormalResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	//Data map[string]interface{} `json:"data"`
}

type MarginStatusResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

type TaskManagerResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}
