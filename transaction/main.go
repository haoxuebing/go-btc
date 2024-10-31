package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-btc/helper"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var (
	cfg                      = &chaincfg.TestNet3Params
	receiveTaprootAddr       = "tb1pcvwe95ec64urxykp2nfdnvxftfk0rvvw0w77u4mauv2355gxrf4qg0r5xj"
	outputAmount       int64 = 1000
	feeRate                  = HourFee
)

// FeeRateType 定义费率类型
type FeeRateType string

const (
	FastestFee  FeeRateType = "fastest"
	HalfHourFee FeeRateType = "halfHour"
	HourFee     FeeRateType = "hour"
	EconomyFee  FeeRateType = "economy"
	MinimumFee  FeeRateType = "minimum"
)

// UTXO represents a single UTXO retrieved from Mempool API
type UTXO struct {
	TxID     string `json:"txid"`
	Vout     uint32 `json:"vout"`
	Amount   int64  `json:"value"`
	PkScript string `json:"scriptPubKey"`
}

func main() {
	mnemonic, err := helper.GetMnemonicFromENV()
	if err != nil {
		log.Fatalf("获取助记词失败: %v", err)
	}

	wif, _, taprootAddr, err := helper.GenerateTaprootAddress(mnemonic, cfg, 0)
	if err != nil {
		log.Fatalf("生成 BIP-84 地址失败: %v", err)
	}

	// 获取所有UTXOs
	utxos, err := getUTXOsFromAPI(taprootAddr.String())
	if err != nil {
		log.Fatalf("获取UTXOs失败: %v", err)
	}

	// 获取动态费率
	feeRate, err := getFeeRate(feeRate)
	if err != nil {
		log.Fatalf("获取动态费率失败: %v", err)
	}

	// 创建交易
	tx, fetcher := createTransaction(utxos, receiveTaprootAddr, taprootAddr.String(), feeRate)

	// 添加见证
	err = addWitnesses(tx, wif, fetcher)
	if err != nil {
		log.Fatalf("添加见证失败: %v", err)
	}

	// 打印交易详情
	printTransactionDetails(tx, fetcher)

	// 序列化交易
	finalRawTx, err := serializeTransaction(tx)
	if err != nil {
		log.Fatalf("序列化交易失败: %v", err)
	}
	fmt.Println("Signed Transaction: ", finalRawTx)

	// 广播交易
	// txHash, err := broadcastTransaction(tx)
	// if err != nil {
	// 	log.Fatalf("广播交易失败: %v", err)
	// }
	// fmt.Println("Transaction Hash: ", txHash.String())
}

