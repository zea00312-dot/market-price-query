package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const moaAPIBase = "https://data.moa.gov.tw/api.aspx"

type PriceRecord struct {
	ProductName string `json:"product_name"`
	MarketName  string `json:"market_name"`
	TransDate   string `json:"trans_date"`
	AvgPrice    string `json:"avg_price"`
	LowPrice    string `json:"low_price"`
	HighPrice   string `json:"high_price"`
	Volume      string `json:"volume"`
}

type APIResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Result  []PriceRecord  `json:"result"`
}

var (
	region    string
	days      int
	startDate string
	endDate   string
)

func queryMOAAPI(product string) ([]PriceRecord, error) {
	params := url.Values{}
	params.Set("$top", "1000")
	params.Set("$format", "json")
	params.Set("$filter", fmt.Sprintf("ProductName eq '%s'", product))

	if region != "" && region != "新竹" {
		params.Set("$filter", fmt.Sprintf("ProductName eq '%s' and MarketName eq '%s'", product, region))
	}

	if startDate != "" {
		params.Set("$filter", fmt.Sprintf("%s and TransDate ge '%s'", params.Get("$filter"), startDate))
	}

	if endDate != "" {
		params.Set("$filter", fmt.Sprintf("%s and TransDate le '%s'", params.Get("$filter"), endDate))
	}

	query := fmt.Sprintf("%s?%s", moaAPIBase, params.Encode())

	resp, err := http.Get(query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []PriceRecord
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

var rootCmd = &cobra.Command{
	Use:   "market-price-query",
	Short: "查詢 MOA 農產品交易行情",
}

var priceCmd = &cobra.Command{
	Use:   "price <品項>",
	Short: "查詢品項近期菜價",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		product := args[0]

		if days > 0 {
			end := time.Now()
			start := end.AddDate(0, 0, -days)
			startDate = start.Format("2006-01-02")
			endDate = end.Format("2006-01-02")
		}

		records, err := queryMOAAPI(product)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}

		if asJSON, _ := cmd.Flags().GetBool("json"); asJSON {
			json.NewEncoder(os.Stdout).Encode(map[string]any{
				"ok":      true,
				"product": product,
				"region":  region,
				"count":   len(records),
				"data":    records,
			})
			return nil
		}

		if len(records) == 0 {
			fmt.Printf("找不到 %s 的行情資料\n", product)
			return nil
		}

		fmt.Printf("品項: %s | 地區: %s | 筆數: %d\n", product, region, len(records))
		fmt.Println("日期          | 平均價  | 最低價  | 最高價  | 交易量")
		fmt.Println("---------- | ------- | ------- | ------- | -------")
		for _, r := range records {
			fmt.Printf("%s | %7s | %7s | %7s | %7s\n", r.TransDate, r.AvgPrice, r.LowPrice, r.HighPrice, r.Volume)
		}

		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可查詢的品項",
	RunE: func(cmd *cobra.Command, args []string) error {
		params := url.Values{}
		params.Set("$top", "1000")
		params.Set("$format", "json")
		params.Set("$select", "ProductName")
		params.Set("$skip", "0")

		query := fmt.Sprintf("%s?%s", moaAPIBase, params.Encode())

		resp, err := http.Get(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}

		var records []map[string]interface{}
		if err := json.Unmarshal(body, &records); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}

		seen := map[string]bool{}
		if asJSON, _ := cmd.Flags().GetBool("json"); asJSON {
			products := []string{}
			for _, r := range records {
				if name, ok := r["ProductName"].(string); ok && !seen[name] {
					products = append(products, name)
					seen[name] = true
				}
			}
			json.NewEncoder(os.Stdout).Encode(map[string]any{
				"ok":    true,
				"count": len(products),
				"data":  products,
			})
		} else {
			fmt.Println("可查詢品項:")
			count := 0
			for _, r := range records {
				if name, ok := r["ProductName"].(string); ok && !seen[name] {
					fmt.Printf("  - %s\n", name)
					seen[name] = true
					count++
				}
			}
			fmt.Printf("\n共 %d 項\n", count)
		}

		return nil
	},
}

func init() {
	priceCmd.Flags().StringVarP(&region, "region", "r", "新竹", "查詢地區（預設：新竹）")
	priceCmd.Flags().IntVarP(&days, "days", "d", 7, "查詢天數（預設：7 天）")
	priceCmd.Flags().BoolP("json", "", false, "JSON 輸出")

	listCmd.Flags().BoolP("json", "", false, "JSON 輸出")

	rootCmd.AddCommand(priceCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "版本資訊",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("market-price-query v0.1.0")
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
