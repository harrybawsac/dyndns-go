package lib

import (
	"bufio"
	"context"
	"errors"
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

	req.Header.Set("User-Agent", "dyndns2-go/1.0")
	req.Header.Set("Accept", "text/plain")
	req.SetBasicAuth(user, pass)

	client := &http.Client{Timeout: 15 * time.Second}
	httpRes, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer httpRes.Body.Close()

	reader := bufio.NewReader(httpRes.Body)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, bufio.ErrBufferFull) && !errors.Is(err, io.EOF) {
		return res, err
	}
	line = strings.TrimSpace(line)
	res.Raw = line
	parseDynDNSResponse(&res)
	return res, nil
}

func parseDynDNSResponse(res *UpdateResult) {
	line := res.Raw
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
	return ip.String()
}
