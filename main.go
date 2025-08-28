package main

import (
	"bufio"
	"context"
	"dyndns-go/lib"
	"dyndns-go/schema"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// UpdateResult holds a parsed outcome of a DynDNS v2 update call.
type UpdateResult struct {
	Status string // e.g., "good", "nochg", "badauth", "nohost", "911", etc.
	IP     string // IP echoed by the server (when present)
	Raw    string // raw response line
}

// UpdateDynDNS sends one DynDNS v2 update request.
// - endpoint: e.g., "https://members.dyndns.org/nic/update"
// - user/pass: HTTP Basic credentials
// - hostname: FQDN to update
// - myIP: set to explicit v4/v6
func UpdateDynDNS(ctx context.Context, endpoint, user, pass, hostname string, myIP net.IP) (UpdateResult, error) {
	var res UpdateResult

	if hostname == "" {
		return res, errors.New("hostname is required")
	}

	u, err := url.Parse(endpoint)

	if err != nil {
		return res, fmt.Errorf("bad endpoint: %w", err)
	}

	q := u.Query()

	q.Set("hostname", hostname)

	if ipStr := ipToParam(myIP); ipStr != "" {
		q.Set("myip", ipStr)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	if err != nil {
		return res, err
	}

	// Many providers require a reasonable User-Agent
	req.Header.Set("User-Agent", "dyndns2-go/1.0")
	req.Header.Set("Accept", "text/plain")
	req.SetBasicAuth(user, pass)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	httpRes, err := client.Do(req)

	if err != nil {
		return res, err
	}

	defer httpRes.Body.Close()

	// Body is a single short line; read the first line.
	reader := bufio.NewReader(httpRes.Body)

	line, err := reader.ReadString('\n')

	if err != nil && !errors.Is(err, bufio.ErrBufferFull) && !errors.Is(err, io.EOF) {
		return res, err
	}

	line = strings.TrimSpace(line)

	res.Raw = line

	// Parse standard dyndns2 responses.
	parseDynDNSResponse(&res)

	return res, nil
}

func parseDynDNSResponse(res *UpdateResult) {
	line := res.Raw

	// Add explainer for each status
	switch {
	case strings.HasPrefix(line, "good "):
		res.Status = "good"
		res.IP = strings.TrimSpace(line[len("good "):])
		res.Raw += " | Explainer: Update successful."
	case strings.HasPrefix(line, "nochg "):
		res.Status = "nochg"
		res.IP = strings.TrimSpace(line[len("nochg "):])
		res.Raw += " | Explainer: No change needed; IP already set."
	case line == "badauth":
		res.Status = "badauth"
		res.Raw += " | Explainer: Authentication failed."
	case line == "nohost":
		res.Status = "nohost"
		res.Raw += " | Explainer: Hostname does not exist."
	case line == "notfqdn":
		res.Status = "notfqdn"
		res.Raw += " | Explainer: Hostname is not a valid FQDN."
	case line == "numhost":
		res.Status = "numhost"
		res.Raw += " | Explainer: Too many hosts specified."
	case line == "abuse":
		res.Status = "abuse"
		res.Raw += " | Explainer: Update blocked due to abuse."
	case line == "badagent":
		res.Status = "badagent"
		res.Raw += " | Explainer: Bad user agent."
	case line == "911":
		res.Status = "911" // server problem; back off and retry later
		res.Raw += " | Explainer: Server error, try again later."
	default:
		res.Status = "unknown"
		res.Raw += " | Explainer: Unknown response."
	}
}

func ipToParam(ip net.IP) string {
	if ip == nil {
		return ""
	}

	// net.IP.String() returns canonical form for v4/v6 (no brackets)
	return ip.String()
}

// DiscoverPublicIP fetches your public IPv4 and IPv6 addresses from a JSON API response.
// It expects the response to contain reportedState.wans[0].ipv4 and reportedState.wans[0].ipv6 fields.
func DiscoverPublicIP(ctx context.Context, config *lib.Config) (ipv4 net.IP, ipv6 net.IP, err error) {
	fmt.Println("test test " + schema.UnifiSiteManagerUrl + config.UnifiSiteManagerHostId)
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

	endpoint := schema.StratoDynDnsUrl

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

	// Update DNS and storage if changed
	if config.UpdateIpv4 && newData.IPv4 != "" {
		res4, err := UpdateDynDNS(ctx, endpoint, config.User, config.Password, config.Host, ipv4)
		fmt.Println("IPv4 update:", res4.Status, res4.IP, res4.Raw, "err:", err)
	}
	if config.UpdateIpv6 && newData.IPv6 != "" {
		res6, err := UpdateDynDNS(ctx, endpoint, config.User, config.Password, config.Host, ipv6)
		fmt.Println("IPv6 update:", res6.Status, res6.IP, res6.Raw, "err:", err)
	}

	// Store new IPs
	err = lib.StoreIPData(storagePath, newData)
	if err != nil {
		fmt.Println("Failed to update storage file:", err)
	}
}
