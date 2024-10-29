package helper

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/joho/godotenv"

	"github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic 生成助记词
func GenerateMnemonic() (string, error) {
	// 生成助记词
	entropy, err := bip39.NewEntropy(128) // 128-bit entropy, 可以使用 256-bit
	if err != nil {
		return "", fmt.Errorf("生成熵失败: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("生成助记词失败: %w", err)
	}

	return mnemonic, nil
}

// GetMnemonicFromENV 从环境变量中获取助记词
func GetMnemonicFromENV() (string, error) {
	// 加载 .env 文件
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mnemonic := os.Getenv("MNEMONIC")
	if mnemonic == "" {
		return "", fmt.Errorf("助记词为空")
	}
	return mnemonic, nil
}

func GenerateTaprootAddress(mnemonic string, netParams *chaincfg.Params, addressIndex uint32) (*btcutil.WIF, *btcec.PublicKey, btcutil.Address, error) {
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, err := hdkeychain.NewMaster(seed, netParams)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建主私钥失败: %w", err)
	}

	path := []uint32{
		86 + hdkeychain.HardenedKeyStart, // purpose
		0 + hdkeychain.HardenedKeyStart,  // coin type
		0 + hdkeychain.HardenedKeyStart,  // account
		0,                                // external chain
		addressIndex,                     // address index
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

	wif, err := btcutil.NewWIF(privateKey, netParams, true)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建WIF失败: %w", err)
	}

	publicKey := privateKey.PubKey()
	// pubKeyHash := btcutil.Hash160(publicKey.SerializeCompressed())

	taprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(
		txscript.ComputeTaprootKeyNoScript(wif.PrivKey.PubKey())),
		netParams,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建taprootAddress地址失败: %w", err)
	}

	return wif, publicKey, taprootAddress, nil
}

func PrintResults(mnemonic string, wif *btcutil.WIF, publicKey *btcec.PublicKey, bech32Address btcutil.Address) {
	fmt.Println("助记词:", mnemonic)
	fmt.Printf("私钥 (WIF): %s\n", wif.String())
	fmt.Printf("公钥 (压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeCompressed()))
	fmt.Printf("公钥 (非压缩格式): %s\n", hex.EncodeToString(publicKey.SerializeUncompressed()))
	fmt.Println("地址:", bech32Address.EncodeAddress())
}
