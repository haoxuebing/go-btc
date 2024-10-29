package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"go-btc/helper"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var (
	cfg                = &chaincfg.TestNet3Params                                           // 测试网参数
	preTxid            = "384f7043cda6eadca5b781689f125778d954cafee9f403a49f96380fe4a4600b" // 前一笔交易的交易ID
	receiveTaprootAddr = "tb1pcvwe95ec64urxykp2nfdnvxftfk0rvvw0w77u4mauv2355gxrf4qg0r5xj"   // 接收地址
)

func main() {
	// 获取助记词
	mnemonic, err := helper.GetMnemonicFromENV()
	if err != nil {
		log.Fatalf("获取助记词失败: %v", err)
	}

	// 生成比特币地址
	wif, _, taprootAddr, err := helper.GenerateTaprootAddress(mnemonic, cfg, 0)
	if err != nil {
		log.Fatalf("生成BIP-84地址失败: %v", err)
	}

	// fmt.Printf("Generated WIF Key: %s \n", wif.String())
	// fmt.Printf("Generated Public Key: %x \n", publicKey.SerializeCompressed())
	// fmt.Printf("Generated taprootAddr : %s \n", taprootAddr.String())

	receiveByteAddr, err := DecodeTaprootAddress(receiveTaprootAddr, cfg)
	if err != nil {
		log.Fatalf("生成BIP-84地址失败: %v", err)
	}
	// fmt.Printf("Generated receiveByteAddr : %x \n", receiveByteAddr)

	sendByteAddr, _ := DecodeTaprootAddress(taprootAddr.String(), cfg)
	sendStrAddr := hex.EncodeToString(sendByteAddr)
	// fmt.Println("sendByteAddr: ", sendStrAddr)
	// 获取未花费的交易输出
	point, fetcher := GetUnspent(taprootAddr.String(), preTxid, sendStrAddr, 1000000)

	// 默认的 version = 1
	tx := wire.NewMsgTx(wire.TxVersion)

	// 以前一笔交易的输出点作为输入
	in := wire.NewTxIn(point, nil, nil)
	tx.AddTxIn(in)

	// 新建输出，支付到指定地址并填充转移多少
	out := wire.NewTxOut(int64(530000), receiveByteAddr)
	tx.AddTxOut(out)

	// 获取前一笔交易
	prevOutput := fetcher.FetchPrevOutput(in.PreviousOutPoint)

	// 使用私钥生成见证脚本
	witness, _ := txscript.TaprootWitnessSignature(tx,
		txscript.NewTxSigHashes(tx, fetcher), 0, prevOutput.Value,
		prevOutput.PkScript, txscript.SigHashDefault, wif.PrivKey)

	// 填充输入的见证脚本
	tx.TxIn[0].Witness = witness

	// 将完成签名的交易转为 hex 形式并输出
	var signedTx bytes.Buffer
	tx.Serialize(&signedTx)
	finalRawTx := hex.EncodeToString(signedTx.Bytes())
	fmt.Println("Signed Transaction: ", finalRawTx)
	// 手动广播交易：通过 https://mempool.space/testnet/tx/push 将交易推送到mempool中

	// 广播交易
	txHash, err := BroadcastTx(tx)
	if err != nil {
		log.Fatalf("广播交易失败: %v", err)
	}

	fmt.Println("Transaction Hash: ", txHash.String())
}

// 广播交易
func BroadcastTx(tx *wire.MsgTx) (*chainhash.Hash, error) {
	url := "newest-black-needle.btc-testnet.quiknode.pro/b33c9ebcd8dc02361c951ebb88f1accc123a262e"
	user := "user"
	pass := "pass"

	client, err := helper.NewClient(url, user, pass)
	if err != nil {
		return nil, err
	}

	return client.SendRawTransaction(tx, false)
}

// 获取未花费的交易输出
func GetUnspent(address, txid, scriptPubKey string, amount int64) (*wire.OutPoint, *txscript.MultiPrevOutFetcher) {
	// 交易的哈希值，并且要指定输出位置
	txHash, _ := chainhash.NewHashFromStr(txid)
	point := wire.NewOutPoint(txHash, uint32(0))

	// 交易的锁定脚本，对应的是 ScriptPubKey 字段
	script, _ := hex.DecodeString(scriptPubKey)
	output := wire.NewTxOut(amount, script)
	fetcher := txscript.NewMultiPrevOutFetcher(nil)
	fetcher.AddPrevOut(*point, output)

	return point, fetcher
}

// 解码 taproot 地址
func DecodeTaprootAddress(strAddr string, cfg *chaincfg.Params) ([]byte, error) {
	taprootAddr, err := btcutil.DecodeAddress(strAddr, cfg)
	if err != nil {
		return nil, err
	}

	byteAddr, err := txscript.PayToAddrScript(taprootAddr)
	if err != nil {
		return nil, err
	}
	return byteAddr, nil
}
