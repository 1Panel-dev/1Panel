package webhook_sender

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	RequestTimeout           = 10 * time.Second
	MaxResponseBodyBytes     = 64 * 1024
	MaxCapturedResponseBytes = 2 * 1024
)

var loadWebhookSystemCertPool = x509.SystemCertPool

type Preset string

const (
	PresetGeneric Preset = "generic"
	PresetSlack   Preset = "slack"
	PresetDiscord Preset = "discord"
	PresetTeams   Preset = "teams"
)

type BodyFormat string

const (
	BodyJSON BodyFormat = "json"
	BodyForm BodyFormat = "form"
	BodyText BodyFormat = "text"
)

type Request struct {
	URL             string
	Preset          Preset
	Format          BodyFormat
	Body            []byte
	Headers         map[string]string
	Transport       *http.Transport
	Resolver        IPResolver
	CaptureResponse bool
}

type IPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type Result struct {
	StatusCode   int
	ResponseSize int
	Duration     time.Duration
	Response     string
}

func Execute(ctx context.Context, input Request) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	contentType, err := validateRenderedBody(input.Format, input.Body)
	if err != nil {
		return Result{}, err
	}
	if err := validateHeaders(input.Headers); err != nil {
		return Result{}, err
	}
	requestContext, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()

	target, err := prepareTarget(requestContext, input.URL, input.Preset, input.Resolver)
	if err != nil {
		return Result{}, errors.New("invalid webhook request URL")
	}

	req, err := http.NewRequestWithContext(requestContext, http.MethodPost, target.URL, bytes.NewReader(input.Body))
	if err != nil {
		return Result{}, errors.New("create webhook request failed")
	}
	req.Host = target.HostHeader
	for name, value := range input.Headers {
		req.Header.Set(name, value)
	}

	req.Header.Set("Content-Type", contentType)

	transport := systemTLSTransport(input.Transport, target.ServerName, target.OriginalURL, target.Addresses)
	var roundTripper http.RoundTripper = transport
	if target.OriginalURL != nil && target.OriginalURL.Scheme == "http" {
		roundTripper = &pinnedHTTPProxyRoundTripper{transport: transport, originalHost: target.HostHeader}
	}
	client := &http.Client{
		Timeout:   RequestTimeout,
		Transport: roundTripper,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	startedAt := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Result{Duration: duration}, errors.New("webhook request timed out")
		}
		if errors.Is(err, context.Canceled) {
			return Result{Duration: duration}, errors.New("webhook request canceled")
		}
		return Result{Duration: duration}, errors.New("webhook request failed")
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBodyBytes+1))
	duration = time.Since(startedAt)
	result := Result{
		StatusCode:   resp.StatusCode,
		ResponseSize: len(responseBody),
		Duration:     duration,
	}
	if input.CaptureResponse {
		result.Response = captureResponse(responseBody, input, target)
	}
	if err != nil {
		return result, errors.New("read webhook response failed")
	}
	if len(responseBody) > MaxResponseBodyBytes {
		return result, errors.New("webhook response exceeded size limit")
	}

	if err := validateResponse(input.Preset, resp.StatusCode, responseBody); err != nil {
		return result, err
	}
	return result, nil
}

type pinnedHTTPProxyRoundTripper struct {
	transport    *http.Transport
	originalHost string
}

func (p *pinnedHTTPProxyRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if p.transport.Proxy == nil {
		return p.transport.RoundTrip(request)
	}
	proxyURL, err := p.transport.Proxy(request)
	if err != nil {
		return nil, errors.New("select webhook proxy failed")
	}
	if proxyURL == nil {
		direct := p.transport.Clone()
		direct.Proxy = nil
		return direct.RoundTrip(request)
	}
	return p.roundTripProxy(request, proxyURL)
}

