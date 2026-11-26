package clusterarchive

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"math"

	"golang.org/x/crypto/scrypt"
)

const (
	archiveMagic = "1PanelHA1"
	saltSize     = 16
	noncePrefix  = 8
	frameSize    = 256 << 10
	headerSize   = len(archiveMagic) + saltSize + noncePrefix
)

type seal struct {
	aead    cipher.AEAD
	nonce   [12]byte
	counter uint32
}

func newSeal(header []byte, password string) (seal, error) {
	if password == "" {
		return seal{}, ErrPassphraseRequired
	}
	key, err := scrypt.Key([]byte(password), header[len(archiveMagic):len(archiveMagic)+saltSize], 1<<15, 8, 1, 32)
	if err != nil {
		return seal{}, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return seal{}, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return seal{}, err
	}
	s := seal{aead: aead}
	copy(s.nonce[:noncePrefix], header[len(archiveMagic)+saltSize:])
	return s, nil
}

func (s *seal) nextNonce() ([]byte, error) {
	if s.counter == math.MaxUint32 {
		return nil, ErrTooLarge
	}
	binary.BigEndian.PutUint32(s.nonce[noncePrefix:], s.counter)
	s.counter++
	return s.nonce[:], nil
}

type sealWriter struct {
	seal
	dst io.Writer
}

func newSealWriter(dst io.Writer, password string) (*sealWriter, error) {
	header := make([]byte, headerSize)
	copy(header, archiveMagic)
	if _, err := rand.Read(header[len(archiveMagic):]); err != nil {
		return nil, err
	}
	s, err := newSeal(header, password)
	if err != nil {
		return nil, err
	}
	if err := writeSealBytes(dst, header); err != nil {
		return nil, err
	}
	return &sealWriter{seal: s, dst: dst}, nil
}

func (w *sealWriter) frame(p []byte) error {
	nonce, err := w.nextNonce()
	if err != nil {
		return err
	}
	encrypted := w.aead.Seal(nil, nonce, p, nil)
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(encrypted)))
	if err := writeSealBytes(w.dst, size[:]); err != nil {
		return err
	}
	return writeSealBytes(w.dst, encrypted)
}

func writeSealBytes(dst io.Writer, p []byte) error {
	n, err := dst.Write(p)
	if n != len(p) && err == nil {
		return io.ErrShortWrite
	}
	return err
}

func (w *sealWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		size := min(len(p), frameSize)
		if err := w.frame(p[:size]); err != nil {
			return written, err
		}
		written += size
		p = p[size:]
	}
	return written, nil
}

func (w *sealWriter) Close() error {
	return w.frame(nil)
}

type sealReader struct {
	seal
	src   io.Reader
	plain []byte
	done  bool
}

func newSealReader(src io.Reader, password string) (*sealReader, error) {
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(src, header); err != nil || string(header[:len(archiveMagic)]) != archiveMagic {
		return nil, ErrFormat
	}
	s, err := newSeal(header, password)
	if err != nil {
		return nil, err
	}
	return &sealReader{seal: s, src: src}, nil
}

func (r *sealReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	if len(r.plain) > 0 {
		n := copy(p, r.plain)
		r.plain = r.plain[n:]
		return n, nil
	}
	var length [4]byte
	if _, err := io.ReadFull(r.src, length[:]); err != nil {
		return 0, ErrTruncated
	}
	size := int(binary.BigEndian.Uint32(length[:]))
	if size < r.aead.Overhead() || size > frameSize+r.aead.Overhead() {
		return 0, ErrBlock
	}
	encoded := make([]byte, size)
	if _, err := io.ReadFull(r.src, encoded); err != nil {
		return 0, ErrTruncated
	}
	nonce, err := r.nextNonce()
	if err != nil {
		return 0, err
	}
	plain, err := r.aead.Open(encoded[:0], nonce, encoded, nil)
	if err != nil {
		return 0, ErrAuthentication
	}
	if len(plain) == 0 {
		var trailing [1]byte
		if n, err := io.ReadFull(r.src, trailing[:]); n != 0 || err != io.EOF {
			return 0, ErrTrailingData
		}
		r.done = true
		return 0, io.EOF
	}
	n := copy(p, plain)
	r.plain = plain[n:]
	return n, nil
}
