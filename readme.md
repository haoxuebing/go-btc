# Go-BTC

## 使用助记词派生比特币地址

**比特币有三种不同的地址格式：**

•Legacy 地址 (P2PKH)：以 1 开头，遵循传统的 BIP-44 派生。 [bip-44代码](account/bip-44/main.go)。eg：**Leather**

•P2SH 地址：以 3 开头，遵循 BIP-49，用于 SegWit 的兼容模式。 [bip-49代码](account/bip-49/main.go)

•Bech32 地址 (Native SegWit, P2WPKH)：以 bc1 开头，遵循 BIP-84，是更现代化的 SegWit 地址格式。 [bip-84代码](account/bip-84/main.go)。eg：**CoinWallet**

example:

``` txt
# Legacy 地址
1CDqGVeqD5mXt3eBedHNC6KJ3xVePkZPzb

# P2SH 地址
3GsQkzzwdi5uaEuEJGt1UofPLk7KEvxn2m

# Bech32 地址
bc1qz66uxud3kv3s79ddpnkyj3s2spc2flqudk4www
```