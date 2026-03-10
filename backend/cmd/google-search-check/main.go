package main

import (
	"backend/internal/config"
	"backend/internal/external/companysearch"
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
)

func main() {
	keyword := flag.String("keyword", "led light manufacturer", "google search keyword")
	pageLimit := flag.Int("pages", 1, "number of pages to fetch")
	flag.Parse()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	fmt.Println("=== Google Search Check ===")
	fmt.Printf("GOOGLE_API_KEY: %s\n", maskSecret(cfg.GoogleAPIKey))
	fmt.Printf("GOOGLE_CX: %s\n", cfg.GoogleCX)
	fmt.Printf("GOOGLE_SEARCH_NUM: %d\n", cfg.GoogleSearchNum)
	if strings.TrimSpace(cfg.GoogleProxyURL) == "" {
		fmt.Println("GOOGLE_PROXY_URL: (not set)")
	} else {
		fmt.Printf("GOOGLE_PROXY_URL: %s\n", cfg.GoogleProxyURL)
	}
	fmt.Println()

	client := companysearch.NewHTTPClient(companysearch.HTTPClientConfig{
		ProxyURL: cfg.GoogleProxyURL,
	})
	provider := companysearch.NewGoogleProvider(client, cfg.GoogleAPIKey, cfg.GoogleCX, cfg.GoogleSearchNum)

	resultCount := 0
	err := provider.Search(context.Background(), companysearch.SearchRequest{
		Keyword:   strings.TrimSpace(*keyword),
		PageLimit: *pageLimit,
	}, func(page companysearch.SearchPage) error {
		fmt.Printf("Page %d: %d results\n", page.PageNo, len(page.Items))
		for index, item := range page.Items {
			if index >= 3 {
				break
			}
			fmt.Printf("  %d. %s\n", index+1, item.CompanyName)
			if item.CompanyURL != "" {
				fmt.Printf("     %s\n", item.CompanyURL)
			}
		}
		resultCount += len(page.Items)
		return nil
	})
	if err != nil {
		log.Fatalf("google search check failed: %v", err)
	}

	fmt.Println()
	fmt.Printf("Search completed successfully, total fetched: %d\n", resultCount)
}

func maskSecret(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 8 {
		if value == "" {
			return "(not set)"
		}
		return "****"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
