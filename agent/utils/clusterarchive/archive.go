package clusterarchive

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

const maxManifestSize = 16 << 10

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrPassphraseRequired Error = "ErrClusterArchivePassphraseRequired"
	ErrTooLarge           Error = "ErrClusterArchiveTooLarge"
	ErrFormat             Error = "ErrClusterArchiveFormat"
	ErrTruncated          Error = "ErrClusterArchiveTruncated"
	ErrBlock              Error = "ErrClusterArchiveBlock"
	ErrAuthentication     Error = "ErrClusterArchiveAuthentication"
	ErrTrailingData       Error = "ErrClusterArchiveTrailingData"
	ErrManifest           Error = "ErrClusterArchiveManifest"
	ErrManifestTooLarge   Error = "ErrClusterArchiveManifestTooLarge"
	ErrInvalid            Error = "ErrClusterArchiveInvalid"
	ErrPayload            Error = "ErrClusterArchivePayload"
	ErrChecksum           Error = "ErrClusterArchiveChecksum"
)

type Manifest struct {
	Schema        int       `json:"schema"`
	App           string    `json:"app"`
	Template      string    `json:"template"`
	EngineVersion string    `json:"engineVersion"`
	Format        string    `json:"format"`
	CreatedAt     time.Time `json:"createdAt"`
	Size          int64     `json:"size"`
	Digest        string    `json:"sha256"`
}

func (m Manifest) validate() error {
	formats := map[string]string{"mysql-cluster": "sql", "postgresql-cluster": "basebackup", "redis-cluster": "rdb"}
	digest, err := hex.DecodeString(m.Digest)
	if m.Schema != 1 || formats[m.App] != m.Format || m.Format == "" || !strings.HasSuffix(m.Template, "-master") ||
		m.EngineVersion == "" || m.CreatedAt.IsZero() || m.Size <= 0 || err != nil || len(digest) != sha256.Size {
		return ErrManifest
	}
	return nil
}

func Pack(dst io.Writer, payload string, manifest Manifest, password string) (Manifest, error) {
	file, err := os.Open(payload)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()
	sum := sha256.New()
	size, err := io.Copy(sum, file)
	if err != nil {
		return Manifest{}, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return Manifest{}, err
	}
	manifest.Schema, manifest.Size, manifest.Digest = 1, size, hex.EncodeToString(sum.Sum(nil))
	if err := manifest.validate(); err != nil {
		return Manifest{}, err
	}
	meta, err := json.Marshal(manifest)
	if err != nil {
		return Manifest{}, err
	}
	if len(meta) > maxManifestSize {
		return Manifest{}, ErrManifestTooLarge
	}
	sealed, err := newSealWriter(dst, password)
	if err != nil {
		return Manifest{}, err
	}
	zip := gzip.NewWriter(sealed)
	tr := tar.NewWriter(zip)
	if err := tr.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0600, Size: int64(len(meta))}); err != nil {
		return Manifest{}, err
	}
	if _, err := tr.Write(meta); err != nil {
		return Manifest{}, err
	}
	if err := tr.WriteHeader(&tar.Header{Name: "payload", Mode: 0600, Size: size}); err != nil {
		return Manifest{}, err
	}
	if _, err := io.Copy(tr, file); err != nil {
		return Manifest{}, err
	}
	if err := tr.Close(); err != nil {
		return Manifest{}, err
	}
	if err := zip.Close(); err != nil {
		return Manifest{}, err
	}
	if err := sealed.Close(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func Unpack(src io.Reader, dst io.Writer, password string) (Manifest, error) {
	sealed, err := newSealReader(src, password)
	if err != nil {
		return Manifest{}, err
	}
	raw := bufio.NewReader(sealed)
	zip, err := gzip.NewReader(raw)
	if err != nil {
		return Manifest{}, archiveReadError(err, ErrInvalid)
	}
	defer zip.Close()
	zip.Multistream(false)
	tr := tar.NewReader(zip)
	header, err := tr.Next()
	if err != nil || header.Typeflag != tar.TypeReg || header.Name != "manifest.json" || header.Size <= 0 || header.Size > maxManifestSize {
		return Manifest{}, archiveReadError(err, ErrInvalid)
	}
	manifestData, err := io.ReadAll(tr)
	if err != nil {
		return Manifest{}, archiveReadError(err, ErrManifest)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return Manifest{}, ErrManifest
	}
	if err := manifest.validate(); err != nil {
		return Manifest{}, err
	}
	header, err = tr.Next()
	if err != nil || header.Typeflag != tar.TypeReg || header.Name != "payload" || header.Size != manifest.Size {
		return Manifest{}, archiveReadError(err, ErrPayload)
	}
	sum := sha256.New()
	n, err := io.Copy(io.MultiWriter(dst, sum), tr)
	if err != nil {
		return Manifest{}, err
	}
	if n != manifest.Size || hex.EncodeToString(sum.Sum(nil)) != manifest.Digest {
		return Manifest{}, ErrChecksum
	}
	if _, err := tr.Next(); err != io.EOF {
		return Manifest{}, archiveReadError(err, ErrInvalid)
	}
	if n, err := io.Copy(io.Discard, zip); n != 0 || err != nil {
		return Manifest{}, archiveReadError(err, ErrInvalid)
	}
	if _, err := raw.ReadByte(); err != io.EOF {
		return Manifest{}, archiveReadError(err, ErrInvalid)
	}
	return manifest, nil
}

func archiveReadError(err error, fallback Error) error {
	var archiveErr Error
	if errors.As(err, &archiveErr) {
		return err
	}
	return fallback
}
