package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"go-btc/helper"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/tyler-smith/go-bip39"
)

// 使用 BIP-49 路径生成比特币地址
func main() {
	// 生成助记词
	mnemonic, err := helper.GetMnemonicFromENV()
	if err != nil {
		log.Fatal(err)
	}

	// 生成种子
	seed := bip39.NewSeed(mnemonic, "")

	// 基于种子生成主私钥
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatal(err)
	}

	// 使用 BIP-49 路径 m/49'/0'/0'/0/0
	purpose, err := masterKey.Derive(49 + hdkeychain.HardenedKeyStart) // 49'
	if err != nil {
		log.Fatal(err)
	}

	coinType, err := purpose.Derive(0 + hdkeychain.HardenedKeyStart) // 0' for Bitcoin
	if err != nil {
		log.Fatal(err)
	}

	account, err := coinType.Derive(0 + hdkeychain.HardenedKeyStart) // 0' for default account
	if err != nil {
		log.Fatal(err)
	}

	external, err := account.Derive(0) // 0 for external chain (receiving addresses)
	if err != nil {
		log.Fatal(err)
	}

	addressIndex, err := external.Derive(0) // Address index 0
	if err != nil {
		log.Fatal(err)
	}

	// 导出私钥
	privateKey, err := addressIndex.ECPrivKey()
	if err != nil {
		log.Fatal(err)
	}
	// 将私钥转换为 WIF 格式
	wif, err := btcutil.NewWIF(privateKey, &chaincfg.MainNetParams, true)
	if err != nil {
		log.Fatal(err)
	}

	// 导出公钥
	publicKey := privateKey.PubKey()

	// 使用哈希160 (RIPEMD160(SHA256(pubKey))) 生成公钥哈希
	pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())

	// 生成 P2SH 地址（包含 P2WPKH）
	// 生成对应的锁定脚本：OP_HASH160 <pubKeyHash> OP_EQUAL
	redeemScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(pubKeyHash).AddOp(txscript.OP_EQUAL).Script()
	if err != nil {
		log.Fatal(err)
	}

	// 生成 P2SH 地址
	p2shAddress, err := btcutil.NewAddressScriptHash(redeemScript, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatal(err)
	}

	// 输出助记词、私钥和公钥
	fmt.Println("助记词:", mnemonic)
	fmt.Printf("私钥 (WIF): %s\n", wif.String())
	fmt.Printf("公钥 (压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeCompressed()))
	fmt.Printf("公钥 (非压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeUncompressed()))

	// 输出 P2SH 地址
	fmt.Println("P2SH 地址:", p2shAddress.EncodeAddress())
}