func (p *pinnedHTTPProxyRoundTripper) roundTripProxy(request *http.Request, proxyURL *url.URL) (*http.Response, error) {
	proxyAddress, err := proxyDialAddress(proxyURL)
	if err != nil {
		return nil, err
	}
	dialContext := p.transport.DialContext
	if dialContext == nil {
		dialContext = (&net.Dialer{}).DialContext
	}
	connection, err := dialContext(request.Context(), "tcp", proxyAddress)
	if err != nil {
		return nil, errors.New("connect webhook proxy failed")
	}
	closeConnection := true
	defer func() {
		if closeConnection {
			_ = connection.Close()
		}
	}()
	if deadline, ok := request.Context().Deadline(); ok {
		_ = connection.SetDeadline(deadline)
	}
	if proxyURL.Scheme == "https" {
		proxyTLSConfig := &tls.Config{ServerName: proxyURL.Hostname()}
		if roots, rootsErr := loadWebhookSystemCertPool(); rootsErr == nil {
			proxyTLSConfig.RootCAs = roots
		}
		tlsConnection := tls.Client(connection, proxyTLSConfig)
		if err := tlsConnection.HandshakeContext(request.Context()); err != nil {
			return nil, errors.New("connect webhook proxy failed")
		}
		connection = tlsConnection
	}

	connectAuthority, err := pinnedHTTPConnectAuthority(request.URL)
	if err != nil {
		return nil, err
	}
	connectRequest := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: connectAuthority},
		Host:   connectAuthority,
		Header: make(http.Header),
	}
	if err := p.addProxyHeaders(connectRequest, proxyURL); err != nil {
		return nil, err
	}
	if err := connectRequest.Write(connection); err != nil {
		return nil, errors.New("connect webhook proxy failed")
	}
	reader := bufio.NewReader(connection)
	connectResponse, err := http.ReadResponse(reader, connectRequest)
	if err != nil {
		return nil, errors.New("connect webhook proxy failed")
	}
	if p.transport.OnProxyConnectResponse != nil {
		if err := p.transport.OnProxyConnectResponse(request.Context(), proxyURL, connectRequest, connectResponse); err != nil {
			return nil, errors.New("connect webhook proxy failed")
		}
	}
	if connectResponse.StatusCode != http.StatusOK {
		return nil, errors.New("connect webhook proxy failed")
	}
	_ = connectResponse.Body.Close()

	wireRequest := request.Clone(request.Context())
	wireRequest.URL = cloneURL(request.URL)
	wireRequest.Host = p.originalHost
	wireRequest.Close = true
	wireRequest.Header = request.Header.Clone()
	if request.GetBody != nil {
		body, bodyErr := request.GetBody()
		if bodyErr != nil {
			return nil, errors.New("prepare webhook proxy request failed")
		}
		wireRequest.Body = body
		defer body.Close()
	}
	if err := wireRequest.Write(connection); err != nil {
		return nil, errors.New("send webhook proxy request failed")
	}
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return nil, errors.New("read webhook proxy response failed")
	}
	response.Body = &proxyConnectionBody{ReadCloser: response.Body, connection: connection}
	closeConnection = false
	return response, nil
}

func (p *pinnedHTTPProxyRoundTripper) addProxyHeaders(request *http.Request, proxyURL *url.URL) error {
	for name, values := range p.transport.ProxyConnectHeader {
		request.Header[name] = append([]string(nil), values...)
	}
	if p.transport.GetProxyConnectHeader != nil {
		target := request.URL.Host
		if target == "" {
			target = request.Host
		}
		headers, err := p.transport.GetProxyConnectHeader(request.Context(), proxyURL, target)
		if err != nil {
			return errors.New("prepare webhook proxy request failed")
		}
		for name, values := range headers {
			request.Header[name] = append([]string(nil), values...)
		}
	}
	if proxyURL.User != nil && request.Header.Get("Proxy-Authorization") == "" {
		password, _ := proxyURL.User.Password()
		credential := base64.StdEncoding.EncodeToString([]byte(proxyURL.User.Username() + ":" + password))
		request.Header.Set("Proxy-Authorization", "Basic "+credential)
	}
	return nil
}

func pinnedHTTPConnectAuthority(target *url.URL) (string, error) {
	if target == nil || target.Hostname() == "" {
		return "", errors.New("prepare webhook proxy request failed")
	}
	port := target.Port()
	if port == "" {
		port = "80"
	}
	return net.JoinHostPort(target.Hostname(), port), nil
}

