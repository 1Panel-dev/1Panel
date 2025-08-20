package webdav

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const XInhibitRedirect = "X-Gowebdav-Inhibit-Redirect"
const template = `<d:propfind xmlns:d='DAV:'>
<d:prop>
	<d:displayname/>
	<d:resourcetype/>
	<d:getcontentlength/>
</d:prop>
</d:propfind>`

type Client struct {
	root    string
	headers http.Header
	c       *http.Client
	auth    Authorizer
}

func NewClient(uri, user, pw string) *Client {
	return NewAuthClient(uri, NewAutoAuth(user, pw))
}

func NewAuthClient(uri string, auth Authorizer) *Client {
	c := &http.Client{
		CheckRedirect: func(rq *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return ErrTooManyRedirects
			}
			if via[0].Header.Get(XInhibitRedirect) != "" {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	return &Client{root: FixSlash(uri), headers: make(http.Header), c: c, auth: auth}
}

func (c *Client) SetTransport(transport http.RoundTripper) {
	c.c.Transport = transport
}

func (c *Client) Connect() error {
	rs, err := c.req("OPTIONS", "/", nil, func(rq *http.Request) { rq.Header.Add("Depth", "0") })
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	if rs.StatusCode != 200 && rs.StatusCode != 204 {
		return fmt.Errorf("check conn failed, code: %d, err: %v", rs.StatusCode, rs.Status)
	}

	return nil
}

type props struct {
	Status string   `xml:"DAV: status"`
	Name   string   `xml:"DAV: prop>displayname,omitempty"`
	Type   xml.Name `xml:"DAV: prop>resourcetype>collection,omitempty"`
	Size   string   `xml:"DAV: prop>getcontentlength,omitempty"`
}

type response struct {
	Href  string  `xml:"DAV: href"`
	Props []props `xml:"DAV: propstat"`
}

func getProps(r *response, status string) *props {
	for _, prop := range r.Props {
		if strings.Contains(prop.Status, status) {
			return &prop
		}
	}
	return nil
}

func (c *Client) ReadDir(path string) ([]os.FileInfo, error) {
	path = FixSlashes(path)
	files := make([]os.FileInfo, 0)
	skipSelf := true
	parse := func(resp interface{}) error {
		r := resp.(*response)

		if skipSelf {
			skipSelf = false
			if p := getProps(r, "200"); p != nil && p.Type.Local == "collection" {
				r.Props = nil
				return nil
			}
			return NewPathError("ReadDir", path, 405)
		}

		if p := getProps(r, "200"); p != nil {
			f := new(File)
			if ps, err := url.PathUnescape(r.Href); err == nil {
				f.name = filepath.Base(ps)
			} else {
				f.name = p.Name
			}
			f.path = path + f.name
			if p.Type.Local == "collection" {
				f.path += "/"
				f.size = 0
				f.isdir = true
			} else {
				f.size = parseInt64(&p.Size)
				f.isdir = false
			}

			files = append(files, *f)
		}

		r.Props = nil
		return nil
	}

	if err := c.propfind(path, false, template, &response{}, parse); err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return files, fmt.Errorf("load files from %s failed, err: %v", path, err)
		}
	}
	return files, nil
}

func (c *Client) Stat(path string) (os.FileInfo, error) {
	var f *File
	parse := func(resp interface{}) error {
		r := resp.(*response)
		if p := getProps(r, "200"); p != nil && f == nil {
			f = new(File)
			f.name = p.Name
			f.path = path

			if p.Type.Local == "collection" {
				if !strings.HasSuffix(f.path, "/") {
					f.path += "/"
				}
				f.size = 0
				f.isdir = true
			} else {
				f.size = parseInt64(&p.Size)
				f.isdir = false
			}
		}

		r.Props = nil
		return nil
	}

	if err := c.propfind(path, true, template, &response{}, parse); err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return f, fmt.Errorf("load file %s failed, path err: %v", path, err)
		}
		return f, fmt.Errorf("load file %s failed, err: %v", path, err)
	}
	return f, nil
}

func (c *Client) RemoveAll(path string) error {
	rs, err := c.req("DELETE", path, nil, nil)
	if err != nil {
		return fmt.Errorf("handle remove file %s failed, err: %s", path, err)
	}
	defer rs.Body.Close()
	if rs.StatusCode == 200 || rs.StatusCode == 204 || rs.StatusCode == 404 {
		return nil
	}
	return fmt.Errorf("handle remove file %s failed, code: %d, err: %s", path, rs.StatusCode, rs.Status)
}

func (c *Client) MkdirAll(path string, _ os.FileMode) (err error) {
	parentPath := filepath.Dir(path)
	if parentPath == "." || parentPath == "/" {
		return nil
	}

	paths := SplitPathToHierarchy(parentPath)
	for _, item := range paths {
		itemFile, err := c.Stat(item)
		if err == nil && itemFile.IsDir() {
			continue
		}
		rs, err := c.req("MKCOL", item, nil, nil)
		if err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", item, err)
		}
		defer rs.Body.Close()
		if rs.StatusCode != 201 || rs.StatusCode == 200 {
			return fmt.Errorf("mkdir %s failed, code: %d,  err: %v", item, rs.StatusCode, rs.Status)
		}
	}
	return nil
}

func (c *Client) ReadStream(path string) (io.ReadCloser, error) {
	rs, err := c.req("GET", path, nil, nil)
	if err != nil {
		return nil, NewPathErrorErr("ReadStream", path, err)
	}

	if rs.StatusCode == 200 {
		return rs.Body, nil
	}

	rs.Body.Close()
	return nil, NewPathError("ReadStream", path, rs.StatusCode)
}

func (c *Client) WriteStream(path string, stream io.Reader, _ os.FileMode) (err error) {
	err = c.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	contentLength := int64(0)
	if seeker, ok := stream.(io.Seeker); ok {
		contentLength, err = seeker.Seek(0, io.SeekEnd)
		if err != nil {
			return err
		}

		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
	} else {
		buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024 /* 1MB */))

		contentLength, err = io.Copy(buffer, stream)
		if err != nil {
			return err
		}

		stream = buffer
	}

	s, err := c.put(path, stream, contentLength)
	if err != nil {
		return err
	}

	switch s {
	case 200, 201, 204:
		return nil

	default:
		return NewPathError("WriteStream", path, s)
	}
}
