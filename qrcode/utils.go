package qrcode

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/nfnt/resize"
	"github.com/rs/zerolog/log"
)

func pngToGrayScale(imageBytes []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		log.Error().Err(err).Msg("failed to decode PNG image")
	}

	// 4. Create a new grayscale image with the same bounds as the original
	bounds := img.Bounds()
	grayImage := image.NewGray(bounds)

	// 5. Draw the original image onto the grayscale image.
	// The draw.Draw function automatically handles the conversion to grayscale.
	draw.Draw(grayImage, bounds, img, bounds.Min, draw.Src)

	//// encode back to bytes
	var buf bytes.Buffer
	err = png.Encode(&buf, grayImage)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode resized PNG image")
	}
	return buf.Bytes(), err
}

func pngResize(imageBytes []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		log.Error().Err(err).Msg("failed to decode PNG image")
	}
	resizedImg := resize.Resize(90, 90, img, resize.Lanczos3)

	//// encode back to bytes
	var buf bytes.Buffer
	err = png.Encode(&buf, resizedImg)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode resized PNG image")
	}
	return buf.Bytes(), err
}

func pngCropBorder(imageBytes []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		log.Error().Err(err).Msg("failed to decode PNG image")
		return nil, err
	}

	bounds := img.Bounds()
	cropBounds, ok := nonWhiteBounds(img, bounds)
	if !ok {
		// The image is completely white/transparent. Return it unchanged.
		return imageBytes, nil
	}

	croppedImage := image.NewRGBA(image.Rect(0, 0, cropBounds.Dx(), cropBounds.Dy()))
	draw.Draw(croppedImage, croppedImage.Bounds(), img, cropBounds.Min, draw.Src)

	var buf bytes.Buffer
	err = png.Encode(&buf, croppedImage)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode cropped PNG image")
	}
	return buf.Bytes(), err
}

func nonWhiteBounds(img image.Image, bounds image.Rectangle) (image.Rectangle, bool) {
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isWhiteLike(img.At(x, y)) {
				continue
			}

			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+1 > maxX {
				maxX = x + 1
			}
			if y+1 > maxY {
				maxY = y + 1
			}
		}
	}

	if minX >= maxX || minY >= maxY {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX, maxY), true
}

func isWhiteLike(c color.Color) bool {
	const whiteThreshold = 245 * 257

	r, g, b, a := c.RGBA()
	// RGBA returns alpha-premultiplied values. Composite against white so that
	// transparent or semi-transparent white border pixels are treated as border.
	r += 0xffff - a
	g += 0xffff - a
	b += 0xffff - a

	return r >= whiteThreshold && g >= whiteThreshold && b >= whiteThreshold
}
