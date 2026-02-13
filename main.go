package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
)

var verbose bool
var proxySourceMap = make(map[string]string)        // proxy -> source URL/file
var proxyProtocols = make(map[string]map[string]bool) // proxy -> set of protocols ("http","https")

// classifySource returns "http" or "https" based on the source URL path
func classifySource(sourceURL string) string {
	// Strip the base GitHub raw URL and check path components for "https"
	path := strings.TrimPrefix(strings.ToLower(sourceURL), "https://raw.githubusercontent.com/")
	for _, part := range strings.Split(path, "/") {
		if strings.Contains(part, "https") {
			return "https"
		}
	}
	return "http"
}

var DEFAULT_SOURCES = []string{
	"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
	"https://raw.githubusercontent.com/iplocate/free-proxy-list/main/protocols/http.txt",
	"https://raw.githubusercontent.com/ClearProxy/checked-proxy-list/main/http/raw/all.txt",
	"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/http.txt",
	"https://raw.githubusercontent.com/vakhov/fresh-proxy-list/master/https.txt",
	"https://raw.githubusercontent.com/monosans/proxy-list/refs/heads/main/proxies/http.txt",
	"https://raw.githubusercontent.com/mmpx12/proxy-list/refs/heads/master/http.txt",
	"https://raw.githubusercontent.com/Zaeem20/FREE_PROXIES_LIST/refs/heads/master/http.txt",
	"https://raw.githubusercontent.com/Zaeem20/FREE_PROXIES_LIST/refs/heads/master/https.txt",
	"https://raw.githubusercontent.com/sunny9577/proxy-scraper/refs/heads/master/proxies.txt",
	"https://raw.githubusercontent.com/mzyui/proxy-list/refs/heads/main/http.txt",
	"https://raw.githubusercontent.com/elliottophellia/proxylist/refs/heads/master/results/http/global/http_checked.txt",
	"https://raw.githubusercontent.com/officialputuid/KangProxy/refs/heads/KangProxy/http/http.txt",
	"https://raw.githubusercontent.com/officialputuid/KangProxy/refs/heads/KangProxy/https/https.txt",
	"https://raw.githubusercontent.com/databay-labs/free-proxy-list/refs/heads/master/http.txt",
	"https://raw.githubusercontent.com/r00tee/Proxy-List/refs/heads/main/Https.txt",
	"https://raw.githubusercontent.com/vmheaven/VMHeaven-Free-Proxy-Updated/refs/heads/main/http.txt",
	"https://raw.githubusercontent.com/Anonym0usWork1221/Free-Proxies/refs/heads/main/proxy_files/http_proxies.txt",
	"https://raw.githubusercontent.com/Anonym0usWork1221/Free-Proxies/refs/heads/main/proxy_files/https_proxies.txt",
	"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/http_checked.txt",
	"https://raw.githubusercontent.com/prxchk/proxy-list/main/http.txt",
	"https://raw.githubusercontent.com/TuanMinPay/live-proxy/master/http.txt",
	"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt",
	"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
	"https://raw.githubusercontent.com/Skillter/ProxyGather/refs/heads/master/proxies/working-proxies-http.txt",
}

