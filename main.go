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

const moaAPIBase = "https://data.moa.gov.tw"

type PriceRecord struct {
	CropCode      string `json:"CropCode"`
	CropName      string `json:"CropName"`
	MarketName    string `json:"MarketName"`
	TransDate     string `json:"TransDate"`
	AvgPrice      string `json:"Avg_Price"`
	UpperPrice    string `json:"Upper_Price"`
	LowerPrice    string `json:"Lower_Price"`
	MiddlePrice   string `json:"Middle_Price"`
	TransQuantity string `json:"Trans_Quantity"`
}

var (
	region    string
	days      int
	startDate string
	endDate   string
)

func dateToROC(t time.Time) string {
	rocYear := t.Year() - 1911
	return fmt.Sprintf("%03d.%02d.%02d", rocYear, t.Month(), t.Day())
}

func queryMOAAPI(product string) ([]PriceRecord, error) {
	params := url.Values{}
	params.Set("CropName", product)
	params.Set("MarketName", region)

	if startDate == "" && endDate == "" && days > 0 {
		end := time.Now()
		start := end.AddDate(0, 0, -days)
		startDate = dateToROC(start)
		endDate = dateToROC(end)
	}

	if startDate != "" {
		params.Set("Start_time", startDate)
	}
	if endDate != "" {
		params.Set("End_time", endDate)
	}

	query := fmt.Sprintf("%s/AgriProductsTransType/?%s", moaAPIBase, params.Encode())

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
		fmt.Println("日期       | 平均價 | 中價  | 高價  | 低價  | 交易量(kg)")
		fmt.Println("---------|--------|-------|-------|-------|----------")
		for _, r := range records {
			fmt.Printf("%s | %6s | %5s | %5s | %5s | %9s\n", r.TransDate, r.AvgPrice, r.MiddlePrice, r.UpperPrice, r.LowerPrice, r.TransQuantity)
		}

		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出可查詢的品項",
	RunE: func(cmd *cobra.Command, args []string) error {
		end := time.Now()
		start := end.AddDate(0, 0, -1)

		params := url.Values{}
		params.Set("Start_time", dateToROC(start))
		params.Set("End_time", dateToROC(end))
		params.Set("MarketName", region)

		query := fmt.Sprintf("%s/AgriProductsTransType/?%s", moaAPIBase, params.Encode())

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

		var records []PriceRecord
		if err := json.Unmarshal(body, &records); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}

		seen := map[string]bool{}
		if asJSON, _ := cmd.Flags().GetBool("json"); asJSON {
			products := []string{}
			for _, r := range records {
				if !seen[r.CropName] {
					products = append(products, r.CropName)
					seen[r.CropName] = true
				}
			}
			json.NewEncoder(os.Stdout).Encode(map[string]any{
				"ok":    true,
				"count": len(products),
				"data":  products,
			})
		} else {
			fmt.Println("可查詢品項 (" + region + "):")
			count := 0
			for _, r := range records {
				if !seen[r.CropName] {
					fmt.Printf("  - %s\n", r.CropName)
					seen[r.CropName] = true
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

	listCmd.Flags().StringVarP(&region, "region", "r", "新竹", "查詢地區（預設：新竹）")
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
