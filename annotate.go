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
	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
	"image"
	//"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
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
	X      int32
	Y      int32
	Width  int32
	Height int32
}

// Color holds RGBA data. Implements image/color's color.Color interface
type Color struct {
	R, G, B, A uint32
}

func (c Color) RGBA() (uint32, uint32, uint32, uint32) {
	return c.R, c.G, c.B, c.A
}

type Line struct {
	Words     []string
	XPos      int32
	YPos      int32
	currWidth int32
	//distance from the base of this line to the base of the line above it.
	// height of line + line-spacing.
	//for the first line it is just the line height
	BaseToBaseHeight int32
}

type Context struct {
	src            image.Image
	fontBackground draw.Image // font background
	font           *truetype.Font
	maxFontSize    int32
	dpi            float64
	lineHeight     float64
	fontColor      image.Image
	fontSize       int32 //Font size calculated to fit inside of hte bounding box
}

// NewContext returns a pointer to a new instance of Context
func NewContext() *Context {
	return &Context{}
}

// SetSrc sets the srouce image. This is the image on which the annotation is to be drawn
// As of right now, Annotate only supports JPEG and PNG formats
func (c *Context) SetSrc(src image.Image) {
	c.src = src
	c.fontBackground = image.NewRGBA(src.Bounds())
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
	case ".png":
		imageDecoded, err = png.Decode(imageRaw)
		break
	case ".jpg", ".jpeg":
		imageDecoded, err = jpeg.Decode(imageRaw)
		break
	default:
		return UnsupportedError(extension)
	}

	if err != nil {
		return err
	}

	c.src = imageDecoded
	c.fontBackground = image.NewRGBA(imageDecoded.Bounds())

	return nil
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
	return err
}

// SetFont Directly sets the truetype.Font data
// See SetFontPath for more information
func (c *Context) SetFont(font *truetype.Font) {
	c.font = font
}

// setMaxFontSize sets the maximum size the text can be in points.
//
// Note: No guerentee is made that the text will be at this size point.
// Annotate will choose the largest font size UP TO this size that will fit
// within the text box; which is specified in writeText
func (c *Context) SetMaxFontSize(size int) {
	c.maxFontSize = int32(size)
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
	c.lineHeight = height
}

// SetFontColor sets a solid color for fonts
func (c *Context) SetFontColor(rgbaColor Color) {
	//draw.Draw(c.fontBackground, c.src.Bounds(),
	//image.NewUniform(rgbaColor), image.ZP, 0)

	//c.fontBackground = *image.NewUniform(rgbaColor)
	c.fontColor = image.NewUniform(rgbaColor)
}

// SetFontImage sets an image as the color for text
func (c *Context) SetFontImage(backgroundImage draw.Image) {
	draw.Draw(c.fontBackground, c.fontBackground.Bounds(), backgroundImage, image.ZP, 0)
}

func (c *Context) wordWidth(word string, fontSize int32) int32 {
	//fUnitsPerEm := c.font.FUnitsPerEm()
	width := int32(0)
	wordLen := len(word)
	for i := 0; i < wordLen; i++ {
		index := c.font.Index(rune(word[i]))
		//64 = units per pixel
		hMetric := c.font.HMetric(fontSize, index)
		width += hMetric.AdvanceWidth
		if i != wordLen-1 {
			index2 := c.font.Index(rune(word[i+1]))
			width += c.font.Kerning(fontSize, index, index2)
		}
	}
	return width
}

func (c *Context) WriteText(text string, boundingBox Rectangle) (error, image.Image) {
	_, lines, fontSize := c.calculateSize(text, boundingBox, c.maxFontSize, -1, 0)
	return c.drawLines(lines, boundingBox, fontSize)
}

func (c *Context) calculateSize(text string, boundingBox Rectangle, fontSize int32, attempt int32, lastFit int32) (bool, []*Line, int32) {
	hardLines := strings.Split(text, "\n")
	lardLinesLen := len(hardLines)
	spaceWidth := c.wordWidth(" ", fontSize)
	lines := []*Line{}
	for i := 0; i < lardLinesLen; i++ {
		hardLine := hardLines[i]
		words := strings.Split(hardLine, " ")
		wordsLen := len(words)

		lines = append(lines, &Line{})
		currLine := lines[len(lines)-1]
		for j := 0; j < wordsLen; j++ {
			word := words[j]
			wordWidth := c.wordWidth(word, fontSize)
			if (wordWidth + currLine.currWidth + spaceWidth) > (int32(boundingBox.Width)) {
				lines = append(lines, &Line{})
				currLine = lines[len(lines)-1]
			}
			currLine.Words = append(currLine.Words, word)
			currLine.currWidth += spaceWidth + wordWidth
		}
	}

	totalHeight := int32(0)
	linesLen := len(lines)
	for i := 0; i < linesLen; i++ {
		line := lines[i]

		if i == 0 {
			line.YPos += boundingBox.Y
			line.BaseToBaseHeight = fontSize
		} else {
			overHeadSpace := int32(c.lineHeight*float64(fontSize)) + fontSize
			line.YPos += lines[i-1].YPos + overHeadSpace
			line.BaseToBaseHeight = overHeadSpace
		}
		line.XPos += boundingBox.X
		totalHeight += line.BaseToBaseHeight
	}
	//we add a little buffer zone to the bottom of the text to make sure some runes like "g" don't get cut off
	totalHeight += int32(c.lineHeight * float64(fontSize))

	attempt++
	// if we are trying with the user specified max font size and it fits, return
	if attempt == 0 && totalHeight <= boundingBox.Height {
		return true, lines, fontSize
	}
	if totalHeight <= boundingBox.Height {
		if fontSize == lastFit {
			return true, lines, fontSize
		} else if math.Abs(float64(fontSize-lastFit)) == 1 {
			return true, lines, larger(lastFit, fontSize)
		} else {
			newFontSize := int32(math.Floor(float64(fontSize+c.maxFontSize)/2 + 0.5))
			return c.calculateSize(text, boundingBox, newFontSize, attempt, fontSize)
		}
	} else {
		if math.Abs(float64(fontSize-lastFit)) == 1 {
			return true, lines, larger(lastFit, fontSize)
		} else {
			newFontSize := int32(math.Floor(float64(fontSize+lastFit)/2 + 0.5))
			return c.calculateSize(text, boundingBox, newFontSize, attempt, lastFit)
		}

	}
}

func larger(a, b int32) int32 {
	if a >= b {
		return a
	} else {
		return b
	}
}

func (c *Context) drawLines(lines []*Line, boundingBox Rectangle, fontSize int32) (error, image.Image) {
	imageContext := freetype.NewContext()
	imageContext.SetFont(c.font)
	imageContext.SetDPI(c.dpi)

	rbga := image.NewRGBA(c.src.Bounds())
	draw.Draw(rbga, rbga.Bounds(), c.src, image.ZP, 0)

	imageContext.SetSrc(c.fontColor)
	imageContext.SetDst(rbga)
	imageContext.SetHinting(freetype.FullHinting)
	imageContext.SetFontSize(float64(fontSize))
	imageContext.SetClip(c.src.Bounds())
	for _, line := range lines {
		words := strings.Join(line.Words, " ")
		_, err := imageContext.DrawString(words, raster.Point{X: raster.Fix32(line.XPos << 8), Y: raster.Fix32((line.YPos) << 8)})
		if err != nil {
			return err, rbga
		}
	}
	return nil, rbga
}
