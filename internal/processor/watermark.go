package processor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"sync"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type WatermarkProcessor struct {
	font       *truetype.Font
	fontSize   float64
	fontColor  color.Color
	bufferPool *sync.Pool
}

type WatermarkOptions struct {
	Weight     float64
	Dimensions string
	Quality    int
}

func NewWatermarkProcessor(fontPath string) (*WatermarkProcessor, error) {
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	return &WatermarkProcessor{
		font:      font,
		fontSize:  24.0,
		fontColor: color.White,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}, nil
}

func (p *WatermarkProcessor) AddWatermark(imageData []byte, opts WatermarkOptions) ([]byte, error) {
	// Detect image format by checking file header
	format := detectImageFormat(imageData)

	var img image.Image
	var err error

	// Decode based on detected format
	switch format {
	case "jpeg", "jpg":
		img, err = jpeg.Decode(bytes.NewReader(imageData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode jpeg: %w", err)
		}
	case "png":
		img, err = png.Decode(bytes.NewReader(imageData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode png: %w", err)
		}
	default:
		// Try generic decode as fallback
		img, _, err = image.Decode(bytes.NewReader(imageData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %w", err)
		}
		format = "jpeg" // Default to jpeg for encoding
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	if err := p.drawText(rgba, opts); err != nil {
		return nil, fmt.Errorf("failed to draw text: %w", err)
	}

	buf := p.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufferPool.Put(buf)
	}()

	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(buf, rgba, &jpeg.Options{Quality: opts.Quality}); err != nil {
			return nil, fmt.Errorf("failed to encode jpeg: %w", err)
		}
	case "png":
		if err := png.Encode(buf, rgba); err != nil {
			return nil, fmt.Errorf("failed to encode png: %w", err)
		}
	default:
		if err := jpeg.Encode(buf, rgba, &jpeg.Options{Quality: opts.Quality}); err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// detectImageFormat detects image format by checking file header
func detectImageFormat(data []byte) string {
	if len(data) < 4 {
		return "jpeg" // Default
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg"
	}

	// PNG: 89 50 4E 47
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}

	// Default to jpeg
	return "jpeg"
}

func (p *WatermarkProcessor) drawText(img *image.RGBA, opts WatermarkOptions) error {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(p.font)
	c.SetFontSize(p.fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(p.fontColor))
	c.SetHinting(font.HintingFull)

	pt := freetype.Pt(20, 50)
	if _, err := c.DrawString(fmt.Sprintf("Weight: %.2f kg", opts.Weight), pt); err != nil {
		return err
	}

	pt = freetype.Pt(20, 90)
	if _, err := c.DrawString(fmt.Sprintf("Dimensions: %s", opts.Dimensions), pt); err != nil {
		return err
	}

	return nil
}

func (p *WatermarkProcessor) ProcessBatch(images [][]byte, opts []WatermarkOptions) ([][]byte, error) {
	if len(images) != len(opts) {
		return nil, fmt.Errorf("images and options length mismatch")
	}

	results := make([][]byte, len(images))
	errors := make([]error, len(images))

	var wg sync.WaitGroup
	for i := range images {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			result, err := p.AddWatermark(images[idx], opts[idx])
			results[idx] = result
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
