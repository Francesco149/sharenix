ShareNix is a ShareX clone for Linux coded in Go. It features image/screenshot and file 
uploading to almost any file/image sharing service that has a public API thanks to the 
easily customizable json configuration.

ShareNix is only available as a command-line interface for now, but it will soon have 
a GUI to manage your settings and view your upload history.

ShareNix uses the same configuration format as ShareX. If you're a ShareX user, you can 
easily import your settings by pasting them in the Services section of sharenix.json. 

Feature progress
============
* Parsing ShareX's json config - done
* Parsing regexp tags - done
* Parsing tags in the parameters - WIP
* File upload - done (./sharenix path/to/file)
* Full-screen screenshot - done (./sharenix -m=fs)
* Upload files and images from clipboard - done (./sharenix -m=c)
* Automatically open uploads in browser if requested - done (-o flag)
* Archiving clipboard and screenshot uploads to a local folder - done (saved in ./archive/)
* Upload multiple files from clipboard - WIP
* Upload text from clipboard - WIP
* URL shortening - WIP
* Screen region selection - WIP
* Upload progress bar - WIP
* Basic upload history csv file - done (./sharenix -history)
* Grep-able upload history output - done (./sharenix -history | grep helloworld)
* GUI tools for config & history - WIP
* Upload GTK notification - WIP
* Screen recording - WIP

Getting started - Prebuilt x64 binaries
============
If you're on amd64 you can get the pre-built binaries in the release section.
The binaries were built on Gentoo 4.7.3-r1 p1.4, pie-0.5.5 and should be stand-alone, 
but make sure that you have >=gtk-3.10 and >=gdk-3.10, those are the only dependencies 
as they aren't pure Go libraries. 

Once you have the binaries, unzip them in a folder and run sharenix like so:

	$ mkdir sharenix
	$ unzip sharenix-0.1a-amd64-gtk_3_10.zip -d sharenix
	Archive:  sharenix-0.1a-amd64.zip
	  inflating: sharenix/sharenix.json  
	  inflating: sharenix/sharenix       
	$ cd sharenix
	$ ./sharenix -h
	
You can now set-up sharenix any way you like: bind it to hotkeys, symlink it 
in /usr/bin to launch it from your terminal, and so on.
Check out the sharenix.json config file for the example configuration (which 
works out of the box and uploads to my file sharing service at hnng.moe).
More info on the regex parsing will be coming soon, but the behaviour is nearly 
the same as ShareX so you could just read through 
[this section](https://github.com/ShareX/ShareX/wiki/Custom%20Uploader) of 
the ShareX guide.

Getting started - Building from the source
============
Before we start building ShareNix, you will need to set up a few dependencies.
* Make sure that you have >=gtk-3.10 and >=gdk-3.10. 
* Make sure that you have go >=1.3.1
* Install my fork of gotk3 by running


		go get -tags gtk_3_10 github.com/Francesco149/gotk3/gtk

	
* Remove the clean gotk3 installation and move my fork of gotk3 to 
  the original gotk3 directory with


		rm -r -f $GOPATH/src/github.com/conformal/
		mkdir $GOPATH/src/github.com/conformal/
		mv $GOPATH/src/github.com/Francesco149/gotk3 $GOPATH/src/github.com/conformal/gotk3
	
	
* Get xgb by running

		go get github.com/BurntSushi/xgb

Once you've done that, all that's left is to clone the repository.
Make sure that you have git and go installed and run

    go get github.com/Francesco149/sharenix


You can also manually clone the repository anywhere you want by running

    git clone https://github.com/Francesco149/sharenix.git
    

To build sharenix, simply run

	go install -tags gtk_3_10 github.com/Francesco149/sharenix
	
and copy the default config file to $GOPATH/bin

	cp $GOPATH/src/github.com/Francesco149/sharenix/sharenix.json $GOPATH/bin/sharenix.json 
	
then run it (in this example I'm going to be uploading a full-screen screenshot to the default site)

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
    

