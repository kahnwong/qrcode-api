package qrcode

import (
	"encoding/base64"
	"log/slog"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

var (
	apiPngGetKey = os.Getenv("QRCODE_IMAGE_GET_API_KEY")
)

type TitleResponse struct {
	Name string `json:"name"`
}

type QrcodeRequestItem struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"` // base64
}

func TitleGetController(c fiber.Ctx) error {
	qrcode, err := Qrcode.GetTitle(c.Context(), _stringToInt(c.Params("id")))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Error obtaining qrcode data")
	}

	return c.Status(fiber.StatusOK).JSON(TitleResponse{
		Name: qrcode.Name,
	})
}

func ImageGetController(c fiber.Ctx) error {
	qrcode, err := Qrcode.GetImage(c.Context(), _stringToInt(c.Params("id")))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Error obtaining qrcode data")
	}

	// because for some reason garmin sdk can't forward header on image request
	reqApiKey := c.Query("apiKey")
	if reqApiKey != apiPngGetKey {
		return c.Status(fiber.StatusUnauthorized).SendString("Nope")
	}

	c.Set(fiber.HeaderContentType, "image/png")
	return c.Status(fiber.StatusOK).Send(qrcode.Image)
}

func AddPostController(c fiber.Ctx) error {
	// parse request
	p := new(QrcodeRequestItem)
	if err := c.Bind().Body(p); err != nil {
		slog.ErrorContext(c.Context(), "error parsing request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON request body",
		})
	}

	// insert
	imageBytes, err := base64.StdEncoding.DecodeString(p.Image)
	if err != nil {
		slog.ErrorContext(c.Context(), "error decoding base64 image", "error", err)
	}

	//// image processing
	imageGrayScaleBytes, _ := pngToGrayScale(imageBytes)
	imageCropBorderBytes, _ := pngCropBorder(imageGrayScaleBytes)
	//// -- resize to 90x90 so garmin doesn't choke
	imageResizedBytes, _ := pngResize(imageCropBorderBytes)

	//// insert to db
	err = Qrcode.Add(c.Context(), QrcodeItem{
		ID:    p.ID,
		Name:  p.Name,
		Image: imageResizedBytes,
	})
	if err != nil {
		slog.ErrorContext(c.Context(), "error adding image", "error", err)
	}

	return c.Status(fiber.StatusOK).SendString("Success")
}

func _stringToInt(s string) int {
	id, err := strconv.Atoi(s)
	if err != nil {
		slog.Error("error converting to int", "value", s, "error", err)
	}

	return id
}