func proxyDialAddress(proxyURL *url.URL) (string, error) {
	if proxyURL == nil || (proxyURL.Scheme != "http" && proxyURL.Scheme != "https") || proxyURL.Hostname() == "" {
		return "", errors.New("unsupported webhook proxy")
	}
	port := proxyURL.Port()
	if port == "" {
		if proxyURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return net.JoinHostPort(proxyURL.Hostname(), port), nil
}

func cloneURL(source *url.URL) *url.URL {
	if source == nil {
		return &url.URL{}
	}
	cloned := *source
	return &cloned
}

type proxyConnectionBody struct {
	io.ReadCloser
	connection net.Conn
}

func (b *proxyConnectionBody) Close() error {
	bodyErr := b.ReadCloser.Close()
	connectionErr := b.connection.Close()
	if bodyErr != nil {
		return bodyErr
	}
	return connectionErr
}

func captureResponse(body []byte, input Request, target preparedTarget) string {
	response := strings.ToValidUTF8(string(body), "\uFFFD")
	redactions := make([]string, 0, len(input.Headers)+8)
	redactions = append(redactions, urlRedactionValues(input.URL)...)
	redactions = append(redactions, urlRedactionValues(target.URL)...)
	if target.OriginalURL != nil {
		redactions = append(redactions, urlRedactionValues(target.OriginalURL.String())...)
	}
	for name, value := range input.Headers {
		if value != "" {
			redactions = append(redactions, value)
		}
		redactions = append(redactions, headerRedactionValues(name, value)...)
	}
	sort.SliceStable(redactions, func(i, j int) bool {
		return len(redactions[i]) > len(redactions[j])
	})
	for _, value := range redactions {
		if value != "" {
			response = strings.ReplaceAll(response, value, "[REDACTED]")
		}
	}
	return truncateUTF8(response, MaxCapturedResponseBytes)
}

func urlRedactionValues(rawURL string) []string {
	trimmed := strings.TrimSpace(rawURL)
	values := []string{trimmed}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return values
	}
	values = append(values, parsed.String())
	for _, segment := range strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/") {
		if segment == "" {
			continue
		}
		values = append(values, segment)
		if decoded, err := url.PathUnescape(segment); err == nil && decoded != segment {
			values = append(values, decoded)
		}
	}
	for _, pair := range strings.Split(parsed.RawQuery, "&") {
		if pair == "" {
			continue
		}
		encodedName, encodedValue, found := strings.Cut(pair, "=")
		if !found {
			values = append(values, encodedName)
			if decoded, err := url.QueryUnescape(encodedName); err == nil && decoded != encodedName {
				values = append(values, decoded)
			}
			continue
		}
		if encodedValue == "" {
			continue
		}
		values = append(values, encodedValue)
		if decoded, err := url.QueryUnescape(encodedValue); err == nil && decoded != encodedValue {
			values = append(values, decoded)
		}
	}
	return values
}

func headerRedactionValues(name, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	var values []string
	switch http.CanonicalHeaderKey(name) {
	case "Authorization", "Proxy-Authorization":
		if _, credential, found := strings.Cut(trimmed, " "); found {
			credential = strings.TrimSpace(credential)
			if credential != "" {
				values = append(values, credential)
			}
		}
	case "Cookie":
		for _, cookie := range strings.Split(trimmed, ";") {
			_, cookieValue, found := strings.Cut(cookie, "=")
			if !found {
				continue
			}
			cookieValue = strings.Trim(strings.TrimSpace(cookieValue), `"`)
			if cookieValue != "" {
				values = append(values, cookieValue)
			}
		}
	}
	return values
}

func truncateUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	value = value[:maxBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

type preparedTarget struct {
	URL         string
	HostHeader  string
	ServerName  string
	OriginalURL *url.URL
	Addresses   []net.IP
}

func prepareTarget(ctx context.Context, rawURL string, preset Preset, resolver IPResolver) (preparedTarget, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return preparedTarget{}, errors.New("invalid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return preparedTarget{}, errors.New("unsupported URL scheme")
	}
	if preset == PresetDiscord {
		query := parsed.Query()
		query.Set("wait", "true")
		parsed.RawQuery = query.Encode()
	}

	hostname := strings.TrimSuffix(parsed.Hostname(), ".")
	if hostname == "" || strings.Contains(hostname, "%") {
		return preparedTarget{}, errors.New("invalid URL host")
	}
	addresses, err := resolveAddresses(ctx, hostname, resolver)
	if err != nil || len(addresses) == 0 {
		return preparedTarget{}, errors.New("resolve webhook URL failed")
	}
	for _, address := range addresses {
		if !publicWebhookIP(address) {
			return preparedTarget{}, errors.New("webhook URL resolved to a blocked address")
		}
	}

	originalHost := parsed.Host
	originalURL := *parsed
	parsed.Host = pinnedHost(addresses[0], parsed.Port())
	return preparedTarget{
		URL:         parsed.String(),
		HostHeader:  originalHost,
		ServerName:  hostname,
		OriginalURL: &originalURL,
		Addresses:   addresses,
	}, nil
}

func resolveAddresses(ctx context.Context, hostname string, resolver IPResolver) ([]net.IP, error) {
	if literal := net.ParseIP(hostname); literal != nil {
		return []net.IP{literal}, nil
	}
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	resolved, err := resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return nil, err
	}
	addresses := make([]net.IP, 0, len(resolved))
	for _, item := range resolved {
		if item.IP != nil {
			addresses = append(addresses, item.IP)
		}
	}
	return addresses, nil
}

func publicWebhookIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	if !ip.IsGlobalUnicast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return false
	}
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	if wellKnownNAT64Prefix.Contains(address) {
		value := address.As16()
		return publicWebhookIP(net.IPv4(value[12], value[13], value[14], value[15]))
	}
	for _, prefix := range globallyReachableSpecialPrefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	for _, prefix := range blockedWebhookPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

var wellKnownNAT64Prefix = netip.MustParsePrefix("64:ff9b::/96")

var globallyReachableSpecialPrefixes = []netip.Prefix{
	netip.MustParsePrefix("192.0.0.9/32"),
	netip.MustParsePrefix("192.0.0.10/32"),
	netip.MustParsePrefix("2001:1::1/128"),
	netip.MustParsePrefix("2001:1::2/128"),
	netip.MustParsePrefix("2001:1::3/128"),
	netip.MustParsePrefix("2001:3::/32"),
	netip.MustParsePrefix("2001:4:112::/48"),
	netip.MustParsePrefix("2001:20::/28"),
	netip.MustParsePrefix("2001:30::/28"),
}

var blockedWebhookPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("100:0:0:1::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("3fff::/20"),
	netip.MustParsePrefix("5f00::/16"),
	netip.MustParsePrefix("fec0::/10"),
}

func pinnedHost(ip net.IP, port string) string {
	host := ip.String()
	if port != "" {
		return net.JoinHostPort(host, port)
	}
	if ip.To4() == nil {
		return "[" + host + "]"
	}
	return host
}

func validateRenderedBody(format BodyFormat, body []byte) (string, error) {
	if format == "" {
		format = BodyJSON
	}
	switch format {
	case BodyJSON:
		if !json.Valid(body) {
			return "", errors.New("rendered webhook JSON body is invalid")
		}
		return "application/json; charset=utf-8", nil
	case BodyForm:
		return "application/x-www-form-urlencoded", nil
	case BodyText:
		return "text/plain; charset=utf-8", nil
	default:
		return "", errors.New("unsupported webhook body format")
	}
}

func validateHeaders(headers map[string]string) error {
	for name, value := range headers {
		if !validHeaderName(name) || strings.ContainsAny(value, "\r\n") || reservedHeader(name) {
			return errors.New("invalid webhook request header")
		}
	}
	return nil
}

func validHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		if !headerTokenByte(name[i]) {
			return false
		}
	}
	return true
}

func headerTokenByte(b byte) bool {
	if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' {
		return true
	}
	return strings.ContainsRune("!#$%&'*+-.^_`|~", rune(b))
}

