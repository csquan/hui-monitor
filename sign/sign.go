package sign

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// head key, case insensitive
const (
	headKeyData              = "date"
	headKeyXAmzDate          = "X-Amz-Date"
	headKeyAuthorization     = "authorization"
	headKeyHost              = "host"
	iSO8601BasicFormat       = "20060102T150405Z"
	iSO8601BasicFormatShort  = "20060102"
	queryKeySignature        = "X-Amz-Signature"
	queryKeyAlgorithm        = "X-Amz-Algorithm"
	queryKeyCredential       = "X-Amz-Credential"
	queryKeyDate             = "X-Amz-Date"
	queryKeySignatureHeaders = "X-Amz-SignedHeaders"
	aws4HmacSha256Algorithm  = "AWS4-HMAC-SHA256"
	AwsV4SigHeader           = "signer.blockchain.amazonaws.com"
)

var lf = []byte{'\n'}

// Key holds a set of Amazon Security Credentials.
type Key struct {
	AccessKey string
	SecretKey string
}

type Payload struct {
	Addrs         []string `json:"addrs"`
	Data          string   `json:"data"`
	Chain         string   `json:"chain"`
	EncryptParams string   `json:"encrypt_params"`
}

type SigReqData struct {
	//ToTag is the input data for contract revoking params
	ToTag    string `json:"to_tag"`
	//Asset    string `json:"asset"`
	Decimal  int    `json:"decimal"`
	//Platform string `json:"platform"`
	Nonce    int    `json:"nonce"`
	From     string `json:"from"`
	//To is the contract Addr
	To string `json:"to"`
	//GasLimit here
	FeeStep string `json:"fee_step"`
	//GasPrice here
	FeePrice string `json:"fee_price"`
	//FeeAsset string `json:"fee_asset"`
	Amount   string `json:"amount"`
	//ChainId  string `json:"chain_id"`
	TaskType  string     `json:"taskType"`  //目前支持withdraw, deploy, callContract
}

type ReqData struct {
	//ToTag is the input data for contract revoking params
	ToTag    string `json:"to_tag"`
	Asset    string `json:"asset"`
	Decimal  int    `json:"decimal"`
	Platform string `json:"platform"`
	Nonce    int    `json:"nonce"`
	From     string `json:"from"`
	//To is the contract Addr
	To string `json:"to"`
	//GasLimit here
	FeeStep string `json:"fee_step"`
	//GasPrice here
	FeePrice string `json:"fee_price"`
	FeeAsset string `json:"fee_asset"`
	Amount   string `json:"amount"`
	ChainId  string `json:"chain_id"`
	//TaskType  string     `json:"taskType"`  //目前支持withdraw, deploy, callContract
}

type EncParams struct {
	//Tasks  []Task        `json:"tasks"`
	//TxType string        `json:"tx_type"`
	From     string      `json:"from"`   //from_address
	To       string      `json:"to"`     //to_address -- call contract: contract Address
	Value    string      `json:"value"`  //value -- call contract: 0
	InputData string     `json:"inputData"` //contract input data
	Chain     string     `json:"chain"`     //destination chain
	Quantity  string     `json:"quantity"`  //token number
	ToAddress string     `json:"toAddress"` //receipt address
	ToTag     string     `json:"toTag"`     //contract input data
	TaskType  string     `json:"taskType"`  //目前支持withdraw, deploy, callContract

}

type Task struct {
	TaskId     string `json:"task_id"`
	UserId     string `json:"user_id"`
	OriginAddr string `json:"origin_addr"`
	TaskType   string `json:"task_type"`
}

type Response struct {
	Result bool     `json:"result"`
	Data   RespData `json:"data"`
}

type RespData struct {
	EncryptData string `json:"encrypt_data"`
	Extra       RespEx `json:"extra"`
}

type RespEx struct {
	Cipher string `json:"cipher"`
	TxHash string `json:"txhash"`
}

