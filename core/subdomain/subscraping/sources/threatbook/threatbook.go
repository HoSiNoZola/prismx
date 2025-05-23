// Package threatbook logic
package threatbook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"prismx_cli/core/subdomain/subscraping"
	"strconv"
)

type threatBookResponse struct {
	ResponseCode int64  `json:"response_code"`
	VerboseMsg   string `json:"verbose_msg"`
	Data         struct {
		Domain     string `json:"domain"`
		SubDomains struct {
			Total string   `json:"total"`
			Data  []string `json:"data"`
		} `json:"sub_domains"`
	} `json:"data"`
}

// Source is the passive scraping agent
type Source struct{}

// Run function returns all subdomains found with the service
func (s *Source) Run(ctx context.Context, domain string, session *subscraping.Session) <-chan subscraping.Result {
	results := make(chan subscraping.Result)

	go func() {
		defer close(results)

		if session.Keys.ThreatBook == "" {
			return
		}

		resp, err := session.SimpleGet(ctx, fmt.Sprintf("https://api.threatbook.cn/v3/domain/sub_domains?apikey=%s&resource=%s", session.Keys.ThreatBook, domain))

		if err != nil && resp == nil {
			results <- subscraping.Result{Source: s.Name()}
			return
		}
		defer resp.Body.Close()
		var response threatBookResponse
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			results <- subscraping.Result{Source: s.Name()}
			return
		}
		if err = json.Unmarshal(all, &response); err != nil {
			results <- subscraping.Result{Source: s.Name()}
			return
		}
		if response.ResponseCode != 0 {
			results <- subscraping.Result{Source: s.Name()}
			return
		}

		total, err := strconv.ParseInt(response.Data.SubDomains.Total, 10, 64)
		if err != nil {
			results <- subscraping.Result{Source: s.Name()}
			return
		}

		if total > 0 {
			for _, subdomain := range response.Data.SubDomains.Data {
				results <- subscraping.Result{Source: s.Name(), Value: subdomain}
			}
		}
	}()

	return results
}

// Name returns the name of the source
func (s *Source) Name() string {
	return "threatbook"
}
