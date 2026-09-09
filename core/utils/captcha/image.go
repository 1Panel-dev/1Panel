package captcha

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand/v2"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var loadFont = sync.OnceValues(func() (*opentype.Font, error) {
	return opentype.Parse(gobold.TTF)
})

func renderQuestion(question string) (string, error) {
	f, err := loadFont()
	if err != nil {
		return "", err
	}
	dst := image.NewRGBA(image.Rect(0, 0, captchaWidth, captchaHeight))
	background := image.NewUniform(color.RGBA{R: 246, G: 248, B: 251, A: 255})
	draw.Draw(dst, dst.Bounds(), background, image.Point{}, draw.Src)
	chars := []rune(question)
	cellWidth := (captchaWidth - 12) / len(chars)
	for i, char := range chars {
		face, err := opentype.NewFace(f, &opentype.FaceOptions{
			Size: float64(24 + rand.IntN(5)), DPI: 72, Hinting: font.HintingFull,
		})
		if err != nil {
			return "", err
		}
		bounds, _ := font.BoundString(face, string(char))
		x := fixed.I(6+i*cellWidth+cellWidth/2) - (bounds.Min.X+bounds.Max.X)/2
		y := fixed.I(captchaHeight/2+rand.IntN(7)-3) - (bounds.Min.Y+bounds.Max.Y)/2
		d := font.Drawer{
			Dst: dst,
			Src: image.NewUniform(color.RGBA{
				R: uint8(20 + rand.IntN(90)), G: uint8(20 + rand.IntN(90)),
				B: uint8(20 + rand.IntN(90)), A: 255,
			}),
			Face: face,
			Dot:  fixed.Point26_6{X: x, Y: y},
		}
		d.DrawString(string(char))
		if err := face.Close(); err != nil {
			return "", err
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
