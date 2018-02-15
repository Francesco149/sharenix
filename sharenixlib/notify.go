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
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/pango"
	"os"
	"path"
	"time"
	"unsafe"
)

func lockName(index int) string {
	return fmt.Sprintf(".notifyf%d", index)
}

func lockFile(index int) (file string, err error) {
	storage, err := GetStorageDir()
	if err != nil {
		return
	}
	file = path.Join(storage, lockName(index))
	return
}

// to prevent notifications overlapping, Notifyf uses lock files for each
// vertical position index.
func lockFree(index int) (available bool, err error) {
	file, err := lockFile(index)
	if err != nil {
		return
	}
	_, staterr := os.Stat(file)
	available = os.IsNotExist(staterr)
	return
}

func lockBegin(index int) (err error) {
	DebugPrintln("Locking position", index)
	file, err := lockFile(index)
	if err != nil {
		return
	}
	f, err := os.Create(file)
	defer f.Close()
	return
}

func lockEnd(index int) (err error) {
	DebugPrintln("Unlocking position", index)
	file, err := lockFile(index)
	if err != nil {
		return
	}
	return os.Remove(file)
}

// Notifyf formats and shows a notification as a bordeless GTK window in the
// bottom right corner of the screen.
// Right-clicking the notification dismisses it and terminates the process.
// expire is the time after which the notification will expire automatically.
// onInit is a goroutine that will be started before the gtk main loop
// blocks the main thread. It takes the notification window as a parameter.
func Notifyf(head uint32, expire time.Duration, onInit func(*gtk.Window),
	format string, a ...interface{}) (err error) {

	win := gtk.NewWindow(gtk.WINDOW_POPUP)
	win.SetTitle(ShareNixVersion)
	win.SetDecorated(false)
	win.SetKeepAbove(true)
	win.Connect("destroy", gtk.MainQuit)

	lockIndex := 0

	// Handle left/right click
	win.Connect("button-press-event", func(ctx *glib.CallbackContext) {
		DebugPrintln("button-press-event")

		arg := ctx.Args(0)
		e := *(**gdk.EventButton)(unsafe.Pointer(&arg))

		DebugPrintln("button =", e.Button)
		switch e.Button {
		case 3: // right click (single)
			lockEnd(lockIndex)
			os.Exit(0)
		}
	})

	l := gtk.NewLabel("")

	// PANGO_ELLIPSIZE_END automatically limits the text length when the
	// window is resized and appends ... at the end
	notiftext := fmt.Sprintf("ShareNix: "+format, a...)
	l.SetSingleLineMode(true)
	l.SetMaxWidthChars(60) // workaround for a positioning bug, see below
	l.SetEllipsize(pango.ELLIPSIZE_END)
	l.Misc.SetAlignment(0.5, 0.5)
	l.SetPadding(10, 10)
	l.SetMarkup(notiftext)
	win.Add(l)
	win.ShowAll() // parent window will adjust to the size of the label here

	X, err := xgb.NewConn()
	if err != nil {
		return
	}

	var rects []*ScreenRect
	rects, err = ScreenRects(X)
	if err != nil {
		return
	}

	X.Close()

	// our window is now the right width for the notification text
	// (clamped to a max of 60 chars).

	// calculate notification position so that it doens't overlap other notifs
	width, height := win.GetSize()
	y := rects[head].Rect.Max.Y - height - 10

	for ; ; lockIndex++ {
		var free bool
		free, err = lockFree(lockIndex)
		if err != nil {
			return
		}
		if free {
			err = lockBegin(lockIndex)
			if err != nil {
				return
			}
			break
		}

		y -= height
		y -= 10

		if y < 0 {
			y = rects[head].Rect.Max.Y - height - 10
		}
	}

	// Position window in the bottom right corner of the screen
	win.Move(rects[head].Rect.Max.X-width-10, y)

	// ghetto way to fix the positioning bug when resizing a pango.ELLIPSIZE_END
	// widget by limiting text width first and then restoring full text after
	// the window has been moved to the proper position.
	// not doing this prevents the window from being moved to the corner
	// corrently when the full text exceeds the width limit (it's like the
	// window manager thinks the window is as wide as the full text even after
	// a resize).
	l.SetMaxWidthChars(-1)
	l.SetMarkup(notiftext)

	if onInit != nil {
		go onInit(win)
	}

	DebugPrintln("starting gtk.Main()")
	dismissed := make(chan bool)
	go func() { // timer that kills the notification if the user doesn't
		select {
		case <-dismissed:
			DebugPrintln("Notification dismissed")
		case <-time.After(expire):
			DebugPrintln("Notification expired after", expire)
			glib.IdleAdd(win.Destroy)
			<-dismissed
		}
		close(dismissed)
	}()
	gtk.Main()
	DebugPrintln("exited gtk.Main()")
	dismissed <- true
	return lockEnd(lockIndex)
}
