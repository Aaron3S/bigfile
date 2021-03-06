//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"io"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/gographics/gmagick"
	"gopkg.in/go-playground/validator.v9"
)

// ImageConvert is used to provide convert image service
type ImageConvert struct {
	BaseService

	Token  *models.Token `validate:"required"`
	File   *models.File  `validate:"required"`
	IP     *string       `validate:"omitempty"`
	Type   string        `validate:"required,oneof=thumb zoom crop"`
	Width  float64       `validate:"required"`
	Height float64       `validate:"required"`
	Left   int           `validate:"omitempty"`
	Top    int           `validate:"omitempty"`
}

// GM just wraps the *gmagick.MagickWand
type GM struct {
	MagickWand *gmagick.MagickWand
}

// Validate is used to validate service params
func (ic *ImageConvert) Validate() ValidateErrors {
	var (
		validateErrors ValidateErrors
		errs           error
	)
	if errs = Validate.Struct(ic); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateToken(ic.DB, ic.IP, true, ic.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("ImageConvert.Token", err))
	}

	if err := ValidateFile(ic.DB, ic.File); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("ImageConvert.File", err))
	} else {
		if err := ic.File.CanBeAccessedByToken(ic.Token, ic.DB); err != nil {
			validateErrors = append(validateErrors, generateErrorByField("ImageConvert.Token", err))
		}
	}

	return validateErrors
}

// Execute is used to convert
// Generate thumbnails via the “ImageThumb” type
// Generate crop via the “ImageCrop” method
// Generate Centered zoom cut via the “ImageZoom” method
func (ic *ImageConvert) Execute(ctx context.Context) ([]byte, error) {
	var err error

	if err = ic.Token.UpdateAvailableTimes(-1, ic.DB); err != nil {
		return nil, err
	}

	if ic.File.Hidden == 1 {
		return nil, ErrReadHiddenFile
	}

	fileReader, err := ic.File.Reader(ic.RootPath, ic.DB)

	return ImageConvertRun(fileReader, int64(ic.File.Size), ic.Type, ic.Width, ic.Height, ic.Left, ic.Top)
}

// NewGm is used to init GM
func NewGm() *GM {
	MagickWand := gmagick.NewMagickWand()
	gmagick.Initialize()
	return &GM{
		MagickWand: MagickWand,
	}
}

// ImageThumb is used to thumbnails the image
func (gm *GM) ImageThumb(width, height float64) error {
	w := gm.MagickWand.GetImageWidth()
	h := gm.MagickWand.GetImageHeight()
	x := float64(w) / float64(h)
	var targetW, targetH uint
	if height*x > width {
		targetW = uint(width)
		targetH = uint(width / x)
	} else {
		targetW = uint(height * x)
		targetH = uint(height)
	}
	return gm.MagickWand.ResizeImage(targetW, targetH, gmagick.FILTER_LANCZOS, 1)
}

// ImageCrop is used to crop the image
func (gm *GM) ImageCrop(width, height float64, left, top int) error {
	return gm.MagickWand.CropImage(uint(width), uint(height), left, top)
}

// ImageZoom is used to centered zoom cut the image
func (gm *GM) ImageZoom(width, height float64) error {
	var left, top int
	var x, xW, xH float64

	w := gm.MagickWand.GetImageWidth()
	h := gm.MagickWand.GetImageHeight()
	xW = float64(w) / width
	xH = float64(h) / height

	if xW < xH {
		x = xW
	} else {
		x = xH
	}
	thumbWidth, thumbHeight := uint(float64(w)/x), uint(float64(h)/x)

	if err := gm.MagickWand.ResizeImage(thumbWidth, thumbHeight, gmagick.FILTER_LANCZOS, 1); err != nil {
		return err
	}

	top = int(thumbHeight-uint(height)) / 2
	left = int(thumbWidth-uint(width)) / 2

	return gm.MagickWand.CropImage(uint(width), uint(height), left, top)
}

// ImageConvertRun is used to convert image
func ImageConvertRun(reader io.Reader, size int64, t string, width, height float64, left, top int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	gm := NewGm()
	defer func() {
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	readErr := gm.MagickWand.ReadImageBlob(buf)
	if readErr != nil {
		return nil, readErr
	}

	switch t {
	case "thumb":
		if err := gm.ImageThumb(width, height); err != nil {
			return nil, err
		}
	case "crop":
		if err := gm.ImageCrop(width, height, left, top); err != nil {
			return nil, err
		}
	case "zoom":
		if err := gm.ImageZoom(width, height); err != nil {
			return nil, err
		}
	}

	if err := gm.MagickWand.SetImageFormat("JPEG"); err != nil {
		return nil, err
	}

	return gm.MagickWand.WriteImageBlob(), nil
}