func fetchNonce(ctx context.Context, archnode, addr string) (int, error) {
	client, err := ethclient.Dial(archnode)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	//addr in hex string
	commonAddr := common.HexToAddress(addr)
	nonce, err := client.NonceAt(ctx, commonAddr, nil)
	if err != nil {
		return 0, err
	}
	return int(nonce), nil
}

//fetchPendingNonce for sending raw tx
func fetchPendingNonce(ctx context.Context, archnode, addr string) (int, error) {
	client, err := ethclient.Dial(archnode)
	if err != nil {
		logrus.Errorf("There is dialing error %v", err)
		return 0, err
	}
	defer client.Close()
	//addr in hex string
	commonAddr := common.HexToAddress(addr)
	nonce, err := client.PendingNonceAt(ctx, commonAddr)
	if err != nil {
		return 0, err
	}
	return int(nonce), nil
}

//estimate the gas before sending to chain
func EstimateGas(archNode, datastr, value, chain string, conf *Conf) (uint64, error) {
	//archNode := viper.GetString("RPCserver." + chain + ".nodeUrl")
	client, err := ethclient.Dial(archNode)
	if err != nil {
		return 0, err
	}
	toaddr := common.HexToAddress(conf.Vip.GetString("common.bridgeContract." + chain))

	va, _ := new(big.Int).SetString(value, 10)

	da, err := hexutil.Decode("0x" + datastr)
	if err != nil {
		return 0, err
	}

	msg := ethereum.CallMsg{
		From:     common.HexToAddress(conf.Vip.GetString("gateway." + chain + ".sysAddr")),
		To:       &toaddr,
		GasPrice: big.NewInt(40000000000),
		Value:    va,
		Data:     da,
	}

	gas, err1 := client.EstimateGas(context.TODO(), msg)
	//if there is error when estimate the gas, make it 400000 as default
	if err1 != nil {
		return uint64(400000), err1
	}
	return gas, nil
}

//estimate GasPrice on chain before sending
func EstimateGasPrice12(archNode string) (uint64, error) {
	client, err := ethclient.Dial(archNode)
	if err != nil {
		//if there is error when estimate the gas, make it 150Gwei as default
		return uint64(150000000000), err
	}
	price, err1 := client.SuggestGasPrice(context.TODO())
	//if there is error when estimate the gas, make it 150Gwei as default
	if err1 != nil {
		return uint64(150000000000), err1
	}
	return price.Uint64() * 6 / 5, nil
}

type UnData struct {
	FromAddr string
	Gas      int
	GasPrice decimal.Decimal
	Nonce    int
	//Proof        string
	UnsignedData []byte
	//TaskNonce    string
}

