package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"go-btc/helper"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/tyler-smith/go-bip39"
)

var NetParams = &chaincfg.TestNet3Params

func main() {
	mnemonic, err := helper.GetMnemonicFromENV()
	if err != nil {
		log.Fatalf("获取助记词失败: %v", err)
	}

	// 生成比特币地址
	privateKey, publicKey, addr, err := generateBitcoinAddress(mnemonic)
	if err != nil {
		log.Fatalf("生成比特币地址失败: %v", err)
	}

	// 输出结果
	printResults(mnemonic, privateKey, publicKey, addr)
}

func generateBitcoinAddress(mnemonic string) (*btcutil.WIF, *btcec.PublicKey, *btcutil.AddressPubKey, error) {
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, err := hdkeychain.NewMaster(seed, NetParams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建主私钥失败: %w", err)
	}

	path := []uint32{
		44 + hdkeychain.HardenedKeyStart, // purpose
		0 + hdkeychain.HardenedKeyStart,  // coin type
		0 + hdkeychain.HardenedKeyStart,  // account
		0,                                // external chain
		1,                                // address index
	}

	key := masterKey
	for _, childNum := range path {
		key, err = key.Derive(childNum)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("派生密钥失败: %w", err)
		}
	}

	privateKey, err := key.ECPrivKey()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取私钥失败: %w", err)
	}

	wif, err := btcutil.NewWIF(privateKey, NetParams, true)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建WIF失败: %w", err)
	}

	publicKey := privateKey.PubKey()

	addr, err := btcutil.NewAddressPubKey(publicKey.SerializeCompressed(), NetParams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建地址失败: %w", err)
	}

	return wif, publicKey, addr, nil
}

func printResults(mnemonic string, wif *btcutil.WIF, publicKey *btcec.PublicKey, addr *btcutil.AddressPubKey) {
	fmt.Println("助记词:", mnemonic)
	fmt.Printf("私钥 (WIF): %s\n", wif.String())
	fmt.Printf("公钥 (压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeCompressed()))
	fmt.Printf("公钥 (非压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeUncompressed()))
	fmt.Println("Legacy 地址:", addr.EncodeAddress())
}
