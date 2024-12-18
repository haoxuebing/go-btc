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
	"github.com/btcsuite/btcd/txscript"
	"github.com/tyler-smith/go-bip39"
)

func main() {
	mnemonic, err := helper.GetMnemonicFromENV()
	if err != nil {
		log.Fatalf("获取助记词失败: %v", err)
	}

	// 生成比特币地址
	wif, publicKey, p2shAddress, err := generateBIP49Address(mnemonic)
	if err != nil {
		log.Fatalf("生成BIP-49地址失败: %v", err)
	}

	// 输出结果
	printResults(mnemonic, wif, publicKey, p2shAddress)
}

func generateBIP49Address(mnemonic string) (*btcutil.WIF, *btcec.PublicKey, *btcutil.AddressScriptHash, error) {
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建主私钥失败: %w", err)
	}

	path := []uint32{
		49 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart,
		0 + hdkeychain.HardenedKeyStart,
		0,
		0,
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

	wif, err := btcutil.NewWIF(privateKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建WIF失败: %w", err)
	}

	publicKey := privateKey.PubKey()
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())

	redeemScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).
		AddData(pubKeyHash).
		Script()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建赎回脚本失败: %w", err)
	}

	p2shAddress, err := btcutil.NewAddressScriptHash(redeemScript, &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建P2SH地址失败: %w", err)
	}

	return wif, publicKey, p2shAddress, nil
}

func printResults(mnemonic string, wif *btcutil.WIF, publicKey *btcec.PublicKey, p2shAddress *btcutil.AddressScriptHash) {
	fmt.Println("助记词:", mnemonic)
	fmt.Printf("私钥 (WIF): %s\n", wif.String())
	fmt.Printf("公钥 (压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeCompressed()))
	fmt.Printf("公钥 (非压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeUncompressed()))
	fmt.Println("P2SH 地址:", p2shAddress.EncodeAddress())
}