func reservedHeader(name string) bool {
	switch http.CanonicalHeaderKey(name) {
	case "Connection", "Content-Length", "Content-Type", "Host", "Proxy-Authorization", "Proxy-Connection", "Te", "Trailer", "Transfer-Encoding", "Upgrade":
		return true
	default:
		return false
	}
}

func systemTLSTransport(input *http.Transport, serverName string, originalURL *url.URL, addresses []net.IP) *http.Transport {
	var transport *http.Transport
	if input != nil {
		transport = input.Clone()
	} else if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = defaultTransport.Clone()
	} else {
		transport = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}

	tlsConfig := &tls.Config{}
	if transport.TLSClientConfig != nil {
		tlsConfig = transport.TLSClientConfig.Clone()
	}
	tlsConfig.InsecureSkipVerify = false
	if roots, err := loadWebhookSystemCertPool(); err == nil {
		tlsConfig.RootCAs = roots
	} else {
		tlsConfig.RootCAs = nil
	}
	tlsConfig.Certificates = nil
	tlsConfig.GetClientCertificate = nil
	tlsConfig.ServerName = serverName
	transport.TLSClientConfig = tlsConfig
	transport.DialTLS = nil
	transport.DialTLSContext = nil
	if transport.Proxy != nil && originalURL != nil {
		proxySelector := transport.Proxy
		selectionURL := *originalURL
		transport.Proxy = func(request *http.Request) (*url.URL, error) {
			selectionRequest := request.Clone(request.Context())
			selectionRequest.URL = &selectionURL
			selectionRequest.Host = selectionURL.Host
			return proxySelector(selectionRequest)
		}
	}
	if len(addresses) > 1 {
		firstAddress := addresses[0]
		baseDial := transport.DialContext
		if baseDial == nil {
			baseDial = (&net.Dialer{}).DialContext
		}
		transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			dialIP := net.ParseIP(host)
			if err != nil || dialIP == nil || !dialIP.Equal(firstAddress) {
				return baseDial(ctx, network, address)
			}
			var lastErr error
			for _, candidate := range addresses {
				connection, dialErr := baseDial(ctx, network, net.JoinHostPort(candidate.String(), port))
				if dialErr == nil {
					return connection, nil
				}
				lastErr = dialErr
			}
			return nil, lastErr
		}
	}
	secureDial := transport.DialContext
	if secureDial == nil {
		secureDial = (&net.Dialer{}).DialContext
	}
	targetPort := "443"
	if originalURL != nil && originalURL.Port() != "" {
		targetPort = originalURL.Port()
	}
	transport.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		connection, err := secureDial(ctx, network, address)
		if err != nil {
			return nil, err
		}
		host, port, splitErr := net.SplitHostPort(address)
		if splitErr != nil {
			_ = connection.Close()
			return nil, splitErr
		}
		firstHopConfig := tlsConfig.Clone()
		if port == targetPort && isPinnedWebhookAddress(host, addresses) {
			firstHopConfig.ServerName = serverName
		} else {
			firstHopConfig.ServerName = strings.TrimSuffix(host, ".")
			firstHopConfig.NextProtos = []string{"http/1.1"}
		}
		tlsConnection := tls.Client(connection, firstHopConfig)
		if err := tlsConnection.HandshakeContext(ctx); err != nil {
			_ = connection.Close()
			return nil, err
		}
		return tlsConnection, nil
	}
	return transport
}

func isPinnedWebhookAddress(host string, addresses []net.IP) bool {
	dialIP := net.ParseIP(host)
	if dialIP == nil {
		return false
	}
	for _, address := range addresses {
		if dialIP.Equal(address) {
			return true
		}
	}
	return false
}

func validateResponse(preset Preset, statusCode int, body []byte) error {
	if preset == "" {
		preset = PresetGeneric
	}
	switch preset {
	case PresetSlack:
		if statusCode != http.StatusOK || strings.TrimSpace(string(body)) != "ok" {
			return errors.New("webhook provider rejected response")
		}
		return nil
	case PresetGeneric, PresetDiscord, PresetTeams:
		if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
			return errors.New("webhook provider rejected response")
		}
		return nil
	default:
		return errors.New("unsupported webhook preset")
	}
}
