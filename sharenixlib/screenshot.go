/*
   Copyright 2014 Franc[e]sco (lolisamurai@tfwno.gf)
   This file is part of sharenix.
   sharenix is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   sharenix is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with sharenix. If not, see <http://www.gnu.org/licenses/>.
*/

package sharenixlib

import (
	"errors"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/BurntSushi/xgb/xproto"
	"image"
	"image/draw"
	"sort"
)

// This file is heavily inspired by https://github.com/vova616/screenshot
// Full credits to vova616 for the method.

// A ScreenRect holds information about the bounds and screen index of a monitor
type ScreenRect struct {
	// Bounds of the screen (position & size)
	Rect image.Rectangle
	// Index of the screen in SetupInfo.Roots for xproto.
	// -1 means the default screen.
	ScreenIndex int
}

// ByX is a sorter for a slice of ScreenRect pointers.
type ByX []*ScreenRect

func (a ByX) Len() int      { return len(a) }
func (a ByX) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByX) Less(i, j int) bool {
	return a[i].Rect.Min.X < a[j].Rect.Min.X
}

// ScreenRects returns a slice of the bounds and scteen id of each monitor.
func ScreenRects(X *xgb.Conn) (rects []*ScreenRect, err error) {
	// xinerama
	for {
		xinerr := xinerama.Init(X)
		if xinerr != nil {
			DebugPrintln(xinerr)
			break
		}

		var reply *xinerama.QueryScreensReply
		reply, err = xinerama.QueryScreens(X).Reply()
		if err != nil {
			DebugPrintln(err)
			break
		}

		if reply.Number >= 2 {
			// multiple xinerama heads
			DebugPrintln("Using", reply.Number, "xinerama heads")
			for i, screen := range reply.ScreenInfo {
				DebugPrintf("%d\tX: %d\tY: %d\tWidth: %d\tHeight: %d\n",
					i, screen.XOrg, screen.YOrg, screen.Width, screen.Height)

				rects = append(rects, &ScreenRect{
					image.Rect(
						int(screen.XOrg), int(screen.YOrg),
						int(screen.XOrg)+int(screen.Width),
						int(screen.YOrg)+int(screen.Height),
					),
					-1,
				})
			}
			return
		}

		break
	}

	// xproto
	setupInfo := xproto.Setup(X)
	if setupInfo == nil {
		err = errors.New("Failed to retrieve X setup info.")
		return
	}

	// no multiple xinerama heads,
	DebugPrintln("Using", len(setupInfo.Roots), "screens")
	x := 0
	for i, s := range setupInfo.Roots {
		DebugPrintf("%d\tX: %d\tY: 0\tWidth: %d\tHeight: %d\n",
			i, x, s.WidthInPixels, s.HeightInPixels)
		rects = append(rects, &ScreenRect{
			image.Rect(0, 0, int(s.WidthInPixels), int(s.HeightInPixels)), i,
		})
		x += x + int(s.WidthInPixels)
	}

	return
}

// CaptureScreen captures all screens and returns an uncompressed image
func CaptureScreen(X *xgb.Conn) (pic *image.RGBA, err error) {
	rects, err := ScreenRects(X)
	if err != nil {
		return
	}

	sort.Sort(ByX(rects))

	// iterate all screens and screenshot them individually. this is necessary
	// even on xinerama setups because otherwise different height monitors would
	// cause garbage image data in the empty areas

	totalWidth := 0
	totalHeight := 0
	for _, r := range rects {
		totalWidth += r.Rect.Dx()
		if r.Rect.Dy() > totalHeight {
			totalHeight = r.Rect.Dy()
		}
	}

	DebugPrintln("Building", totalWidth, "x", totalHeight, "image")
	pic = image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))
	draw.Draw(pic, pic.Bounds(), image.Transparent, image.ZP, draw.Src)

	x := 0
	for _, r := range rects {
		var screen *image.RGBA
		screen, err = CaptureRect(X, r.ScreenIndex, r.Rect)
		if err != nil {
			return
		}
		finalrect := screen.Bounds().Add(image.Pt(x, 0))
		draw.Draw(pic, finalrect, screen, image.ZP, draw.Src)
		DebugPrintln("Drawing", finalrect)
		x += r.Rect.Bounds().Dx()
	}

	return
}

// CaptureRect captures a section of the desired screen.
// Returns an uncompressed image.
// screenIndex = -1 gets the default screen.
func CaptureRect(X *xgb.Conn, screenIndex int, rect image.Rectangle) (
	pic *image.RGBA, err error) {

	// get setup info
	setupInfo := xproto.Setup(X)
	if setupInfo == nil {
		err = errors.New("Failed to retrieve X setup info!")
		return
	}

	var screen *xproto.ScreenInfo
	if screenIndex == -1 {
		screen = setupInfo.DefaultScreen(X)
		if screen == nil {
			err = errors.New("No default screen found")
			return
		}
	} else {
		screen = &setupInfo.Roots[screenIndex]
	}

	// capture screen
	xImg, err := xproto.GetImage(X, xproto.ImageFormatZPixmap,
		xproto.Drawable(screen.Root), int16(rect.Min.X), int16(rect.Min.Y),
		uint16(rect.Dx()), uint16(rect.Dy()), 0xFFFFFFFF).Reply()
	if err != nil {
		return
	}

	// convert to rgba
	data := xImg.Data
	for i := 0; i < len(data); i += 4 {
		data[i], data[i+2], data[i+3] = data[i+2], data[i], 255
	}

	pic = &image.RGBA{data, 4 * rect.Dx(),
		image.Rect(0, 0, rect.Dx(), rect.Dy())}
	return
}