/*
//UnsignDataEvmChain to fetch the unsigned data
func UnsignDataEvmChain(nonce int, chain, dataStr string, conf *Conf) (*UnData, error) {
	archNode := conf.Vip.GetString("RPCserver." + chain + ".nodeUrl")
	sysAddr := conf.Vip.GetString("gateway." + chain + ".sysAddr")
	amstr := "0"

	//the estimate gas is
	gas, err := EstimateGas(archNode, dataStr, amstr, chain, conf)
	if err != nil {
		//there is error when estimate gasLimit, make it 400000 as default
		gas = uint64(400000)
		logrus.Debugf("estimate gas error, use default gas %d instead", gas)
	}

	logrus.Infof("The estimate gas is %d", gas)
	gaslimit := strconv.FormatUint(gas*3/2, 10)
	logrus.Infof("The gaslimit for contract interaction 150 is %s", gaslimit)

	//fetch fromAddr nonce
	//nonce, err := getNonce(chain)
	//if err != nil {
	//	logrus.Errorf("There is error when fetch nonce during signing gateway: %v", err)
	//	return nil, fmt.Errorf("there is error when fetch nonce during signing gateway: %v", err)
	//}

	//estimate gas price
	price, err := EstimateGasPrice12(archNode)
	if err != nil {
		logrus.Debugf("Estimate price error, use default gasPrice 150Gwei")
		price = uint64(150000000000)
	}
	feePrice := strconv.FormatUint(price, 10)

	//the bridge contract address
	contractAddr := conf.Vip.GetString("common.bridgeContract." + chain)
	//assemble the data field for sending transaction
	reqData := &ReqData{
		To:    contractAddr,
		ToTag: dataStr,
		Nonce: nonce,
		//The Asset used in different chains
		//Asset: "ht",
		Asset:   conf.Vip.GetString("gateway." + chain + ".asset"),
		Decimal: 18,
		//The platform used in different chains
		//Platform: "starlabsne3",
		Platform: conf.Vip.GetString("gateway." + chain + ".platform"),
		From:     sysAddr,
		//GasLimit 2000000
		FeeStep: gaslimit,
		//GasPrice 1.2*suggestGasprice, or 150Gwei by default
		FeePrice: feePrice,
		//FeeAsset: "ht",
		FeeAsset: conf.Vip.GetString("gateway." + chain + ".feeAsset"),
		Amount:   amstr,
		ChainId:  conf.Vip.GetString("gateway." + chain + ".chainId"),
	}

	reqDataByte, err := json.Marshal(reqData)
	if err != nil {
		logrus.Errorf("json unmarshal request data error: %v", err)
		return nil, err
	}
	return &UnData{
		FromAddr:     sysAddr,
		Gas:          int(gas * 3 / 2),
		GasPrice:     decimal.NewFromInt(int64(price)),
		Nonce:        nonce,
		UnsignedData: reqDataByte,
	}, nil
}
*/


type SigningReq struct {
	AppId 	string          `json:"appId"`
	SReq    SignReq         `json:"sReq"`
}

type SignReq struct {
	SiReq   SigReqData         `json:"siReq"`
	AuReq   BusData         `json:"auReq"`
}


//SignGatewayEvmChain for EVM compatible chain support
func SignGatewayEvmChain(signReq SignReq, appId string) (encResp Response, err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

	chain := signReq.AuReq.Chain
	switch chain {
	case "bsc":
		chain = "bnb1"
	case "heco":
		chain = "ht2"
	case "eth":
		chain = "eth"
	}
	conf := RemoteConfig(appId)
	//reqData := signReq.SiReq
	reqData := ReqData{
		Asset: conf.Vip.GetString("gateway." + chain + ".asset"),
		Platform: conf.Vip.GetString("gateway." + chain + ".platform"),
		FeeAsset: conf.Vip.GetString("gateway." + chain + ".feeAsset"),
		ChainId: conf.Vip.GetString("gateway." + chain + ".chainId"),
		ToTag: signReq.SiReq.ToTag,
		Decimal: signReq.SiReq.Decimal,
		From: strings.ToLower(signReq.SiReq.From),
		To: strings.ToLower(signReq.SiReq.To),
		FeeStep: signReq.SiReq.FeeStep,
		FeePrice: signReq.SiReq.FeePrice,
		Amount: signReq.SiReq.Amount,
		Nonce: signReq.SiReq.Nonce,
	}

	//the gateway url for signing according to different chains
	Url := conf.Vip.GetString("gateway." + chain + ".url")
	//sysAddr := conf.Vip.GetString("gateway." + chain + ".sysAddr")
	sysAddr := reqData.From

	encPara := &EncParams{
		From: sysAddr,
		To: reqData.To,
		Value: reqData.Amount,
		InputData: reqData.ToTag,
		Chain: chain,
		Quantity: signReq.AuReq.Quantity,
		ToAddress: strings.ToLower(signReq.AuReq.ToAddress),
		ToTag: reqData.ToTag,
		TaskType: signReq.SiReq.TaskType,
	}
	encParaByte, err := json.Marshal(encPara)
	if err != nil {
		return
	}

	//marshal the req data into []byte
	reqDataByte, err := json.Marshal(&reqData)
	if err != nil {
		logrus.Errorf("json unmarshal request data error: %v", err)
		return
	}
	logrus.Infof("The request ++++++++ is %s", string(reqDataByte))

	data := &Payload{
		Addrs:         []string{sysAddr},
		Chain:         conf.Vip.GetString("gateway." + chain + ".payloadChain"),
		Data:          string(reqDataByte),
		EncryptParams: string(encParaByte),
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return
	}

	logrus.Infof("The request body is %s", string(payloadBytes))

	body := bytes.NewReader(payloadBytes)


	req1, err := http.NewRequest("POST", Url, body)
	req1.Header.Set("content-type", "application/json")
	req1.Header.Set("Host", "signer.blockchain.amazonaws.com")
	key := &Key{
		AccessKey: conf.Vip.GetString("gateway." + chain + ".accessKey"),
		SecretKey: conf.Vip.GetString("gateway." + chain + ".secretKey"),
	}

	req1.Host = AwsV4SigHeader
	_, err = SignRequestWithAwsV4UseQueryString(req1, key, "blockchain", "signer")
	//logrus.Infof("the sp is %v", sp)
	resp, err := myclient.Do(req1)
	if err != nil {
		logrus.Errorf("http do request error")
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("response from gateway error")
		return
	}

	//fmt.Println(string(respBody))
	//unmarshal the respBody
	var result Response
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		logrus.Errorf("json unmarshal error")
		return
	}

	//check the signing result is returned with true status
	if !result.Result {
		logrus.Errorf("signing result from gateway is failed")
		return
	}

	//fmt.Println("The encrypted data is:", result.Data.EncryptData)
	encResp = result
	return encResp, nil

}

