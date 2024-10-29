# Go-BTC

## 前置知识：UTXO、交易结构与脚本语言

- 测试网交易池可视化：https://mempool.space/testnet

- 测试代币水龙头：https://bitcoinfaucet.uo1.net/

- 测试私钥生成（主网尽量不要使用）：https://iancoleman.io/bip39/

golang 下的 Bitcoin 工具库为：github.com/btcsuite/btcd

## 使用助记词派生比特币地址

**比特币有三种不同的地址格式：**

1. Legacy 地址 (P2PKH)：以 1 开头，遵循传统的 BIP-44 派生。 [bip-44代码](account/bip-44/main.go)。

2. P2SH 地址：以 3 开头，遵循 BIP-49，用于 SegWit 的兼容模式。 [bip-49代码](account/bip-49/main.go)

3. Bech32 地址 (Native SegWit, P2WPKH)：以 bc1 开头，遵循 BIP-84，是更现代化的 SegWit 地址格式。 [bip-84代码](account/bip-84/main.go)。

example:

``` txt
# Legacy 地址，使用这种地址的钱包：Leather
1CDqGVeqD5mXt3eBedHNC6KJ3xVePkZPzb

# P2SH 地址
3GsQkzzwdi5uaEuEJGt1UofPLk7KEvxn2m

# Bech32 地址，使用这种地址的钱包：CoinWallet
bc1qz66uxud3kv3s79ddpnkyj3s2spc2flqudk4www
```

## 创建一笔交易

- 测试网下 [创建一笔交易](transaction/main.go)
- 使用 https://mempool.space/testnet/tx/push 将交易推送到 mempool 中
