---
name: market-price-query
description: "查詢 MOA 交易行情菜價，當使用者問農產品當季價格、品項行情、地區菜價時觸發"
---

# market-price-query — MOA 農產品菜價查詢

## 前置確認：確保 binary 存在

```bash
ls ~/.local/bin/market-price-query 2>/dev/null || (
  mkdir -p ~/.local/bin &&
  curl -sL https://github.com/zea00312-dot/market-price-query/releases/latest/download/market-price-query-linux-arm64 -o ~/.local/bin/market-price-query &&
  chmod +x ~/.local/bin/market-price-query
)
```

## 觸發情境

| 情境 | 訊號 | 動作 |
|------|------|------|
| 查詢單一品項菜價 | 「高麗菜多少錢」「番茄漲了嗎」「最近蔥價格」 | `~/.local/bin/market-price-query price <品項> --region <地區> --days <天數> --json` |
| 查詢可用品項 | 「有什麼菜可以查」「支援哪些品項」 | `~/.local/bin/market-price-query list --json` |
| 查詢特定地區價格 | 「台北的番茄多少」「新竹蔬菜行情」 | `~/.local/bin/market-price-query price <品項> --region 台北 --json` |
| 查詢時間範圍 | 「最近兩週高麗菜走勢」 | `~/.local/bin/market-price-query price <品項> --days 14 --json` |

## 指令

```bash
# 人類閱讀格式
~/.local/bin/market-price-query price <品項> [--region <地區>] [--days <天數>]
~/.local/bin/market-price-query list
~/.local/bin/market-price-query version

# JSON 格式（agent 用）
~/.local/bin/market-price-query price <品項> [--region <地區>] [--days <天數>] --json
~/.local/bin/market-price-query list --json
```

## 使用範例

### 範例 1：查詢高麗菜近 7 天新竹價格

**輸入：** 「高麗菜現在新竹多少錢」

**命令：**
```bash
~/.local/bin/market-price-query price 高麗菜 --region 新竹 --days 7 --json
```

**預期輸出（JSON）：**
```json
{
  "ok": true,
  "product": "高麗菜",
  "region": "新竹",
  "count": 7,
  "data": [
    {
      "product_name": "高麗菜",
      "market_name": "新竹",
      "trans_date": "2026-05-10",
      "avg_price": "25.5",
      "low_price": "24.0",
      "high_price": "27.0",
      "volume": "1500"
    },
    ...
  ]
}
```

**回應給使用者：** 整理資料成易讀格式，例如「高麗菜最近一週新竹平均 25.5 元，最便宜 24 元，最貴 27 元」

### 範例 2：列出可查詢品項

**輸入：** 「支援查哪些菜」

**命令：**
```bash
~/.local/bin/market-price-query list --json
```

**預期輸出：**
```json
{
  "ok": true,
  "count": 42,
  "data": ["高麗菜", "番茄", "蔥", ...]
}
```

## 行為規則

### 該做的

- 使用者問菜價時，**先確認品項是否存在**（可用 `list` 查詢）
- **預設地區新竹**，除非使用者指定其他地區
- **預設時間範圍 7 天**，除非使用者要求更長
- **整理 JSON 結果成人類可讀格式**（表格、趨勢說明等）
- 當查詢無結果時，明確告知「找不到該品項」或「該時間範圍無資料」

### 不該做的

- 不要在沒有確認地區的情況下假設用戶所在位置（總是明確顯示查詢結果的地區）
- 不要自動展開過去 30 天以上的資料，除非明確要求
- 不要回傳原始 JSON 給使用者（解析後轉成易讀表格或敘述）

## 相關 skill

- [git.md](git.md) — 若後續要 commit 工具更新或修復