// Sign ...
func (k *Key) Sign(t time.Time, region, name string) []byte {
	h := ghmac([]byte("AWS4"+k.SecretKey), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(region))
	h = ghmac(h, []byte(name))
	h = ghmac(h, []byte("aws4_request"))
	return h
}
func SignRequestWithAwsV4UseQueryString(req *http.Request, key *Key, region, name string) (sp *SignProcess, err error) {
	date := req.Header.Get(headKeyData)
	t := time.Now().UTC()
	if date != "" {
		t, err = time.Parse(http.TimeFormat, date)
		if err != nil {
			return
		}
	}
	values := req.URL.Query()
	values.Set(headKeyXAmzDate, t.Format(iSO8601BasicFormat))

	//req.Header.Set(headKeyHost, req.Host)

	sp = new(SignProcess)
	sp.Key = key.Sign(t, region, name)

	values.Set(queryKeyAlgorithm, aws4HmacSha256Algorithm)
	values.Set(queryKeyCredential, key.AccessKey+"/"+creds(t, region, name))
	cc := bytes.NewBufferString("")
	writeHeaderList(req, nil, cc, false)
	values.Set(queryKeySignatureHeaders, cc.String())
	req.URL.RawQuery = values.Encode()

	writeStringToSign(t, req, nil, sp, false, region, name)
	values = req.URL.Query()
	values.Set(queryKeySignature, hex.EncodeToString(sp.AllSHA256))
	req.URL.RawQuery = values.Encode()

	return
}

func creds(t time.Time, region, name string) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + region + "/" + name + "/aws4_request"
}

func gsha256(data []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func ghmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

type SignProcess struct {
	Key           []byte
	Body          []byte
	BodySHA256    []byte
	Request       []byte
	RequestSHA256 []byte
	All           []byte
	AllSHA256     []byte
}

func writeHeaderList(r *http.Request, signedHeadersMap map[string]bool, requestData io.Writer, isServer bool) {
	a := make([]string, 0)
	for k := range r.Header {
		if isServer {
			if _, ok := signedHeadersMap[strings.ToLower(k)]; !ok {
				continue
			}
		}
		a = append(a, strings.ToLower(k))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write([]byte{';'})
		}
		_, _ = requestData.Write([]byte(s))
	}
}

