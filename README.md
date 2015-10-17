<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Feature progress](#feature-progress)
- [Getting started - Prebuilt x64 binaries](#getting-started---prebuilt-x64-binaries)
- [Screenshotting areas or windows](#screenshotting-areas-or-windows)
- [Getting started - Building from the source](#getting-started---building-from-the-source)
- [Documentation](#documentation)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

ShareNix is a ShareX clone for Linux coded in Go. It features image/screenshot 
and file uploading to almost any file/image sharing service that has a public 
API thanks to the easily customizable json configuration.

ShareNix is only available as a command-line interface for now, but it will soon 
have a GUI to manage your settings and view your upload history.

ShareNix uses the same configuration format as ShareX. If you're a ShareX user, 
you can easily import your settings by pasting them in the Services section of 
sharenix.json. 

Feature progress
============
* Parsing ShareX's json config - done
* Parsing regexp tags - done (no named groups)
* Parsing tags in the parameters - done
* File upload - done (./sharenix path/to/file)
* Full-screen screenshot - done (./sharenix -m=fs)
* Upload files and images from clipboard - done (./sharenix -m=c)
* Automatically open uploads in browser if requested - done (-o flag)
* Archiving clipboard and screenshot uploads to a local folder - done 
  (saved in ./archive/)
* Upload multiple files from clipboard - WIP
* Upload text from clipboard - done
* URL shortening - done
* Screen region selection - WIP
* Upload progress bar - WIP
* Basic upload history csv file - done (./sharenix -history)
* Grep-able upload history output - done (./sharenix -history | grep helloworld)
* GUI tools for config & history - WIP
* Upload GTK notification - WIP
* Screen recording - WIP

Getting started - Prebuilt binaries
============
If you're on amd64 or i386 you can get the pre-built binaries in the release 
section.
The binaries were built on Ubuntu x64 and should be stand-alone, but make sure 
that you have >=gtk-3.10 and >=gdk-3.10, those are the only dependencies as 
they aren't pure Go libraries. 

Once you have the binaries, unzip them in a folder and run sharenix like so:

	tar -zxvf sharenix-linux-amd64.tar.gz
	mv sharenix-linux-amd64 ~/.sharenix
	cd ~/.sharenix
	./sharenix -h
	
You can now set-up sharenix any way you like: bind it to hotkeys, symlink it 
in /usr/bin to launch it from your terminal, and so on.
Check out the sharenix.json config file for the example configuration (which 
works out of the box and uploads to my file sharing service at hnng.moe).
More info on the regex parsing will be coming soon, but the behaviour is nearly 
the same as ShareX so you could just read through 
[this section](https://github.com/ShareX/ShareX/wiki/Custom%20Uploader) of 
the ShareX guide.

Screenshotting areas or windows
============
Until actual area/window grabbing is implemented, you can hack together region 
and window screenshotting using these bash scripts (xclip and gnome-screenshot 
are required).
Remember to replace loli with your username.

sharenix-section.sh

	#!/bin/sh

	# take a screenshot using gnome-screenshot
	sharenixtmp=$(mktemp /tmp/nicememe.XXXXXXXXXXXXXXXXXXX.png)
	xclip -i -selection clipboard -t text/uri-list $sharenixtmp
	gnome-screenshot -a -f $sharenixtmp

	# check file size (0 bytes means that gnome-screenshot was cancelled)
	sharenixtmpsize=$(wc -c <"$sharenixtmp")
	if [ $sharenixtmpsize != 0 ]; then
		/home/loli/.sharenix/sharenix -o $sharenixtmp
	fi
	
sharenix-window.sh

	#!/bin/sh

	# take a screenshot using gnome-screenshot
	sharenixtmp=$(mktemp /tmp/nicememe.XXXXXXXXXXXXXXXXXXX.png)
	xclip -i -selection clipboard -t text/uri-list $sharenixtmp
	gnome-screenshot -w -f $sharenixtmp

	# check file size (0 bytes means that gnome-screenshot was cancelled)
	sharenixtmpsize=$(wc -c <"$sharenixtmp")
	if [ $sharenixtmpsize != 0 ]; then
		/home/loli/.sharenix/sharenix -o $sharenixtmp
	fi
	
You can bind them to hotkeys in CompizConfig Settings Manager under commands 
like so:

![](http://hnng.moe/f/35d)

![](http://hnng.moe/f/35e)

Getting started - Building from the source
============
Before we start building ShareNix, you will need to set up a few dependencies.
* Make sure that you have >=gtk-3.10 and >=gdk-3.10. 
* Get the dev headers for glib, cairo, pango and gtk3. On Ubuntu 15.04, the 
  required packages are: libglib2.0-dev, libcairo-dev, libpango1.0-dev
  and libgtk-3-dev.
* Make sure that you have go >=1.3.1
* Install my fork of gotk3 by running


		go get -tags gtk_3_10 github.com/Francesco149/gotk3/gtk

	
* Remove the clean gotk3 installation and move my fork of gotk3 to 
  the original gotk3 directory with


		rm -r -f $GOPATH/src/github.com/conformal/
		rm -r $GOPATH/pkg/linux_amd64/github.com/conformal/
		rm -r $GOPATH/pkg/linux_386/github.com/conformal/
		mkdir $GOPATH/src/github.com/conformal/
		mv $GOPATH/src/github.com/Francesco149/gotk3 $GOPATH/src/github.com/conformal/gotk3

	
* Get xgb by running

		go get github.com/BurntSushi/xgb
	
* Get osext

		go get github.com/kardianos/osext

Once you've done that, all that's left is to clone the repository.
Make sure that you have git and go installed and run

    go get github.com/Francesco149/sharenix


You can also manually clone the repository anywhere you want by running

    git clone https://github.com/Francesco149/sharenix.git
    

To build sharenix, simply run

	go install -tags gtk_3_10 github.com/Francesco149/sharenix
	
and copy the default config file to $GOPATH/bin

	cp $GOPATH/src/github.com/Francesco149/sharenix/sharenix.json $GOPATH/bin/sharenix.json 
	
then run it (in this example I'm going to be uploading a full-screen screenshot 
to the default site)

	cd $GOPATH/bin
	./sharenix -m=fs
    
Documentation
============
To see a list of the available options, run
	./sharenix -h

You can get the code documentation with the built-in godoc 

    godoc github.com/Francesco149/sharenix
    
If you're looking for a specific function or type just use

    godoc github.com/Francesco149/sharenix MyFunction
    

