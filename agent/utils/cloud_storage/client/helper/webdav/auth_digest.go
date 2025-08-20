package webdav

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DigestAuth struct {
	user        string
	pw          string
	digestParts map[string]string
}

func NewDigestAuth(login, secret string, rs *http.Response) (Authenticator, error) {
	return &DigestAuth{user: login, pw: secret, digestParts: digestParts(rs)}, nil
}

func (d *DigestAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	d.digestParts["uri"] = path
	d.digestParts["method"] = rq.Method
	d.digestParts["username"] = d.user
	d.digestParts["password"] = d.pw
	rq.Header.Set("Authorization", getDigestAuthorization(d.digestParts))
	return nil
}

func (d *DigestAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if rs.StatusCode == 401 {
		if isStaled(rs) {
			redo = true
			err = ErrAuthChanged
		} else {
			err = NewPathError("Authorize", path, rs.StatusCode)
		}
	}
	return
}

func (d *DigestAuth) Close() error {
	return nil
}

func (d *DigestAuth) Clone() Authenticator {
	parts := make(map[string]string, len(d.digestParts))
	for k, v := range d.digestParts {
		parts[k] = v
	}
	return &DigestAuth{user: d.user, pw: d.pw, digestParts: parts}
}

func (d *DigestAuth) String() string {
	return fmt.Sprintf("DigestAuth login: %s", d.user)
}

func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop", "opaque", "algorithm", "entityBody"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Trim(
						strings.SplitN(r, `=`, 2)[1],
						`"`,
					)
				}
			}
		}
	}
	return result
}

func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthorization(digestParts map[string]string) string {
	d := digestParts

	var (
		ha1        string
		ha2        string
		nonceCount = 00000001
		cnonce     = getCnonce()
		response   string
	)

	switch d["algorithm"] {
	case "MD5", "":
		ha1 = getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	case "MD5-sess":
		ha1 = getMD5(
			fmt.Sprintf("%s:%v:%s",
				getMD5(d["username"]+":"+d["realm"]+":"+d["password"]),
				nonceCount,
				cnonce,
			),
		)
	}

	switch d["qop"] {
	case "auth", "":
		ha2 = getMD5(d["method"] + ":" + d["uri"])
	case "auth-int":
		if d["entityBody"] != "" {
			ha2 = getMD5(d["method"] + ":" + d["uri"] + ":" + getMD5(d["entityBody"]))
		}
	}

	switch d["qop"] {
	case "":
		response = getMD5(
			fmt.Sprintf("%s:%s:%s",
				ha1,
				d["nonce"],
				ha2,
			),
		)
	case "auth", "auth-int":
		response = getMD5(
			fmt.Sprintf("%s:%s:%v:%s:%s:%s",
				ha1,
				d["nonce"],
				nonceCount,
				cnonce,
				d["qop"],
				ha2,
			),
		)
	}

	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", nc=%v, cnonce="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], nonceCount, cnonce, response)

	if d["qop"] != "" {
		authorization += fmt.Sprintf(`, qop=%s`, d["qop"])
	}

	if d["opaque"] != "" {
		authorization += fmt.Sprintf(`, opaque="%s"`, d["opaque"])
	}

	return authorization
}

func isStaled(rs *http.Response) bool {
	header := rs.Header.Get("Www-Authenticate")
	if len(header) > 0 {
		directives := strings.Split(header, ",")
		for i := range directives {
			name, value, _ := strings.Cut(strings.Trim(directives[i], " "), "=")
			if strings.EqualFold(name, "stale") {
				return strings.EqualFold(value, "true")
			}
		}
	}
	return false
}
