package qrcode

import (
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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

func TitleGetController(c *gin.Context) {
	qrcode, err := Qrcode.GetTitle(_stringToInt(c.Param("id")))
	if err != nil {
		c.String(http.StatusNotFound, "Error obtaining qrcode data")
		return
	}

	c.JSON(http.StatusOK, TitleResponse{
		Name: qrcode.Name,
	})
}

func ImageGetController(c *gin.Context) {
	qrcode, err := Qrcode.GetImage(_stringToInt(c.Param("id")))
	if err != nil {
		c.String(http.StatusNotFound, "Error obtaining qrcode data")
		return
	}

	// because for some reason garmin sdk can't forward header on image request
	reqApiKey := c.Query("apiKey")
	if reqApiKey != apiPngGetKey {
		c.String(http.StatusUnauthorized, "Nope")
		return
	}

	c.Data(http.StatusOK, "image/png", qrcode.Image)
}

func AddPostController(c *gin.Context) {
	// parse request
	p := new(QrcodeRequestItem)
	if err := c.ShouldBindJSON(p); err != nil {
		log.Error().Err(err).Msg("Error parsing request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot parse JSON request body",
		})
		return
	}

	// insert
	imageBytes, err := base64.StdEncoding.DecodeString(p.Image)
	if err != nil {
		log.Error().Err(err).Msg("Error decoding base64 image")
	}

	//// image processing
	imageGrayScaleBytes, _ := pngToGrayScale(imageBytes)
	imageCropBorderBytes, _ := pngCropBorder(imageGrayScaleBytes)
	//// -- resize to 90x90 so garmin doesn't choke
	imageResizedBytes, _ := pngResize(imageCropBorderBytes)

	//// insert to db
	err = Qrcode.Add(QrcodeItem{
		ID:    p.ID,
		Name:  p.Name,
		Image: imageResizedBytes,
	})
	if err != nil {
		log.Printf("Error adding image: %v", err)
	}

	c.String(http.StatusOK, "Success")
}

func _stringToInt(s string) int {
	id, err := strconv.Atoi(s)
	if err != nil {
		log.Error().Err(err).Msgf("Error converting to int: %s", s)
	}

	return id
}
