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
	"github.com/BurntSushi/xgb/xproto"
	"image"
)

// This file is heavily inspired by https://github.com/vova616/screenshot
// Full credits to vova616 for the method.

// defaultSceen returns the default screen info
func defaultScreen(con *xgb.Conn) (screen *xproto.ScreenInfo, err error) {
	// get setup info
	setupInfo := xproto.Setup(con)
	if setupInfo == nil {
		err = errors.New("Failed to retrieve X setup info!")
		return
	}

	// get default screen TODO: iterate all screens for multi monitor setups
	screen = setupInfo.DefaultScreen(con)
	if screen == nil {
		err = errors.New("No screens detected!")
	}

	return
}

// FullScreenRect returns a rectangle of the entire screen
func FullScreenRect() (rect image.Rectangle, err error) {
	// connect to the X server
	con, err := xgb.NewConn()
	if err != nil {
		return
	}
	defer con.Close()

	screen, err := defaultScreen(con)
	if err != nil {
		return
	}

	rect = image.Rect(0, 0, int(screen.WidthInPixels),
		int(screen.HeightInPixels))
	return
}

// CaptureScreen captures the default screen and returns an uncompressed image
func CaptureScreen() (pic *image.RGBA, err error) {
	rect, err := FullScreenRect()
	if err != nil {
		return
	}
	return CaptureRect(rect)
}

// CaptureRect captures the given section of
// the screen and returns an uncompressed image
func CaptureRect(rect image.Rectangle) (pic *image.RGBA, err error) {
	con, err := xgb.NewConn()
	if err != nil {
		return
	}
	defer con.Close()

	screen, err := defaultScreen(con)
	if err != nil {
		return
	}

	// capture screen
	xImg, err := xproto.GetImage(con, xproto.ImageFormatZPixmap,
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