// ProxyInfo holds detailed information about a working proxy
type ProxyInfo struct {
	Proxy       string     `json:"proxy"`
	TestedAt    string     `json:"tested_at"`
	IP          string     `json:"ip"`
	Success     bool       `json:"success"`
	Type        string     `json:"type"`
	Continent   string     `json:"continent"`
	Country     string     `json:"country"`
	CountryCode string     `json:"country_code"`
	Region      string     `json:"region"`
	RegionCode  string     `json:"region_code"`
	City        string     `json:"city"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	IsEU        bool       `json:"is_eu"`
	Postal      string     `json:"postal"`
	Connection  Connection `json:"connection"`
}

type Connection struct {
	ASN    int    `json:"asn"`
	Org    string `json:"org"`
	ISP    string `json:"isp"`
	Domain string `json:"domain"`
}

// Logging helpers
func logInfo(msg string, args ...any) {
	if verbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("%s[%s]%s %s%s%s ", Dim, timestamp, Reset, Cyan, msg, Reset)
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				fmt.Printf("%s%v%s=%v ", Yellow, args[i], Reset, args[i+1])
			}
		}
		fmt.Println()
	}
}

func logSuccess(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s[%s]%s %s✓ %s%s ", Dim, timestamp, Reset, Green+Bold, msg, Reset)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			fmt.Printf("%s%v%s=%v ", Yellow, args[i], Reset, args[i+1])
		}
	}
	fmt.Println()
}

func logError(msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s[%s]%s %s✗ %s%s ", Dim, timestamp, Reset, Red+Bold, msg, Reset)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			fmt.Printf("%s%v%s=%v ", Yellow, args[i], Reset, args[i+1])
		}
	}
	fmt.Println()
}

func logWarn(msg string, args ...any) {
	if verbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("%s[%s]%s %s⚠ %s%s ", Dim, timestamp, Reset, Yellow+Bold, msg, Reset)
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				fmt.Printf("%s%v%s=%v ", Yellow, args[i], Reset, args[i+1])
			}
		}
		fmt.Println()
	}
}

func logDebug(msg string, args ...any) {
	if verbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("%s[%s] %s ", Dim, timestamp, msg)
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				fmt.Printf("%v=%v ", args[i], args[i+1])
			}
		}
		fmt.Println(Reset)
	}
}

func printProgress(current, total, working int, proxy string) {
	if !verbose {
		percent := float64(current) / float64(total) * 100
		barWidth := 30
		filledWidth := int(float64(barWidth) * float64(current) / float64(total))
		bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

		// Truncate proxy for display - fixed width
		displayProxy := proxy
		if len(displayProxy) > 30 {
			displayProxy = displayProxy[:27] + "..."
		}
		// Pad to fixed width to overwrite old content
		displayProxy = fmt.Sprintf("%-30s", displayProxy)

		// Clear line with ANSI escape code and print
		fmt.Printf("\r\033[K%s[%s]%s %s%s%s %s%5.1f%%%s %s[%d/%d]%s %sWorking: %s%d%s %s%s%s",
			Dim, time.Now().Format("15:04:05"), Reset,
			Cyan, bar, Reset,
			Bold, percent, Reset,
			Dim, current, total, Reset,
			Green, Bold, working, Reset,
			Dim, displayProxy, Reset)
	}
}

// checkSourceFreshness checks if a GitHub raw URL's file was updated within the last 24 hours
func checkSourceFreshness(rawURL string, client *http.Client) (bool, time.Time) {
	trimmed := strings.TrimPrefix(rawURL, "https://raw.githubusercontent.com/")
	parts := strings.SplitN(trimmed, "/", 3) // owner, repo, rest
	if len(parts) < 3 {
		return true, time.Time{} // can't parse, assume fresh
	}
	owner := parts[0]
	repo := parts[1]
	rest := parts[2]

	var ref, filePath string
	if strings.HasPrefix(rest, "refs/heads/") {
		afterRefs := strings.TrimPrefix(rest, "refs/heads/")
		idx := strings.Index(afterRefs, "/")
		if idx < 0 {
			return true, time.Time{}
		}
		ref = afterRefs[:idx]
		filePath = afterRefs[idx+1:]
	} else {
		idx := strings.Index(rest, "/")
		if idx < 0 {
			return true, time.Time{}
		}
		ref = rest[:idx]
		filePath = rest[idx+1:]
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?path=%s&sha=%s&per_page=1",
		owner, repo, filePath, ref)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return true, time.Time{}
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ProxyWiz/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return true, time.Time{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return true, time.Time{} // rate limited or error, assume fresh
	}

	var commits []struct {
		Commit struct {
			Committer struct {
				Date string `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return true, time.Time{}
	}

	if err := json.Unmarshal(body, &commits); err != nil || len(commits) == 0 {
		return true, time.Time{}
	}

	lastUpdated, err := time.Parse(time.RFC3339, commits[0].Commit.Committer.Date)
	if err != nil {
		return true, time.Time{}
	}

	return time.Since(lastUpdated) <= 24*time.Hour, lastUpdated
}

// Function to fetch proxy list and save lines to Slice
func getList(proxyList string, custom bool) []string {
	var proxies []string

	if custom {
		// Use custom file
		content, err := os.Open(proxyList)
		if err != nil {
			logError("Failed to open proxy list", "file", proxyList, "error", err)
			os.Exit(1)
		}
		defer content.Close()
		scanner := bufio.NewScanner(content)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				proxies = append(proxies, line)
				proxySourceMap[line] = proxyList
				if proxyProtocols[line] == nil {
					proxyProtocols[line] = make(map[string]bool)
				}
				proxyProtocols[line]["http"] = true
			}
		}
		logSuccess("Discovered proxies", "count", len(proxies), "file", proxyList)
	} else {
		// Scrape from DEFAULT_SOURCES
		logInfo("Scraping proxies from default sources", "sources", len(DEFAULT_SOURCES))

		client := &http.Client{Timeout: 30 * time.Second}
		seen := make(map[string]bool) // Deduplicate proxies

		type sourceResult struct {
			url   string
			count int
		}
		var results []sourceResult

		// Pre-check freshness of all sources
		if !verbose {
			fmt.Printf("\r\033[K  %s⏳%s Checking source freshness via GitHub API...\n", Yellow, Reset)
		}
		var freshSources []string
		var skippedSources []string
		for i, source := range DEFAULT_SOURCES {
			if !verbose {
				short := strings.TrimPrefix(source, "https://raw.githubusercontent.com/")
				if len(short) > 50 {
					short = short[:47] + "..."
				}
				fmt.Printf("\r\033[K  %s[%d/%d]%s Checking %s", Dim, i+1, len(DEFAULT_SOURCES), Reset, short)
			}
			fresh, lastUpdated := checkSourceFreshness(source, client)
			if fresh {
				freshSources = append(freshSources, source)
				logDebug("Source is fresh", "url", source, "updated", lastUpdated.Format("2006-01-02 15:04"))
			} else {
				skippedSources = append(skippedSources, source)
				ago := time.Since(lastUpdated).Truncate(time.Hour)
				logDebug("Source stale, skipping", "url", source, "updated", lastUpdated.Format("2006-01-02 15:04"), "ago", ago)
			}
		}
		if !verbose {
			fmt.Println() // clear the checking line
		}

		if len(skippedSources) > 0 {
			logSuccess(fmt.Sprintf("Source freshness check: %s%d fresh%s, %s%d stale (>24h)%s",
				Green, len(freshSources), Reset+Green+Bold,
				Yellow, len(skippedSources), Reset))
			for _, src := range skippedSources {
				short := strings.TrimPrefix(src, "https://raw.githubusercontent.com/")
				if len(short) > 60 {
					short = short[:57] + "..."
				}
				fmt.Printf("  %s⏭  %s%s\n", Yellow, short, Reset)
			}
		} else {
			logSuccess(fmt.Sprintf("All %d sources are fresh (updated within 24h)", len(freshSources)))
		}

		for i, source := range freshSources {
			if verbose {
				logInfo("Fetching source", "current", i+1, "total", len(freshSources), "url", source)
			} else {
				printProgress(i+1, len(freshSources), len(proxies), source)
			}

			protocol := classifySource(source)

			resp, err := client.Get(source)
			if err != nil {
				logDebug("Failed to fetch source", "url", source, "error", err)
				results = append(results, sourceResult{url: source, count: -1})
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				logDebug("Failed to read response", "url", source, "error", err)
				results = append(results, sourceResult{url: source, count: -1})
				continue
			}

			lines := strings.Split(string(body), "\n")
			added := 0
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !seen[line] {
					seen[line] = true
					proxies = append(proxies, line)
					proxySourceMap[line] = source
					added++
				}
				if line != "" {
					if proxyProtocols[line] == nil {
						proxyProtocols[line] = make(map[string]bool)
					}
					proxyProtocols[line][protocol] = true
				}
			}
			results = append(results, sourceResult{url: source, count: added})
			logDebug("Added proxies from source", "count", added, "url", source)
		}

		if !verbose {
			fmt.Println() // New line after progress
		}
		logSuccess("Scraped proxies from default sources", "total", len(proxies), "fresh_sources", len(freshSources), "skipped", len(skippedSources))
	}

	return proxies
}

