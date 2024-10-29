package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

func main() {
	connCfg := &rpcclient.ConnConfig{
		Host:         "newest-black-needle.btc-testnet.quiknode.pro/b33c9ebcd8dc02361c951ebb88f1accc123a262e",
		User:         "user", // 不填会报错
		Pass:         "pass", // 不填会报错
		HTTPPostMode: true,
		DisableTLS:   false,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("连接到 RPC 节点失败: %v", err)
	}
	defer client.Shutdown()

	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatalf("获取区块高度失败: %v", err)
	}
	log.Printf("当前区块高度: %d", blockCount)

	txid := "c1c9e592f3c32a08302e4a99238c53987b9e842965947192a3db758610043e3e"
	err = GetRawTransaction(client, txid)
	if err != nil {
		log.Fatalf("获取交易详情失败: %v", err)
	}
}

// 获取交易详情
func GetRawTransaction(client *rpcclient.Client, txid string) error {
	hash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return err
	}

	// 获取交易详情
	rawTx, err := client.GetRawTransaction(hash)
	if err != nil {
		return err
	}

	// 遍历每个输出并打印 ScriptPubKey
	for _, output := range rawTx.MsgTx().TxOut {
		fmt.Printf("Value: %d\n", output.Value)
		fmt.Printf("ScriptPubKey: %s\n", hex.EncodeToString(output.PkScript))
	}

	return nil
}
