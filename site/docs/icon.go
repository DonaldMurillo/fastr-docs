package docs

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// DefaultIconPNG is a neutral starter mark for the generated PWA. Replace
// this function with a branded PNG or call uihost.WithAppIcon with your own
// source when the project acquires a visual identity.
func DefaultIconPNG() []byte {
	const size = 256
	canvas := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 248, G: 246, B: 240, A: 255}}, image.Point{}, draw.Src)

	// A compact three-bar mark stays legible when GoFastr derives 32/180/192/
	// 512px variants from this source image.
	draw.Draw(canvas, image.Rect(34, 34, 222, 222), &image.Uniform{C: color.RGBA{R: 27, G: 29, B: 27, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(54, 54, 202, 202), &image.Uniform{C: color.RGBA{R: 248, G: 246, B: 240, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(70, 126, 94, 184), &image.Uniform{C: color.RGBA{R: 241, G: 102, B: 35, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(116, 96, 140, 184), &image.Uniform{C: color.RGBA{R: 241, G: 102, B: 35, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(162, 68, 186, 184), &image.Uniform{C: color.RGBA{R: 241, G: 102, B: 35, A: 255}}, image.Point{}, draw.Src)

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, canvas); err != nil {
		panic(err)
	}
	return encoded.Bytes()
}
