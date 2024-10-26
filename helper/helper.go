package helper

import (
	"fmt"
	"log"
	"os"

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