// loadProxies loads proxies from file or scrapes from default sources
func loadProxies(proxyFile string) []string {
	if proxyFile == "" {
		return getList("", false)
	}
	return getList(proxyFile, true)
}

// shortSource returns a truncated source name for display
func shortSource(source string, maxLen int) string {
	short := strings.TrimPrefix(source, "https://raw.githubusercontent.com/")
	if len(short) > maxLen {
		short = short[:maxLen-3] + "..."
	}
	return short
}

// fetchAllSourcesForAnalysis fetches from all default sources and tracks ALL sources per proxy.
// Unlike getList, this does NOT deduplicate across sources — it records every source a proxy appears in.
func fetchAllSourcesForAnalysis() ([]string, map[string][]string, map[string]int) {
	multiSourceMap := make(map[string][]string) // proxy -> all source URLs
	sourceTotalCount := make(map[string]int)    // source -> total proxy count
	client := &http.Client{Timeout: 30 * time.Second}

	// Check freshness
	if !verbose {
		fmt.Printf("\r\033[K  %s⏳%s Checking source freshness via GitHub API...\n", Yellow, Reset)
	}
	var freshSources []string
	var skippedSources []string
	for i, source := range DEFAULT_SOURCES {
		if !verbose {
			short := strings.TrimPrefix(source, "https://raw.githubusercontent.com/")
			if len(short) > 50 {
				short = short[:47] + "..."
			}
			fmt.Printf("\r\033[K  %s[%d/%d]%s Checking %s", Dim, i+1, len(DEFAULT_SOURCES), Reset, short)
		}
		fresh, lastUpdated := checkSourceFreshness(source, client)
		if fresh {
			freshSources = append(freshSources, source)
		} else {
			skippedSources = append(skippedSources, source)
			ago := time.Since(lastUpdated).Truncate(time.Hour)
			logDebug("Source stale, skipping", "url", source, "ago", ago)
		}
	}
	if !verbose {
		fmt.Println()
	}

	if len(skippedSources) > 0 {
		logSuccess(fmt.Sprintf("Source freshness: %s%d fresh%s, %s%d stale (>24h)%s",
			Green, len(freshSources), Reset+Green+Bold,
			Yellow, len(skippedSources), Reset))
		for _, src := range skippedSources {
			short := strings.TrimPrefix(src, "https://raw.githubusercontent.com/")
			if len(short) > 60 {
				short = short[:57] + "..."
			}
			fmt.Printf("  %s⏭  %s%s\n", Yellow, short, Reset)
		}
	} else {
		logSuccess(fmt.Sprintf("All %d sources are fresh (updated within 24h)", len(freshSources)))
	}

	// Fetch from each source — track ALL sources per proxy (no dedup skipping)
	for i, source := range freshSources {
		if !verbose {
			short := strings.TrimPrefix(source, "https://raw.githubusercontent.com/")
			if len(short) > 50 {
				short = short[:47] + "..."
			}
			fmt.Printf("\r\033[K  %s[%d/%d]%s Fetching %s", Dim, i+1, len(freshSources), Reset, short)
		}

		resp, err := client.Get(source)
		if err != nil {
			logDebug("Failed to fetch", "url", source, "error", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		lines := strings.Split(string(body), "\n")
		count := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				multiSourceMap[line] = append(multiSourceMap[line], source)
				count++
			}
		}
		sourceTotalCount[source] = count
	}
	if !verbose {
		fmt.Println()
	}

	// Build unique proxy list
	proxies := make([]string, 0, len(multiSourceMap))
	for p := range multiSourceMap {
		proxies = append(proxies, p)
		proxySourceMap[p] = multiSourceMap[p][0]
	}

	logSuccess("Fetched all sources for analysis", "unique_proxies", len(proxies), "sources", len(freshSources))
	return proxies, multiSourceMap, sourceTotalCount
}