func writeStringToSign(
	t time.Time,
	r *http.Request,
	signedHeadersMap map[string]bool,
	sp *SignProcess,
	isServer bool,
	region, name string) {
	lastData := bytes.NewBufferString(aws4HmacSha256Algorithm)
	lastData.Write(lf)

	lastData.Write([]byte(t.Format(iSO8601BasicFormat)))
	lastData.Write(lf)

	lastData.Write([]byte(creds(t, region, name)))
	lastData.Write(lf)

	writeRequest(r, signedHeadersMap, sp, isServer)
	lastData.WriteString(hex.EncodeToString(sp.RequestSHA256))
	// fmt.Fprintf(lastData, "%x", sp.RequestSHA256)

	sp.All = lastData.Bytes()
	sp.AllSHA256 = ghmac(sp.Key, sp.All)
}

func writeRequest(r *http.Request, signedHeadersMap map[string]bool, sp *SignProcess, isServer bool) {
	requestData := bytes.NewBufferString("")
	//content := strings.Split(r.Host, ":")
	r.Header.Set(headKeyHost, "signer.blockchain.amazonaws.com")

	requestData.Write([]byte(r.Method))
	requestData.Write(lf)

	writeURI(r, requestData)
	requestData.Write(lf)

	writeQuery(r, requestData)
	requestData.Write(lf)

	writeHeader(r, signedHeadersMap, requestData, isServer)
	requestData.Write(lf)
	requestData.Write(lf)

	writeHeaderList(r, signedHeadersMap, requestData, isServer)
	requestData.Write(lf)

	writeBody(r, requestData, sp)

	sp.Request = requestData.Bytes()
	sp.RequestSHA256 = gsha256(sp.Request)
}

func writeURI(r *http.Request, requestData io.Writer) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	_, _ = requestData.Write([]byte(path))
}

func writeQuery(r *http.Request, requestData io.Writer) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		if strings.ToLower(k) == queryKeySignature {
			continue
		}
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write([]byte{'&'})
		}
		_, _ = requestData.Write([]byte(s))
	}
}

func writeHeader(r *http.Request, signedHeadersMap map[string]bool, requestData *bytes.Buffer, isServer bool) {
	a := make([]string, 0)
	for k, v := range r.Header {
		if isServer {
			if _, ok := signedHeadersMap[strings.ToLower(k)]; !ok {
				continue
			}
		}
		sort.Strings(v)
		a = append(a, strings.ToLower(k)+":"+strings.Join(v, ","))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write(lf)
		}
		_, _ = requestData.WriteString(s)
	}
}

func writeBody(r *http.Request, requestData io.StringWriter, sp *SignProcess) {
	var b []byte
	// If the payload is empty, use the empty string as the input to the SHA256 function
	// http://docs.amazonwebservices.com/general/latest/gr/sigv4-create-canonical-request.html
	if r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}
	sp.Body = b

	sp.BodySHA256 = gsha256(b)
	_, _ = requestData.WriteString(hex.EncodeToString(sp.BodySHA256))
}

func (p *SignProcess) String() string {
	result := new(strings.Builder)
	fmt.Fprintf(result, "key(hex): %s\n\n", hex.EncodeToString(p.Key))
	fmt.Fprintf(result, "body:\n%s\n", string(p.Body))
	fmt.Fprintf(result, "body sha256: %s\n\n", hex.EncodeToString(p.BodySHA256))
	fmt.Fprintf(result, "request:\n%s\n", string(p.Request))
	fmt.Fprintf(result, "request sha256: %s\n\n", hex.EncodeToString(p.RequestSHA256))
	fmt.Fprintf(result, "all:\n%s\n", string(p.All))
	fmt.Fprintf(result, "all sha256: %s\n", hex.EncodeToString(p.AllSHA256))
	return result.String()
}
