# market-price-query

查詢 MOA（農委會）農產品交易行情與菜價。

## 用途

快速查詢當季蔬果肉品的市場價格、交易量等資訊。支援依地區、時間篩選。

## 安裝

### 二進位發布（推薦）

```bash
mkdir -p ~/.local/bin
curl -sL https://github.com/zea00312-dot/market-price-query/releases/latest/download/market-price-query-linux-arm64 \
  -o ~/.local/bin/market-price-query
chmod +x ~/.local/bin/market-price-query
```

### 本地編譯

需要 Go 1.21+

```bash
git clone https://github.com/zea00312-dot/market-price-query.git
cd market-price-query
go build -o market-price-query .
./market-price-query --help
```

## 用法

### 查詢菜價

```bash
# 查詢高麗菜近 7 天新竹地區價格
market-price-query price 高麗菜

# 指定地區（台北）
market-price-query price 番茄 --region 台北

# 查詢 14 天資料
market-price-query price 蕃茄 --days 14

# JSON 格式（給 agent 用）
market-price-query price 高麗菜 --json
```

### 列出可查詢品項

```bash
# 列出所有品項
market-price-query list

# JSON 格式
market-price-query list --json
```

### 版本

```bash
market-price-query version
```

## 引數說明

| 引數 | 預設值 | 說明 |
|------|--------|------|
| `--region, -r` | 新竹 | 查詢地區 |
| `--days, -d` | 7 | 查詢天數 |
| `--json` | false | JSON 輸出格式 |

## 資料來源

[MOA 交易行情 API](https://data.moa.gov.tw/api.aspx#operations-tag-%E4%BA%A4%E6%98%93%E8%A1%8C%E6%83%85)

## 許可

MIT