// testProxy tests a single proxy and returns ProxyInfo if it works, nil otherwise
func testProxy(proxyStr string, timeout int) *ProxyInfo {
	//originalProxy := proxyStr
	// Only add http:// if the proxy doesn't already have a scheme
	if !strings.HasPrefix(proxyStr, "http://") && !strings.HasPrefix(proxyStr, "https://") {
		proxyStr = "http://" + proxyStr
	}

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		logWarn("Invalid proxy format, skipping", "proxy", proxyStr, "error", err)
		return nil
	}

	tr := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    time.Duration(timeout) * time.Second,
		DisableCompression: true,
		Proxy:              http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	resp, err := client.Get("https://ipwho.is/")
	if err != nil {
		logDebug("Failed", "proxy", proxyStr, "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logDebug("Bad status", "proxy", proxyStr, "status", resp.Status)
		return nil
	}

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logWarn("Failed to read response body", "proxy", proxyStr, "error", err)
		return nil
	}

	// Parse JSON response from ipwho.is
	var ipData struct {
		IP          string  `json:"ip"`
		Success     bool    `json:"success"`
		Type        string  `json:"type"`
		Continent   string  `json:"continent"`
		Country     string  `json:"country"`
		CountryCode string  `json:"country_code"`
		Region      string  `json:"region"`
		RegionCode  string  `json:"region_code"`
		City        string  `json:"city"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		IsEU        bool    `json:"is_eu"`
		Postal      string  `json:"postal"`
		Connection  struct {
			ASN    int    `json:"asn"`
			Org    string `json:"org"`
			ISP    string `json:"isp"`
			Domain string `json:"domain"`
		} `json:"connection"`
	}

	if err := json.Unmarshal(body, &ipData); err != nil {
		logWarn("Failed to parse JSON response", "proxy", proxyStr, "error", err)
		return nil
	}

	if !ipData.Success {
		logDebug("IP lookup failed", "proxy", proxyStr)
		return nil
	}

	logSuccess("Working proxy found", "proxy", proxyStr, "ip", ipData.IP, "country", ipData.Country, "city", ipData.City)

	return &ProxyInfo{
		Proxy:       proxyStr,
		TestedAt:    time.Now().UTC().Format(time.RFC3339),
		IP:          ipData.IP,
		Success:     ipData.Success,
		Type:        ipData.Type,
		Continent:   ipData.Continent,
		Country:     ipData.Country,
		CountryCode: ipData.CountryCode,
		Region:      ipData.Region,
		RegionCode:  ipData.RegionCode,
		City:        ipData.City,
		Latitude:    ipData.Latitude,
		Longitude:   ipData.Longitude,
		IsEU:        ipData.IsEU,
		Postal:      ipData.Postal,
		Connection: Connection{
			ASN:    ipData.Connection.ASN,
			Org:    ipData.Connection.Org,
			ISP:    ipData.Connection.ISP,
			Domain: ipData.Connection.Domain,
		},
	}
}

// saveJSON saves the working proxies to a JSON file in the given directory
func saveJSON(workingProxies []ProxyInfo, dir, name string) {
	if len(workingProxies) == 0 {
		return
	}
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, name+".json")
	jsonData, err := json.MarshalIndent(workingProxies, "", "  ")
	if err != nil {
		logError("Failed to marshal JSON", "error", err)
		return
	}
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		logError("Failed to write JSON file", "error", err)
	} else {
		logInfo("Saved detailed proxy info", "file", path, "count", len(workingProxies))
	}
}

// saveCSV saves working proxies as ip:port,country,region to a CSV file in the given directory
func saveCSV(workingProxies []ProxyInfo, dir, name string) {
	if len(workingProxies) == 0 {
		return
	}
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, name+".csv")
	file, err := os.Create(path)
	if err != nil {
		logError("Failed to create CSV file", "error", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"proxy", "country", "region"})
	for _, p := range workingProxies {
		proxy := strings.TrimPrefix(p.Proxy, "http://")
		proxy = strings.TrimPrefix(proxy, "https://")
		writer.Write([]string{proxy, p.Country, p.Region})
	}
	logInfo("Saved CSV proxy list", "file", path, "count", len(workingProxies))
}

// saveTXT saves the proxy list as plain text in the given directory
func saveTXT(workingProxies []ProxyInfo, dir, name string) {
	if len(workingProxies) == 0 {
		return
	}
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, name+".txt")
	file, err := os.Create(path)
	if err != nil {
		logError("Failed to create txt file", "error", err)
		return
	}
	defer file.Close()
	for _, p := range workingProxies {
		proxy := strings.TrimPrefix(p.Proxy, "http://")
		proxy = strings.TrimPrefix(proxy, "https://")
		file.WriteString(proxy + "\n")
	}
	logInfo("Saved proxy list", "file", path, "count", len(workingProxies))
}

// saveAllFormats saves working proxies split by protocol into http/ and https/ folders
func saveAllFormats(workingProxies []ProxyInfo) {
	var httpProxies, httpsProxies []ProxyInfo
	for _, pi := range workingProxies {
		raw := strings.TrimPrefix(pi.Proxy, "http://")
		raw = strings.TrimPrefix(raw, "https://")
		protocols := proxyProtocols[raw]
		if protocols["http"] {
			httpProxies = append(httpProxies, pi)
		}
		if protocols["https"] {
			httpsProxies = append(httpsProxies, pi)
		}
		// If no protocol info (custom file), default to http
		if len(protocols) == 0 {
			httpProxies = append(httpProxies, pi)
		}
	}

	saveTXT(httpProxies, "http", "http")
	saveCSV(httpProxies, "http", "http")
	saveJSON(httpProxies, "http", "http")

	saveTXT(httpsProxies, "https", "https")
	saveCSV(httpsProxies, "https", "https")
	saveJSON(httpsProxies, "https", "https")
}

// gitPush stages, commits and pushes http/ and https/ folders to the remote repo
func gitPush(httpCount, httpsCount, totalCount int) {
	files := []string{"http/", "https/"}
	logSuccess("Pushing results to GitHub...")

	// Stage files
	args := append([]string{"add"}, files...)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		logError("git add failed", "error", err, "output", string(out))
		return
	}

	// Commit
	msg := fmt.Sprintf("Update proxy list: %d http + %d https working out of %d tested", httpCount, httpsCount, totalCount)
	if out, err := exec.Command("git", "commit", "-m", msg).CombinedOutput(); err != nil {
		if strings.Contains(string(out), "nothing to commit") {
			logInfo("No changes to commit")
			return
		}
		logError("git commit failed", "error", err, "output", string(out))
		return
	}

	// Push
	if out, err := exec.Command("git", "push").CombinedOutput(); err != nil {
		logError("git push failed", "error", err, "output", string(out))
		return
	}

	logSuccess("Results pushed to GitHub", "files", "http/ https/")
}

// checkerMain orchestrates the proxy checking process with concurrent workers
func checkerMain(proxyFile string, timeout int, numWorkers int) {
	raw := loadProxies(proxyFile)

	// Deduplicate right before checking
	seen := make(map[string]bool)
	proxies := make([]string, 0, len(raw))
	for _, p := range raw {
		if !seen[p] {
			seen[p] = true
			proxies = append(proxies, p)
		}
	}
	if len(raw) != len(proxies) {
		logSuccess("Deduplicated proxy list", "before", len(raw), "after", len(proxies), "removed", len(raw)-len(proxies))
	}

	rand.Shuffle(len(proxies), func(i, j int) {
		proxies[i], proxies[j] = proxies[j], proxies[i]
	})

	// Create output directories
	os.MkdirAll("http", 0755)
	os.MkdirAll("https", 0755)

	// Thread-safe storage for results
	var workingProxies []ProxyInfo
	var mu sync.Mutex // protects workingProxies and counters
	var tested int

	// Channels
	jobs := make(chan string, numWorkers*2) // buffered job queue
	var wg sync.WaitGroup

	// Handle Ctrl+C to save before exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println()
		logWarn("Interrupt received, saving progress...")
		mu.Lock()
		saveAllFormats(workingProxies)
		mu.Unlock()
		os.Exit(0)
	}()

	// Spawn workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for proxyStr := range jobs {
				// Test the proxy
				proxyInfo := testProxy(proxyStr, timeout)

				// Update shared state safely
				mu.Lock()
				tested++
				current := tested
				working := len(workingProxies)

				if proxyInfo != nil {
					workingProxies = append(workingProxies, *proxyInfo)
					working = len(workingProxies)
					saveAllFormats(workingProxies)
				}

				if !verbose {
					printProgress(current, len(proxies), working, proxyStr)
				}
				mu.Unlock()
			}
		}(i)
	}

	// Send all proxies to the job queue
	for _, proxy := range proxies {
		jobs <- proxy
	}
	close(jobs) // signal no more jobs

	// Wait for all workers to finish
	wg.Wait()

	if !verbose {
		fmt.Println()
	}
	saveAllFormats(workingProxies)

	// Count per protocol for summary
	var httpCount, httpsCount int
	for _, pi := range workingProxies {
		raw := strings.TrimPrefix(pi.Proxy, "http://")
		raw = strings.TrimPrefix(raw, "https://")
		if proxyProtocols[raw]["http"] || len(proxyProtocols[raw]) == 0 {
			httpCount++
		}
		if proxyProtocols[raw]["https"] {
			httpsCount++
		}
	}
	logSuccess("Check complete", "http", httpCount, "https", httpsCount, "total_tested", len(proxies))

	// Per-source working proxy summary
	if len(workingProxies) > 0 {
		// Tally working proxies per source
		sourceCounts := make(map[string]int)
		var sourceOrder []string
		sourceSeen := make(map[string]bool)
		for _, pi := range workingProxies {
			// Strip http:// or https:// prefix to match the raw key in proxySourceMap
			raw := pi.Proxy
			raw = strings.TrimPrefix(raw, "http://")
			raw = strings.TrimPrefix(raw, "https://")
			src, ok := proxySourceMap[raw]
			if !ok {
				src = "unknown"
			}
			if !sourceSeen[src] {
				sourceSeen[src] = true
				sourceOrder = append(sourceOrder, src)
			}
			sourceCounts[src]++
		}

		fmt.Printf("\n%s%s  ┌─────────────────────────────────────────────────────────────────────┐%s\n", Bold, Dim, Reset)
		fmt.Printf("%s%s  │%s %s%-54s%s %s%9s%s %s│%s\n", Bold, Dim, Reset, Cyan+Bold, "SOURCE", Reset, Yellow+Bold, "WORKING", Reset, Dim, Reset)
		fmt.Printf("%s%s  ├─────────────────────────────────────────────────────────────────────┤%s\n", Bold, Dim, Reset)
		for _, src := range sourceOrder {
			short := src
			short = strings.TrimPrefix(short, "https://raw.githubusercontent.com/")
			if len(short) > 54 {
				short = short[:51] + "..."
			}
			fmt.Printf("%s  │%s %-54s %s%9d%s %s│%s\n", Dim, Reset, short, Green+Bold, sourceCounts[src], Reset, Dim, Reset)
		}
		fmt.Printf("%s%s  └─────────────────────────────────────────────────────────────────────┘%s\n", Bold, Dim, Reset)
	}

	// Auto-push results to GitHub
	gitPush(httpCount, httpsCount, len(proxies))
}

// analyzeSourcesMain runs source quality analysis to identify bad/redundant proxy sources
func analyzeSourcesMain(timeout, numWorkers int) {
	proxies, multiSourceMap, sourceTotalCount := fetchAllSourcesForAnalysis()

	// ── Pre-test overlap analysis ──────────────────────────────────────
	fmt.Printf("\n%s%s  ┌─────────────────────────────────────────────────────────────────────────────┐%s\n", Bold, Dim, Reset)
	fmt.Printf("%s%s  │%s %s%-73s%s %s│%s\n", Bold, Dim, Reset, Yellow+Bold, "PRE-TEST SOURCE OVERLAP ANALYSIS", Reset, Dim, Reset)
	fmt.Printf("%s%s  └─────────────────────────────────────────────────────────────────────────────┘%s\n\n", Bold, Dim, Reset)

	// Compute per-source unique count (proxies found ONLY in that source)
	sourceUniqueCount := make(map[string]int)
	for _, sources := range multiSourceMap {
		if len(sources) == 1 {
			sourceUniqueCount[sources[0]]++
		}
	}

	fmt.Printf("  %s%-44s %7s %7s %9s%s\n", Bold, "SOURCE", "TOTAL", "UNIQUE", "OVERLAP%", Reset)
	fmt.Printf("  %s%s%s\n", Dim, strings.Repeat("─", 70), Reset)

	for _, source := range DEFAULT_SOURCES {
		total := sourceTotalCount[source]
		if total == 0 {
			continue
		}
		unique := sourceUniqueCount[source]
		overlapPct := float64(total-unique) / float64(total) * 100

		overlapColor := Green
		if overlapPct > 95 {
			overlapColor = Red + Bold
		} else if overlapPct > 80 {
			overlapColor = Yellow
		}

		fmt.Printf("  %-44s %s%7d%s %s%7d%s %s%8.1f%%%s\n",
			shortSource(source, 44), Dim, total, Reset, Cyan, unique, Reset,
			overlapColor, overlapPct, Reset)
	}

	// ── Pairwise overlap for worst offenders ───────────────────────────
	sourceProxies := make(map[string]map[string]bool)
	for proxy, sources := range multiSourceMap {
		for _, src := range sources {
			if sourceProxies[src] == nil {
				sourceProxies[src] = make(map[string]bool)
			}
			sourceProxies[src][proxy] = true
		}
	}

	type overlapPair struct {
		a, b       string
		shared     int
		pctSmaller float64
	}

	srcList := make([]string, 0, len(sourceProxies))
	for s := range sourceProxies {
		srcList = append(srcList, s)
	}
	sort.Strings(srcList)

	var pairs []overlapPair
	for i := 0; i < len(srcList); i++ {
		for j := i + 1; j < len(srcList); j++ {
			shared := 0
			for p := range sourceProxies[srcList[i]] {
				if sourceProxies[srcList[j]][p] {
					shared++
				}
			}
			if shared > 0 {
				smaller := len(sourceProxies[srcList[i]])
				if len(sourceProxies[srcList[j]]) < smaller {
					smaller = len(sourceProxies[srcList[j]])
				}
				pairs = append(pairs, overlapPair{
					a: srcList[i], b: srcList[j],
					shared:     shared,
					pctSmaller: float64(shared) / float64(smaller) * 100,
				})
			}
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].pctSmaller > pairs[j].pctSmaller
	})

	fmt.Printf("\n  %s%sHighly Overlapping Pairs (>80%% of smaller source shared):%s\n", Bold, Yellow, Reset)
	fmt.Printf("  %s%s%s\n", Dim, strings.Repeat("─", 70), Reset)

	shown := 0
	for _, p := range pairs {
		if p.pctSmaller < 80 || shown >= 10 {
			break
		}
		fmt.Printf("  %s%d shared (%.0f%%)%s: %s <-> %s\n",
			Red, p.shared, p.pctSmaller, Reset,
			shortSource(p.a, 30), shortSource(p.b, 30))
		shown++
	}
	if shown == 0 {
		fmt.Printf("  %sNo pairs with >80%% overlap found%s\n", Dim, Reset)
	}

	// ── Test all unique proxies ────────────────────────────────────────
	fmt.Printf("\n%s%s  ┌─────────────────────────────────────────────────────────────────────────────┐%s\n", Bold, Dim, Reset)
	fmt.Printf("%s%s  │%s %s%-73s%s %s│%s\n", Bold, Dim, Reset, Yellow+Bold,
		fmt.Sprintf("Testing %d unique proxies...", len(proxies)), Reset, Dim, Reset)
	fmt.Printf("%s%s  └─────────────────────────────────────────────────────────────────────────────┘%s\n\n", Bold, Dim, Reset)

	rand.Shuffle(len(proxies), func(i, j int) {
		proxies[i], proxies[j] = proxies[j], proxies[i]
	})

	var workingProxies []ProxyInfo
	var mu sync.Mutex
	var tested int

	jobs := make(chan string, numWorkers*2)
	var wg sync.WaitGroup

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println()
		logWarn("Interrupt received, saving progress...")
		mu.Lock()
		saveAllFormats(workingProxies)
		mu.Unlock()
		os.Exit(0)
	}()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for proxyStr := range jobs {
				proxyInfo := testProxy(proxyStr, timeout)
				mu.Lock()
				tested++
				if proxyInfo != nil {
					workingProxies = append(workingProxies, *proxyInfo)
				}
				if !verbose {
					printProgress(tested, len(proxies), len(workingProxies), proxyStr)
				}
				mu.Unlock()
			}
		}()
	}

	for _, proxy := range proxies {
		jobs <- proxy
	}
	close(jobs)
	wg.Wait()

	if !verbose {
		fmt.Println()
	}
	logSuccess("Testing complete", "working", len(workingProxies), "total", len(proxies))

	// ── Post-test source results ───────────────────────────────────────
	fmt.Printf("\n%s%s  ┌───────────────────────────────────────────────────────────────────────────────────┐%s\n", Bold, Dim, Reset)
	fmt.Printf("%s%s  │%s %s%-79s%s %s│%s\n", Bold, Dim, Reset, Yellow+Bold, "SOURCE ANALYSIS RESULTS", Reset, Dim, Reset)
	fmt.Printf("%s%s  └───────────────────────────────────────────────────────────────────────────────────┘%s\n\n", Bold, Dim, Reset)

	// Tally working proxies per source (working proxy counts for ALL its sources)
	sourceWorkingCount := make(map[string]int)
	sourceWorkingUniqueCount := make(map[string]int)

	for _, pi := range workingProxies {
		raw := strings.TrimPrefix(pi.Proxy, "http://")
		raw = strings.TrimPrefix(raw, "https://")
		if sources, ok := multiSourceMap[raw]; ok {
			for _, src := range sources {
				sourceWorkingCount[src]++
			}
			if len(sources) == 1 {
				sourceWorkingUniqueCount[sources[0]]++
			}
		}
	}

	fmt.Printf("  %s%-44s %6s %6s %5s %6s  %-6s%s\n", Bold, "SOURCE", "TOTAL", "UNIQ", "WORK", "W.UNIQ", "STATUS", Reset)
	fmt.Printf("  %s%s%s\n", Dim, strings.Repeat("─", 81), Reset)

	var removeSources []string
	for _, source := range DEFAULT_SOURCES {
		total := sourceTotalCount[source]
		if total == 0 {
			continue
		}
		unique := sourceUniqueCount[source]
		working := sourceWorkingCount[source]
		workingUniq := sourceWorkingUniqueCount[source]

		status := Green + Bold + "KEEP  " + Reset
		if working < 2 {
			status = Red + Bold + "REMOVE" + Reset
			removeSources = append(removeSources, source)
		}

		workColor := Green
		if working < 2 {
			workColor = Red + Bold
		} else if working < 5 {
			workColor = Yellow
		}

		fmt.Printf("  %-44s %s%6d%s %s%6d%s %s%5d%s %s%6d%s  %s\n",
			shortSource(source, 44), Dim, total, Reset, Cyan, unique, Reset,
			workColor, working, Reset, workColor, workingUniq, Reset, status)
	}

	// ── Recommendations ────────────────────────────────────────────────
	fmt.Printf("\n%s%s  ┌───────────────────────────────────────────────────────────────────────────────────┐%s\n", Bold, Dim, Reset)
	fmt.Printf("%s%s  │%s %s%-79s%s %s│%s\n", Bold, Dim, Reset, Yellow+Bold, "RECOMMENDATIONS", Reset, Dim, Reset)
	fmt.Printf("%s%s  └───────────────────────────────────────────────────────────────────────────────────┘%s\n\n", Bold, Dim, Reset)

	if len(removeSources) > 0 {
		fmt.Printf("  %s%sRemove these sources (<2 working proxies):%s\n\n", Bold, Red, Reset)
		for i, src := range removeSources {
			working := sourceWorkingCount[src]
			unique := sourceUniqueCount[src]
			fmt.Printf("  %s%d.%s %s\n", Red, i+1, Reset, shortSource(src, 80))
			fmt.Printf("     %s%d working%s, %s%d unique to this source%s\n\n",
				Yellow, working, Reset, Cyan, unique, Reset)
		}
	} else {
		fmt.Printf("  %sAll sources contribute 2+ working proxies.%s\n", Green+Bold, Reset)
	}

	kept := len(sourceTotalCount) - len(removeSources)
	fmt.Printf("  %sSummary: %sKeep %d%s / %sRemove %d%s out of %d active sources%s\n",
		Bold, Green, kept, Reset+Bold, Red, len(removeSources), Reset+Bold, len(sourceTotalCount), Reset)

	// Save analysis to JSON
	type SourceAnalysis struct {
		Source        string  `json:"source"`
		TotalProxies  int    `json:"total_proxies"`
		UniqueProxies int    `json:"unique_proxies"`
		Working       int    `json:"working_proxies"`
		WorkingUnique int    `json:"working_unique_proxies"`
		OverlapPct    float64 `json:"overlap_percent"`
		Status        string  `json:"status"`
	}

	var analysisResults []SourceAnalysis
	for _, source := range DEFAULT_SOURCES {
		total := sourceTotalCount[source]
		if total == 0 {
			continue
		}
		unique := sourceUniqueCount[source]
		overlapPct := float64(total-unique) / float64(total) * 100

		status := "keep"
		if sourceWorkingCount[source] < 2 {
			status = "remove"
		}
		analysisResults = append(analysisResults, SourceAnalysis{
			Source:        source,
			TotalProxies:  total,
			UniqueProxies: unique,
			Working:       sourceWorkingCount[source],
			WorkingUnique: sourceWorkingUniqueCount[source],
			OverlapPct:    overlapPct,
			Status:        status,
		})
	}

	jsonData, err := json.MarshalIndent(analysisResults, "", "  ")
	if err == nil {
		os.WriteFile("source_analysis.json", jsonData, 0644)
		fmt.Printf("\n  %sFull analysis saved to %ssource_analysis.json%s\n", Dim, Cyan, Reset)
	}
}

func worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Printf("Worker %d processing: %s\n", id, job)
		time.Sleep(500 * time.Millisecond) // simulate work
	}
}
func printBanner() {
	banner := []string{
		" ███████████  ███████████      ███████    █████ █████ █████ █████    █████   ███   █████ █████ ███████████",
		"░░███░░░░░███░░███░░░░░███   ███░░░░░███ ░░███ ░░███ ░░███ ░░███    ░░███   ░███  ░░███ ░░███ ░█░░░░░░███ ",
		" ░███    ░███ ░███    ░███  ███     ░░███ ░░███ ███   ░░███ ███      ░███   ░███   ░███  ░███ ░     ███░  ",
		" ░██████████  ░██████████  ░███      ░███  ░░█████     ░░█████       ░███   ░███   ░███  ░███      ███    ",
		" ░███░░░░░░   ░███░░░░░███ ░███      ░███   ███░███     ░░███        ░░███  █████  ███   ░███     ███     ",
		" ░███         ░███    ░███ ░░███     ███   ███ ░░███     ░███         ░░░█████░█████░    ░███   ████     █",
		" █████        █████   █████ ░░░███████░   █████ █████    █████          ░░███ ░░███      █████ ███████████",
		"░░░░░        ░░░░░   ░░░░░    ░░░░░░░    ░░░░░ ░░░░░    ░░░░░            ░░░   ░░░      ░░░░░ ░░░░░░░░░░░",
	}

	// Color gradient top to bottom
	colors := []string{Magenta, Magenta, Cyan, Cyan, Cyan, Blue, Blue, Dim}

	fmt.Println()
	for i, line := range banner {
		fmt.Printf("%s%s%s\n", colors[i], line, Reset)
	}

	fmt.Println()
	fmt.Printf("%s%s   🧙 The Fastest HTTP/S Proxy Checker/Scraper 🧙%s\n", Green+Bold, strings.Repeat(" ", 15), Reset)
	fmt.Printf("%s%s@Syn2Much%s\n\n", Dim, strings.Repeat(" ", 50), Reset)
}

func main() {
	var numWorkers int
	var timeout int
	var verboseInput string
	printBanner()

	var fileName string
	fmt.Printf("%sEnter proxy file name (leave empty to scrape from default sources):%s ", Cyan, Reset)
	fmt.Scanln(&fileName)

	fmt.Printf("%sAmount of Workers:%s ", Cyan, Reset)
	fmt.Scanln(&numWorkers)
	if numWorkers < 1 {
		numWorkers = 10 // sensible default
	}

	fmt.Printf("%sTimeout (seconds):%s ", Cyan, Reset)
	fmt.Scanln(&timeout)
	if timeout < 1 {
		timeout = 8 // sensible default
	}

	fmt.Printf("%sVerbose mode? (y/N):%s ", Cyan, Reset)
	fmt.Scanln(&verboseInput)
	verbose = strings.ToLower(verboseInput) == "y" || strings.ToLower(verboseInput) == "yes"

	// Source analysis mode (only available when using default sources)
	var analyzeInput string
	if fileName == "" {
		fmt.Printf("%sRun source analysis mode? Identifies bad/redundant sources (y/N):%s ", Cyan, Reset)
		fmt.Scanln(&analyzeInput)
	}

	if strings.ToLower(analyzeInput) == "y" || strings.ToLower(analyzeInput) == "yes" {
		analyzeSourcesMain(timeout, numWorkers)
	} else {
		checkerMain(fileName, timeout, numWorkers)
	}

	fmt.Println("All done")
}