// getFeeRate 获取指定类型的费率
func getFeeRate(feeType FeeRateType) (int64, error) {
	resp, err := http.Get("https://mempool.space/testnet/api/v1/fees/recommended")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var feeData struct {
		FastestFee  int64 `json:"fastestFee"`
		HalfHourFee int64 `json:"halfHourFee"`
		HourFee     int64 `json:"hourFee"`
		EconomyFee  int64 `json:"economyFee"`
		MinimumFee  int64 `json:"minimumFee"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&feeData); err != nil {
		return 0, err
	}

	switch feeType {
	case FastestFee:
		return feeData.FastestFee, nil
	case HalfHourFee:
		return feeData.HalfHourFee, nil
	case HourFee:
		return feeData.HourFee, nil
	case EconomyFee:
		return feeData.EconomyFee, nil
	case MinimumFee:
		return feeData.MinimumFee, nil
	default:
		return feeData.HourFee, nil
	}
}

// getUTXOsFromAPI 从 Mempool API 获取 UTXOs
func getUTXOsFromAPI(address string) ([]UTXO, error) {
	resp, err := http.Get(fmt.Sprintf("https://mempool.space/testnet/api/address/%s/utxo", address))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var utxos []UTXO
	if err := json.NewDecoder(resp.Body).Decode(&utxos); err != nil {
		return nil, err
	}

	// 获取 pkScript 并填充到 utxos 中
	pkScript, _ := decodeTaprootAddress(address, cfg)

	// 打印获取的 UTXO，以检查 pkScript
	for i := range utxos {
		utxos[i].PkScript = hex.EncodeToString(pkScript)
		// utxos[i].PkScript, _ = getPKScriptFromAPI(utxo.TxID, utxo.Vout)
	}

	return utxos, nil
}

// getPKScriptFromAPI 从 Mempool API 获取 pkScript
func getPKScriptFromAPI(txId string, vout uint32) (string, error) {
	url := fmt.Sprintf("https://mempool.space/testnet/api/tx/%s", txId)
	txResp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer txResp.Body.Close()

	var txDetails struct {
		Vout []struct {
			Value        float64 `json:"value"`
			N            int     `json:"n"`
			ScriptPubKey string  `json:"scriptPubKey"`
		} `json:"vout"`
	}
	if err := json.NewDecoder(txResp.Body).Decode(&txDetails); err != nil {
		return "", err
	}

	return txDetails.Vout[vout].ScriptPubKey, nil
}

// createTransaction 创建交易
func createTransaction(utxos []UTXO, receiveAddr, changeAddr string, feeRate int64) (*wire.MsgTx, *txscript.MultiPrevOutFetcher) {
	tx := wire.NewMsgTx(wire.TxVersion)
	fetcher := txscript.NewMultiPrevOutFetcher(nil)

	var totalInputAmount int64
	for _, utxo := range utxos {
		txHash, _ := chainhash.NewHashFromStr(utxo.TxID)
		point := wire.NewOutPoint(txHash, utxo.Vout)
		tx.AddTxIn(wire.NewTxIn(point, nil, nil))
		totalInputAmount += utxo.Amount
		pkScript, _ := hex.DecodeString(utxo.PkScript)
		fetcher.AddPrevOut(*point, wire.NewTxOut(utxo.Amount, pkScript))

		// 重新计算费用
		estimatedTxSize := estimateTxSize(len(tx.TxIn), 2) // 1个输出是接收地址，1个是找零地址
		fee := int64(estimatedTxSize) * feeRate

		if totalInputAmount >= outputAmount+fee {
			break
		}
	}

	receiveByteAddr, _ := decodeTaprootAddress(receiveAddr, cfg)
	tx.AddTxOut(wire.NewTxOut(outputAmount, receiveByteAddr))

	estimatedTxSize := estimateTxSize(len(tx.TxIn), 2)
	fee := int64(estimatedTxSize) * feeRate

	changeAmount := totalInputAmount - outputAmount - fee
	if changeAmount > 0 {
		changeByteAddr, _ := decodeTaprootAddress(changeAddr, cfg)
		tx.AddTxOut(wire.NewTxOut(changeAmount, changeByteAddr))
	}

	return tx, fetcher
}

// estimateTxSize 估算交易大小
func estimateTxSize(numInputs, numOutputs int) int {
	return numInputs*180 + numOutputs*34 + 10 // 估算交易大小
}

// addWitnesses 添加见证
func addWitnesses(tx *wire.MsgTx, wif *btcutil.WIF, fetcher *txscript.MultiPrevOutFetcher) error {
	sigHashes := txscript.NewTxSigHashes(tx, fetcher)

	for i, txIn := range tx.TxIn {
		prevOutput := fetcher.FetchPrevOutput(txIn.PreviousOutPoint)
		if prevOutput == nil {
			log.Fatalf("无法获取前一笔输出: %v", txIn.PreviousOutPoint)
		}
		// 生成见证脚本
		witness, err := txscript.TaprootWitnessSignature(
			tx,
			sigHashes,
			i,
			prevOutput.Value,
			prevOutput.PkScript,
			txscript.SigHashDefault,
			wif.PrivKey,
		)
		if err != nil {
			return err
		}
		tx.TxIn[i].Witness = witness
	}
	return nil
}

// serializeTransaction 序列化交易
func serializeTransaction(tx *wire.MsgTx) (string, error) {
	var signedTx bytes.Buffer
	if err := tx.Serialize(&signedTx); err != nil {
		return "", err
	}
	return hex.EncodeToString(signedTx.Bytes()), nil
}

// broadcastTransaction 广播交易
func broadcastTransaction(tx *wire.MsgTx) (*chainhash.Hash, error) {
	url := os.Getenv("RPC_URL")
	user := "user"
	pass := "pass"

	client, err := helper.NewClient(url, user, pass)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	return client.SendRawTransaction(tx, false)
}

// decodeTaprootAddress 解码 Taproot 地址
func decodeTaprootAddress(strAddr string, cfg *chaincfg.Params) ([]byte, error) {
	taprootAddr, err := btcutil.DecodeAddress(strAddr, cfg)
	if err != nil {
		return nil, err
	}
	return txscript.PayToAddrScript(taprootAddr)
}

// printTransactionDetails 打印交易详细信息
func printTransactionDetails(tx *wire.MsgTx, fetcher *txscript.MultiPrevOutFetcher) {
	var totalIn, totalOut int64

	fmt.Println("交易详情:")
	fmt.Println("输入:")
	for i, in := range tx.TxIn {
		prevOut := fetcher.FetchPrevOutput(in.PreviousOutPoint)
		if prevOut == nil {
			log.Fatal("无法获取输入UTXO信息")
		}
		totalIn += prevOut.Value
		fmt.Printf("  输入 %d: %s:%d, 金额: %d\n", i, in.PreviousOutPoint.Hash, in.PreviousOutPoint.Index, prevOut.Value)
	}

	fmt.Println("输出:")
	for i, out := range tx.TxOut {
		totalOut += out.Value
		fmt.Printf("  输出 %d: 金额: %d\n", i, out.Value)
	}

	actualFee := totalIn - totalOut
	fmt.Printf("总输入: %d, 总输出: %d, 手续费: %d\n", totalIn, totalOut, actualFee)
}
