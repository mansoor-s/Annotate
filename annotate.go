/*
   Copyright (c) 2014 Triple Crown Sports Inc

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in
   all copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
   THE SOFTWARE.
*/

// Package Annotate provides an intuitive and high-level API for adding annotations to images
// It builds on top of the code.google.com/p/freetype-go package
package Annotate

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// UnsupportedError type returned for unsupported image types
type UnsupportedError string

func (f UnsupportedError) Error() string {
	return "Image format not supported: " + string(f)
}

// Rectangle is used for specifying text bounding box
type Rectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Color holds RGBA data. Implements image/color's color.Color interface
type Color struct {
	R, G, B, A uint32
}

func (c *Color) RGBA() (uint32, uint32, uint32, uint32) {
	return c.R, c.G, c.B, c.A
}

type Context struct {
	src                 image.Image
	dst                 draw.Image
	font                truetype.Font
	maxFontSize         int
	dpi                 float64
	writingArea         Rectangle
	fontBackgroundImage image.Image
}

// NewContext returns a pointer to a new instance of Context
func NewContext() *Context {
	return &Context{}
}

// SetSrc sets the srouce image. This is the image on which the annotation is to be drawn
// As of right now, Annotate only supports JPEG and PNG formats
func (c *Context) SetSrc(src image.Image) {
	c.src = src
	c.dst = image.NewRGBA(src.Bounds())
}

// SetSrcPath does same thing as SetSrc, except it will fetch the image, given the path to it.
// As of right now, Annotate only supports JPEG and PNG formats
func (c *Context) SetSrcPath(path string) error {
	imageRaw, err := os.Open(path)
	if err != nil {
		return err
	}

	var imageDecoded image.Image

	extension := strings.ToLower(filepath.Ext(path))

	switch extension {
	case "png":
		imageDecoded, err = png.Decode(imageRaw)
		break
	case "jpg", "jpeg":
		imageDecoded, err = jpeg.Decode(imageRaw)
		break
	default:
		return UnsupportedError{extension}
	}

	if err != nil {
		return err
	}

	c.src = imageDecoded
	c.dst = image.NewRGBA(imageDecoded.Bounds())
}

// SetFontPath loads and parses the font for which the path is specified.
//
// Only TTF or TTC formats are supported.
// From freetype-go docs: "For TrueType Collections, the first font in the collection is parsed"
func (c *Context) SetFontPath(path string) error {
	fontRaw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	c.font, err = freetype.ParseFont(fontRaw)
	if err != nil {
		return err
	}
}

// SetFont Directly sets the truetype.Font data
// See SetFontPath for more information
func (c *Context) SetFont(font truetype.Font) {
	c.font = font
}

// setMaxFontSize sets the maximum size the text can be in points.
//
// Note: No guerentee is made that the text will be at this size point.
// Annotate will choose the largest font size UP TO this size that will fit
// within the text box; which is specified in writeText
func (c *Context) SetMaxFontSize(size float64) {
	c.maxFontSize = size
}

// SetDPI sets the screen resolution in pixels per inch. Defaults to 81.58
func (c *Context) SetDPI(dpi float64) {
	c.dpi = dpi
}

// SetLineHeight sets the line height in em units to be used when annotating the image. It scales to the
// font size
//
// Note: 1 em = 100% font size.
//
// Example: If max font size is 80 and lineHeight is set to .5, but it turns out that the
// Largest font size that can be fitted inside ofthe box is 60, then the line height is actully
// 30 points, not 40.
func (c *Context) SetLineHeight(height float64) {

}

// SetFontColor sets a solid color for fonts
func (c *Context) SetFontColor(rgbaColor Color) {
	c.fontBackgroundImage = image.NewUniform(rgbaColor)
}

// SetFontImage sets an image as the color for text
func (c *Context) SetFontImage(backgroundImage image.Image) {
	c.fontBackgroundImage = backgroundImages
}

func (c *Context) wordWidth(word string, fontSize int32) {
	width := 0
	wordLen := len(word)
	for i := 0; i < wordLen; i++ {
		index := c.font.Index(word[i])
		//64 = units per pixel
		hMetric := c.font.HMetric(fontSize*64, index)
		width += hMetric.AdvanceWidth
		if i != wordLen-1 {
			index2 := c.font.Index(word[i+1])
			width += c.font.Kerning(fontSize*64, index, index2)
		}
	}
}

func (c *Context) WriteText(text string, boundingBox Rectangle) {
	fUnitsPerEm := c.font.FUnitsPerEm()
	hardLines := strings.Split(text, "\n")
	lardLinesLen := len(hardLines)

	spaceWidth := wordWidth(" ", c.maxFontSize)

	for i := 0; i < lardLinesLen; i++ {
		hardLine := hardLines[i]
		words := strings.Split(hardLine, " ")
	}
}
