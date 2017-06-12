<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Feature progress](#feature-progress)
- [Getting started - Prebuilt binaries](#getting-started---prebuilt-binaries)
- [Notifications and canceling uploads](#notifications-and-canceling-uploads)
- [Screenshotting areas or windows](#screenshotting-areas-or-windows)
- [Getting started - Building from the source](#getting-started---building-from-the-source)
- [Plugins](#plugins)
- [Using a Plugin](#using-a-plugin)
- [Writing a Plugin](#writing-a-plugin)
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

[Demonstration video](http://hnng.moe/f/3CI).

Feature progress
============
* Parsing ShareX's json config - done
* Parsing regexp tags - done (no named groups)
* Parsing tags in the parameters - done
* JSON syntax ```$json:some.json.field$``` - done
* XML syntax ```$xml:/root/some/xml/field$``` - done (untested)
* Custom Headers - done
* File upload - done (./sharenix path/to/file)
* Full-screen screenshot - done (./sharenix -m=fs)
* Upload files and images from clipboard - done (./sharenix -m=c)
* Automatically open uploads in browser if requested - done (-o flag)
* Archiving clipboard and screenshot uploads to a local folder - done
  (saved in ./archive/)
* Plugin system - done (still very early)
* Upload multiple files from clipboard - WIP
* Upload text from clipboard - done
* URL shortening - done
* Screen region selection - WIP
* Upload progress bar - WIP
* Basic upload history csv file - done (./sharenix -history)
* Grep-able upload history output - done (./sharenix -history | grep helloworld)
* GUI tools for config & history - WIP
* Clickable GTK notifications - done (-n flag)
* Screen recording - WIP

Getting started - Prebuilt binaries
============
If you're on amd64 or i386 you can get the pre-built binaries in the release
section.
The binaries were built on Ubuntu x64 and should be stand-alone, but make sure
that you have >=gtk-3.10 and >=gdk-3.10, those are the only dependencies as
they aren't pure Go libraries.

Once you have the binaries, unzip them in a folder and run sharenix like so:

```bash
tar -zxvf sharenix-linux-amd64.tar.gz
mv sharenix-linux-amd64 ~/.sharenix
cd ~/.sharenix
./sharenix -h
```

You can now set-up sharenix any way you like: bind it to hotkeys, symlink it
in /usr/bin to launch it from your terminal, and so on.
Check out the sharenix.json config file for the example configuration (which
works out of the box and uploads to my file sharing service at hnng.moe).
More info on the regex parsing will be coming soon, but the behaviour is nearly
the same as ShareX so you could just read through
[this section](https://getsharex.com/docs/custom-uploader) of
the ShareX guide.

Notifications and canceling uploads
============
Using the -n flag will enable notifications for uploads in the bottom right
corner of your screen.

Right-clicking a "upload in progress" notification will cancel the upload and
dismiss the notification.

Right-click a "upload completed" notification will dimiss it, while left
clicking the url will open it in your default browser.

Screenshotting areas or windows
============
Until actual area/window grabbing is implemented, you can hack together region
and window screenshotting using these bash scripts (xclip and gnome-screenshot
are required).
Remember to replace loli with your username.

sharenix-section.sh

```bash
#!/bin/sh

# take a screenshot using gnome-screenshot
sharenixtmp=$(mktemp /tmp/nicememe.XXXXXXXXXXXXXXXXXXX.png)
xclip -i -selection clipboard -t text/uri-list $sharenixtmp
gnome-screenshot -a -f $sharenixtmp

# check file size (0 bytes means that gnome-screenshot was cancelled)
sharenixtmpsize=$(wc -c <"$sharenixtmp")
if [ $sharenixtmpsize != 0 ]; then
    /home/loli/.sharenix/sharenix -n $sharenixtmp
fi
```

sharenix-window.sh

```bash
#!/bin/sh

# take a screenshot using gnome-screenshot
sharenixtmp=$(mktemp /tmp/nicememe.XXXXXXXXXXXXXXXXXXX.png)
xclip -i -selection clipboard -t text/uri-list $sharenixtmp
gnome-screenshot -w -f $sharenixtmp

# check file size (0 bytes means that gnome-screenshot was cancelled)
sharenixtmpsize=$(wc -c <"$sharenixtmp")
if [ $sharenixtmpsize != 0 ]; then
    /home/loli/.sharenix/sharenix -n $sharenixtmp
fi
```

Now you can bind these scripts to hotkeys using whatever configuration
your DE/Window Manager has.

On ubuntu and similar distros, you can bind them to hotkeys in CompizConfig
Settings Manager under commands like so:

![](http://hnng.moe/f/3CQ)

![](http://hnng.moe/f/3CR)

JWM example config (```~/.jwmrc```):
```xml
    <Key mask="CS" key="2">exec:sharenix -m="fs" -n -o</Key>
    <Key mask="CS" key="3">exec:/path/to/sharenix-window.sh</Key>
    <Key mask="CS" key="4">exec:/path/to/sharenix-section.sh</Key>
    <Key mask="CS" key="5">exec:sharenix -m="c" -n</Key>
    <Key mask="CS" key="i">exec:sharenix -m="c" -n -s="twitter (gweet)"</Key>
```

i3wm example config (```~/.i3/config```):
```
bindsym Ctrl+Shift+2 exec sharenix -m="fs" -n -o
bindsym Ctrl+Shift+3 exec /path/to/sharenix-window.sh
bindsym Ctrl+Shift+4 exec /path/to/sharenix-section.sh
bindsym Ctrl+Shift+5 exec sharenix -m="c" -n
bindsym Ctrl+Shift+i exec sharenix -m="c" -s="twitter" -n
```

fluxbox example config (```~/.fluxbox/keys```):
```
Control Shift 2 :Exec sharenix -m="fs" -n -o
Control Shift 3 :Exec /home/loli/scr/sharenix-window.sh
Control Shift 4 :Exec /home/loli/scr/sharenix-section.sh
Control Shift 5 :Exec sharenix -m="c" -n
Control Shift i :Exec sharenix -m="c" -n -s="twitter (gweet)"
```

Getting started - Building from the source
============
NOTE: this codebase is quite outdated (it was written back in go 1.4 or
something like that. I don't plan on refactoring the code for now. If you
encounter issues while trying to compile it, please downgrade to go 1.7.1 or earlier.

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


* Get the other required libraries

        go get github.com/BurntSushi/xgb
        go get github.com/kardianos/osext
        go get github.com/mvdan/xurls
        go get github.com/ChrisTrenkamp/goxpath
        go get github.com/NodePrime/jsonpath

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

Plugins
============
Sharenix has a very early form of plugins as of 0.3.0a. Feel free to contact me
if you wrote a plugin and want it in this list, but be advised that the plugin
specification is still subject to changes.
* [gweet: Upload to twitter](https://github.com/Francesco149/gweet)

Using a Plugin
============
Plugins come as one executable but might also include some extra files.

Plugin authors are highly advised to provide specific install instructions for
their plugin. I will however provide generic guidelines in this section that
will usually apply to every plugin to a certain extent.

To install a plugin, all you have to do is copy all the plugin's files to the
plugins directory in your sharenix folder (~/.sharenix if you followed the
getting started guide). If the plugins directory doesn't exist, create it.

The plugin authors should always provide an example sharenix.json config entry,
or at least a list of parameters you can use. For a generic example of a config
entry, see the last step of "Writing a Plugin".

Writing a Plugin
============
Sharenix has a very early and basic plugin system that might be subject to
changes as the development progress.
* Each plugin is a stand-alone executable that will be placed in the plugins
  directory. You can write a plugin in any language you like as long as it
  follows the specification.
* The last line of the combined stdout & stderr output is used and parsed as
  the plugin's output.
* Command-line parameters must be [go-style](https://golang.org/pkg/flag).
* The plugin will recieve the sharenix.json Arguments list as command-line
  parameters. Additionally, a special _tail parameter can be used to append
  anonymous arguments at the end of the argument list.
* The sharenix.json config entry should have this format:
  ```json

    {
        "Name": "My Awesome Plugin!",
        "RequestType": "PLUGIN",
        "RequestURL": "executable-name",
        "FileFormName": "",
        "Arguments": {
            "_tail": "$input$",
            "foo": "bar",
            "someflag": "true"
        },
        "ResponseType": "Text",
        "RegexList": [],
        "URL": "",
        "ThumbnailURL": "",
        "DeletionURL": ""
    },

    ```

    which will call executable-name like so:
    ```bash

    executable-name -foo=bar -someflag=true /path/to/file or http://url/to/shorten

    ```

I am well aware that this plugin system lacks security, but defending yourself
from malicious plugins is not hard. Avoid non-opensource plugins at all costs
and if in doubt, ask someone to check a plugin's code or check it yourself.

Documentation
============
To see a list of the available options, run
    ./sharenix -h

You can get the code documentation with the built-in godoc

    godoc github.com/Francesco149/sharenix

If you're looking for a specific function or type just use

    godoc github.com/Francesco149/sharenix MyFunction


