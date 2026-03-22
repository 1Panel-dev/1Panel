package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type ToolboxNetService struct{}

type IToolboxNetService interface {
	Ping(req dto.NetToolReq) (string, error)
	Traceroute(req dto.NetToolReq) (string, error)
	DNSLookup(req dto.NetToolReq) (string, error)
	HTTPCheck(req dto.NetToolReq) (string, error)
}

func NewIToolboxNetService() IToolboxNetService {
	return &ToolboxNetService{}
}

func (t *ToolboxNetService) Ping(req dto.NetToolReq) (string, error) {
	if err := validateHost(req.Host); err != nil {
		return "", err
	}
	count := "4"
	if req.Count > 0 && req.Count <= 20 {
		count = fmt.Sprintf("%d", req.Count)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ping", "-c", count, "-W", "5", req.Host).CombinedOutput()
	if err != nil && len(out) == 0 {
		return "", fmt.Errorf("ping failed: %v", err)
	}
	return string(out), nil
}

func (t *ToolboxNetService) Traceroute(req dto.NetToolReq) (string, error) {
	if err := validateHost(req.Host); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Try traceroute first, fall back to tracepath
	binary := "traceroute"
	if !cmd.Which("traceroute") {
		binary = "tracepath"
		if !cmd.Which("tracepath") {
			return "", fmt.Errorf("neither traceroute nor tracepath is installed")
		}
	}

	var args []string
	if binary == "traceroute" {
		args = []string{"-m", "20", "-w", "3", req.Host}
	} else {
		args = []string{"-m", "20", req.Host}
	}

	out, err := exec.CommandContext(ctx, binary, args...).CombinedOutput()
	if err != nil && len(out) == 0 {
		return "", fmt.Errorf("%s failed: %v", binary, err)
	}
	return string(out), nil
}

func (t *ToolboxNetService) DNSLookup(req dto.NetToolReq) (string, error) {
	if err := validateHost(req.Host); err != nil {
		return "", err
	}
	var result strings.Builder

	ips, err := net.LookupHost(req.Host)
	if err != nil {
		return "", fmt.Errorf("DNS lookup failed: %v", err)
	}
	result.WriteString(fmt.Sprintf("Host: %s\n\n", req.Host))
	result.WriteString("IP Addresses:\n")
	for _, ip := range ips {
		result.WriteString(fmt.Sprintf("  %s\n", ip))
	}

	cname, err := net.LookupCNAME(req.Host)
	if err == nil && cname != req.Host+"." {
		result.WriteString(fmt.Sprintf("\nCNAME: %s\n", cname))
	}

	mxs, err := net.LookupMX(req.Host)
	if err == nil && len(mxs) > 0 {
		result.WriteString("\nMX Records:\n")
		for _, mx := range mxs {
			result.WriteString(fmt.Sprintf("  %s (priority: %d)\n", mx.Host, mx.Pref))
		}
	}

	nss, err := net.LookupNS(req.Host)
	if err == nil && len(nss) > 0 {
		result.WriteString("\nNS Records:\n")
		for _, ns := range nss {
			result.WriteString(fmt.Sprintf("  %s\n", ns.Host))
		}
	}

	txts, err := net.LookupTXT(req.Host)
	if err == nil && len(txts) > 0 {
		result.WriteString("\nTXT Records:\n")
		for _, txt := range txts {
			result.WriteString(fmt.Sprintf("  %s\n", txt))
		}
	}

	return result.String(), nil
}

func (t *ToolboxNetService) HTTPCheck(req dto.NetToolReq) (string, error) {
	url := req.Host
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	start := time.Now()
	resp, err := client.Get(url)
	elapsed := time.Since(start)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("URL: %s\n", url))
	result.WriteString(fmt.Sprintf("Status: %s\n", resp.Status))
	result.WriteString(fmt.Sprintf("Response Time: %dms\n", elapsed.Milliseconds()))
	result.WriteString(fmt.Sprintf("Protocol: %s\n", resp.Proto))
	result.WriteString(fmt.Sprintf("\nHeaders:\n"))
	for key, values := range resp.Header {
		for _, value := range values {
			result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	return result.String(), nil
}

func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host is required")
	}
	if strings.ContainsAny(host, ";|&$`(){}[]!><\n\r") {
		return fmt.Errorf("invalid host: contains illegal characters")
	}
	if len(host) > 253 {
		return fmt.Errorf("host too long")
	}
	return nil
}
