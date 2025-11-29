package processor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// WatermarkProcessor handles the logic of adding a text watermark to an image.

type WatermarkProcessor struct {
	font         *truetype.Font
	fontSize     float64
	fontColor    color.Color
	imageQuality int
}

// NewWatermarkProcessor initializes a processor with font and style settings.
func NewWatermarkProcessor(fontBytes []byte, fontSize float64, fontColor color.Color, imageQuality int) (*WatermarkProcessor, error) {
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	return &WatermarkProcessor{
		font:         font,
		fontSize:     fontSize,
		fontColor:    fontColor,
		imageQuality: imageQuality,
	}, nil
}

// AddWatermark takes an image byte slice and adds a text overlay.
func (p *WatermarkProcessor) AddWatermark(imageBytes []byte, text string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	bounds := rgba.Bounds()
	point := p.calculateTextPosition(bounds, text)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(p.font)
	c.SetFontSize(p.fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.NewUniform(p.fontColor))
	c.SetHinting(font.HintingNone)

	_, err = c.DrawString(text, point)
	if err != nil {
		return nil, fmt.Errorf("failed to draw string: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, rgba, &jpeg.Options{Quality: p.imageQuality}); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// calculateTextPosition determines where to place the watermark text.
// Currently, it centers the text at the bottom of the image.
func (p *WatermarkProcessor) calculateTextPosition(bounds image.Rectangle, text string) fixed.Point26_6 {
	face := truetype.NewFace(p.font, &truetype.Options{Size: p.fontSize})
	textWidth := font.MeasureString(face, text)

	x := (bounds.Max.X - int(textWidth)>>6) / 2
	y := bounds.Max.Y - int(p.fontSize*1.5) // Positioned slightly above the bottom edge

	return fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}
}
