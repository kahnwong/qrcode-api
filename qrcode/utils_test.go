package qrcode

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestPngCropBorderRemovesWhiteAndNearWhiteMargins(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{R: 248, G: 248, B: 248, A: 255})
		}
	}
	for y := 5; y < 15; y++ {
		for x := 4; x < 16; x++ {
			img.Set(x, y, color.Black)
		}
	}

	var input bytes.Buffer
	if err := png.Encode(&input, img); err != nil {
		t.Fatalf("encode input: %v", err)
	}

	croppedBytes, err := pngCropBorder(input.Bytes())
	if err != nil {
		t.Fatalf("pngCropBorder returned error: %v", err)
	}

	cropped, err := png.Decode(bytes.NewReader(croppedBytes))
	if err != nil {
		t.Fatalf("decode cropped image: %v", err)
	}

	got := cropped.Bounds()
	want := image.Rect(0, 0, 12, 10)
	if got != want {
		t.Fatalf("cropped bounds = %v, want %v", got, want)
	}
}

func TestPngCropBorderLeavesAllWhiteImageUnchanged(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 7, 9))
	for y := 0; y < 9; y++ {
		for x := 0; x < 7; x++ {
			img.Set(x, y, color.White)
		}
	}

	var input bytes.Buffer
	if err := png.Encode(&input, img); err != nil {
		t.Fatalf("encode input: %v", err)
	}

	croppedBytes, err := pngCropBorder(input.Bytes())
	if err != nil {
		t.Fatalf("pngCropBorder returned error: %v", err)
	}

	cropped, err := png.Decode(bytes.NewReader(croppedBytes))
	if err != nil {
		t.Fatalf("decode cropped image: %v", err)
	}

	got := cropped.Bounds()
	want := image.Rect(0, 0, 7, 9)
	if got != want {
		t.Fatalf("cropped bounds = %v, want %v", got, want)
	}
}
