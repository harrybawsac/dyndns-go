package main

import (
	"context"
	"dyndns-go/lib"
	"dyndns-go/schema"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// DiscoverPublicIP fetches your public IPv4 and IPv6 addresses from a JSON API response.
// It expects the response to contain reportedState.wans[0].ipv4 and reportedState.wans[0].ipv6 fields.
func DiscoverPublicIP(ctx context.Context, config *lib.Config) (ipv4 net.IP, ipv6 net.IP, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, schema.UnifiSiteManagerUrl+config.UnifiSiteManagerHostId, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "dyndns2-go/1.0")
	req.Header.Set("X-API-KEY", config.UnifiSiteManagerApiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	// Parse JSON response
	type wanInfo struct {
		IPv4 string `json:"ipv4"`
		IPv6 string `json:"ipv6"`
	}
	type reportedState struct {
		Wans []wanInfo `json:"wans"`
	}
	type data struct {
		ReportedState reportedState `json:"reportedState"`
	}
	type response struct {
		Data data `json:"data"`
	}

	var resp response
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&resp); err != nil {
		return nil, nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	if len(resp.Data.ReportedState.Wans) == 0 {
		return nil, nil, fmt.Errorf("no WAN info found in response")
	}

	wan := resp.Data.ReportedState.Wans[0]
	if wan.IPv4 != "" {
		ipv4 = net.ParseIP(wan.IPv4)
	}
	if wan.IPv6 != "" {
		ipv6 = net.ParseIP(wan.IPv6)
	}
	if ipv4 == nil && wan.IPv4 != "" {
		return nil, nil, fmt.Errorf("failed to parse IPv4 from %q", wan.IPv4)
	}
	if ipv6 == nil && wan.IPv6 != "" {
		return nil, nil, fmt.Errorf("failed to parse IPv6 from %q", wan.IPv6)
	}
	return ipv4, ipv6, nil
}

func main() {
	ctx := context.Background()

	// Parse flags for config and storage paths
	var configPath string
	var storagePath string
	flag.StringVar(&configPath, "config", "", "Path to the configuration file (required)")
	flag.StringVar(&storagePath, "storage", "", "Path to the storage file (required)")
	flag.Parse()

	if configPath == "" {
		fmt.Println("Error: -config flag is required.")
		return
	}
	if storagePath == "" {
		fmt.Println("Error: -storage flag is required.")
		return
	}

	// Load configuration from configPath
	config, err := lib.LoadConfig(configPath)
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	var ipv4, ipv6 net.IP

	ip4, ip6, discoverErr := DiscoverPublicIP(ctx, config)
	if discoverErr != nil {
		fmt.Println("discover public IPs error:", discoverErr)
		return
	}
	if config.UpdateIpv4 {
		ipv4 = ip4
	}
	if config.UpdateIpv6 {
		ipv6 = ip6
	}
	fmt.Println("Discovered IPv4:", ipv4, "IPv6:", ipv6)

	// Read previous IPs from storage
	prev, _ := lib.ReadIPData(storagePath) // ignore error if file doesn't exist

	newData := lib.IPData{
		IPv4: "",
		IPv6: "",
	}
	if config.UpdateIpv4 && ipv4 != nil {
		newData.IPv4 = ipv4.String()
	}
	if config.UpdateIpv6 && ipv6 != nil {
		newData.IPv6 = ipv6.String()
	}

	// Check if IPs changed
	if prev.IPv4 == newData.IPv4 && prev.IPv6 == newData.IPv6 {
		fmt.Println("No IP change detected. Exiting.")
		return
	}

	// Select registrar implementation
	var registrar lib.Registrar
	switch config.Registrar {
	case "strato":
		registrar = &lib.StratoRegistrar{}
	default:
		fmt.Println("Registrar not supported:", config.Registrar)
		os.Exit(1)
	}

	// Update DNS using registrar
	err = registrar.UpdateDNS(ctx, config, newData.IPv4, newData.IPv6)
	if err != nil {
		fmt.Println("DNS update error:", err)
	}

	// Store new IPs
	err = lib.StoreIPData(storagePath, newData)
	if err != nil {
		fmt.Println("Failed to update storage file:", err)
	}
}
