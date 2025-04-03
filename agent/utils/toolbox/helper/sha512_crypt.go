package helper

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"strconv"
)

const (
	SaltLenMin    = 1
	SaltLenMax    = 16
	RoundsMin     = 1000
	RoundsMax     = 999999999
	RoundsDefault = 5000
)

var _rounds = []byte("rounds=")

func Generate(key []byte) (string, error) {
	var rounds int
	var isRoundsDef bool

	salt := generateWRounds()
	magicPrefix := []byte("$6$")
	if !bytes.HasPrefix(salt, magicPrefix) {
		return "", errors.New("invalid magic prefix")
	}

	saltItem := bytes.Split(salt, []byte{'$'})
	if len(saltItem) < 3 {
		return "", errors.New("invalid salt format")
	}

	if bytes.HasPrefix(saltItem[2], _rounds) {
		isRoundsDef = true
		pr, err := strconv.ParseInt(string(saltItem[2][7:]), 10, 32)
		if err != nil {
			return "", errors.New("invalid rounds")
		}
		rounds = int(pr)
		if rounds < RoundsMin {
			rounds = RoundsMin
		} else if rounds > RoundsMax {
			rounds = RoundsMax
		}
		salt = saltItem[3]
	} else {
		rounds = RoundsDefault
		salt = saltItem[2]
	}

	if len(salt) > SaltLenMax {
		salt = salt[0:SaltLenMax]
	}

	Alternate := sha512.New()
	Alternate.Write(key)
	Alternate.Write(salt)
	Alternate.Write(key)
	AlternateSum := Alternate.Sum(nil)

	A := sha512.New()
	A.Write(key)
	A.Write(salt)
	i := len(key)
	for ; i > 64; i -= 64 {
		A.Write(AlternateSum)
	}
	A.Write(AlternateSum[0:i])

	for i = len(key); i > 0; i >>= 1 {
		if (i & 1) != 0 {
			A.Write(AlternateSum)
		} else {
			A.Write(key)
		}
	}
	A_sum := A.Sum(nil)

	P := sha512.New()
	for i = 0; i < len(key); i++ {
		P.Write(key)
	}
	P_sum := P.Sum(nil)
	P_seq := make([]byte, 0, len(key))
	for i = len(key); i > 64; i -= 64 {
		P_seq = append(P_seq, P_sum...)
	}
	P_seq = append(P_seq, P_sum[0:i]...)

	S := sha512.New()
	for i = 0; i < (16 + int(A_sum[0])); i++ {
		S.Write(salt)
	}
	S_sum := S.Sum(nil)
	S_seq := make([]byte, 0, len(salt))
	for i = len(salt); i > 64; i -= 64 {
		S_seq = append(S_seq, S_sum...)
	}
	S_seq = append(S_seq, S_sum[0:i]...)

	C_sum := A_sum

	for i = 0; i < rounds; i++ {
		C := sha512.New()
		if (i & 1) != 0 {
			C.Write(P_seq)
		} else {
			C.Write(C_sum)
		}
		if (i % 3) != 0 {
			C.Write(S_seq)
		}
		if (i % 7) != 0 {
			C.Write(P_seq)
		}
		if (i & 1) != 0 {
			C.Write(C_sum)
		} else {
			C.Write(P_seq)
		}

		C_sum = C.Sum(nil)
	}

	out := make([]byte, 0, 123)
	out = append(out, magicPrefix...)
	if isRoundsDef {
		out = append(out, []byte("rounds="+strconv.Itoa(rounds)+"$")...)
	}
	out = append(out, salt...)
	out = append(out, '$')
	out = append(out, base64_24Bit([]byte{
		C_sum[42], C_sum[21], C_sum[0],
		C_sum[1], C_sum[43], C_sum[22],
		C_sum[23], C_sum[2], C_sum[44],
		C_sum[45], C_sum[24], C_sum[3],
		C_sum[4], C_sum[46], C_sum[25],
		C_sum[26], C_sum[5], C_sum[47],
		C_sum[48], C_sum[27], C_sum[6],
		C_sum[7], C_sum[49], C_sum[28],
		C_sum[29], C_sum[8], C_sum[50],
		C_sum[51], C_sum[30], C_sum[9],
		C_sum[10], C_sum[52], C_sum[31],
		C_sum[32], C_sum[11], C_sum[53],
		C_sum[54], C_sum[33], C_sum[12],
		C_sum[13], C_sum[55], C_sum[34],
		C_sum[35], C_sum[14], C_sum[56],
		C_sum[57], C_sum[36], C_sum[15],
		C_sum[16], C_sum[58], C_sum[37],
		C_sum[38], C_sum[17], C_sum[59],
		C_sum[60], C_sum[39], C_sum[18],
		C_sum[19], C_sum[61], C_sum[40],
		C_sum[41], C_sum[20], C_sum[62],
		C_sum[63],
	})...)

	A.Reset()
	Alternate.Reset()
	P.Reset()
	for i = 0; i < len(A_sum); i++ {
		A_sum[i] = 0
	}
	for i = 0; i < len(AlternateSum); i++ {
		AlternateSum[i] = 0
	}
	for i = 0; i < len(P_seq); i++ {
		P_seq[i] = 0
	}

	return string(out), nil
}

func generateWRounds() []byte {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	magicPrefix := "$6$"
	out := make([]byte, len(magicPrefix)+5000)
	copy(out, magicPrefix)
	copy(out[len(magicPrefix):], base64_24Bit(salt))
	return out
}

func base64_24Bit(src []byte) (hash []byte) {
	if len(src) == 0 {
		return []byte{}
	}
	alphabet := "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	hashSize := (len(src) * 8) / 6
	if (len(src) % 6) != 0 {
		hashSize += 1
	}
	hash = make([]byte, hashSize)

	dst := hash
	for len(src) > 0 {
		switch len(src) {
		default:
			dst[0] = alphabet[src[0]&0x3f]
			dst[1] = alphabet[((src[0]>>6)|(src[1]<<2))&0x3f]
			dst[2] = alphabet[((src[1]>>4)|(src[2]<<4))&0x3f]
			dst[3] = alphabet[(src[2]>>2)&0x3f]
			src = src[3:]
			dst = dst[4:]
		case 2:
			dst[0] = alphabet[src[0]&0x3f]
			dst[1] = alphabet[((src[0]>>6)|(src[1]<<2))&0x3f]
			dst[2] = alphabet[(src[1]>>4)&0x3f]
			src = src[2:]
			dst = dst[3:]
		case 1:
			dst[0] = alphabet[src[0]&0x3f]
			dst[1] = alphabet[(src[0]>>6)&0x3f]
			src = src[1:]
			dst = dst[2:]
		}
	}

	return
}
