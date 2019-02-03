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
	"encoding/json"
	"io/ioutil"
	"path"
)

// A SiteConfig holds the json ShareX config for a single site
type SiteConfig struct {
	Name         string
	RequestType  string            `json:",omitempty"`
	Headers      map[string]string `json:",omitempty"`
	RequestURL   string
	FileFormName string            `json:",omitempty"`
	Arguments    map[string]string `json:",omitempty"`
	ResponseType string
	RegexList    []string `json:",omitempty"`
	URL          string   `json:",omitempty"`
	ThumbnailURL string   `json:",omitempty"`
	DeletionURL  string   `json:",omitempty"`
	Username     string   `json:",omitempty"`
	Password     string   `json:",omitempty"`
}

// A Config holds the json ShareX config for all sites plus the default upload
// targets
type Config struct {
	DefaultFileUploader  string
	DefaultImageUploader string
	DefaultUrlShortener  string
	XineramaHead         uint32  `json:",omitempty"`
	NotificationTime     float64 `json:",omitempty"`
	NotifyUploading      bool    `json:",omitempty"`
	NotifyCommand        string  `json:",omitempty"`
	ClipboardTime        float64 `json:",omitempty"`
	SaveFolder           string  `json:",omitempty"`
	OrganizedFolders     bool    `json:",omitempty"`
	Services             []SiteConfig
}

// GetServiceByName finds a site config by site name and returns it
func (cfg *Config) GetServiceByName(name string) *SiteConfig {
	for i := 0; i < len(cfg.Services); i++ {
		site := &cfg.Services[i]
		if site.Name == name {
			return site
		}
	}
	return nil
}

/*
	Parse returns the correct SiteConfig for the current mode and sitename

	Upload modes:
		f/file: upload file
		fs/fullscreen: screenshot entire screen and upload
		s/section: select screen region and upload
		c/clipboard: upload clipboard contents
		r/record: record screen region and upload
		u/url: shorten url
*/
func (cfg *Config) Parse(mode string, site string, silent bool) (
	sitecfg *SiteConfig, err error) {

	DebugPrintln("Parsing config...")

	if site == "default" {
		switch mode {
		case "f", "file", "c", "clipboard":
			site = cfg.DefaultFileUploader

		case "fs", "fullscreen":
			site = cfg.DefaultImageUploader

		case "u", "url":
			site = cfg.DefaultUrlShortener

		case "s", "section":
			err = &NotImplementedError{}
		}
	}

	sitecfg = cfg.GetServiceByName(site)
	if sitecfg == nil {
		err = &SiteNotFoundError{site}
	}

	return
}

// IsDefaultSite returns true if the given site name is on if the default
// sites in the config
func (cfg *Config) IsDefaultSite(site string) bool {
	switch site {
	case cfg.DefaultFileUploader,
		cfg.DefaultImageUploader,
		cfg.DefaultUrlShortener:
		return true

	default:
		return false
	}

	return false
}

// HandleFileType tries to find and return the most suitable site to upload the
// file to if a default site is currently selected.
func (cfg *Config) HandleFileType(currentsitecfg *SiteConfig,
	filePath string, silent bool) (sitecfg *SiteConfig, err error) {

	DebugPrintln("Checking filetype...")

	// not a default site, so we're not gonna switch
	if !cfg.IsDefaultSite(currentsitecfg.Name) {
		return currentsitecfg, nil
	}

	newsite := currentsitecfg.Name
	mimeType, err := SniffMimeType(filePath)
	if err != nil {
		return nil, err
	}

	// see if we can find a more suitable site for the filetype
	switch {
	// image upload
	case IsImage(mimeType):
		if newsite != cfg.DefaultImageUploader {
			DebugPrintln("Switching to default image uploader")
			newsite = cfg.DefaultImageUploader
		}

	// not an image - we're handling a file
	case newsite != cfg.DefaultFileUploader:
		DebugPrintln("Switching to default file uploader")
		newsite = cfg.DefaultFileUploader

	default:
		return currentsitecfg, nil
	}

	sitecfg = cfg.GetServiceByName(newsite)
	if sitecfg == nil {
		err = &SiteNotFoundError{newsite}
	}

	return
}

func LoadConfig() (cfg *Config, err error) {
	cfg = &Config{}
	cfg.NotificationTime = 30
	cfg.ClipboardTime = 5

	exeFolder, err := GetExeDir()
	if err != nil {
		return
	}

	cfgName := "sharenix.json"

	cfgPaths := [...]string{
		path.Join(GetHome(), "."+cfgName),
		path.Join(exeFolder, cfgName),
		"/etc/" + cfgName,
	}

	var file []byte

	for _, path := range cfgPaths {
		file, err = ioutil.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	err = json.Unmarshal(file, &cfg)
	return
}
