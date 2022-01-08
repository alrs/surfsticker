/*
surfsticker - opens sticky surf windows
Copyright (C) 2020 Lars Lehtonen

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"syscall"
	"time"
	"unicode"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
	"k8s.io/utils/inotify"
)

var X *xgbutil.XUtil
var sticker string

func init() {
	flag.StringVar(&sticker, "sticker", "default", "single surf window")
	flag.Parse()
	// the sticker variable is used both as an xproperty value as well as
	// part of the stylesheet argument to exec.Command when it is executing
	// surf. Keep it simple. Alphanumeric should be plenty-enough.
	err := validateSticker(sticker)
	if err != nil {
		log.Fatal(err)
	}
}

func validateSticker(s string) error {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return fmt.Errorf("%s cannot contain punctuation", s)
		}
	}
	return nil
}

func constructStylePath(s string) string {
	return fmt.Sprintf("~/.surf/styles/%s.css", s)
}

func openURL(w xproto.Window, u string) error {
	changeProp := func(u string) error {
		return xprop.ChangeProp(X, w, 8, "_SURF_GO", "STRING", []byte(u))
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return err
	}
	if parsed.Scheme != "file" {
		return changeProp(u)
	}
	watcher, err := inotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	err = watcher.AddWatch(parsed.Path, syscall.IN_CLOSE)
	if err != nil {
		return err
	}
	err = changeProp(parsed.Path)
	if err != nil {
		return err
	}
	select {
	case <-watcher.Event:
		return nil
	case err := <-watcher.Error:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout watching tempfile: %v", parsed.Path)
	}

}

func findRunningSurf(sticker string) (*xproto.Window, error) {
	var id *xproto.Window
	clientids, err := ewmh.ClientListGet(X)
	if err != nil {
		return id, err
	}
	for _, cid := range clientids {
		surfProp, _ := xprop.GetProperty(X, cid, "_SURF_URI")
		stickerProp, _ := xprop.GetProperty(X, cid, "_STICKER")
		stickerVal := []rune{}
		if stickerProp != nil {
			for _, r := range stickerProp.Value {
				stickerVal = append(stickerVal, rune(r))
			}
		}
		if surfProp != nil && sticker == string(stickerVal) {
			return &cid, nil
		}
	}
	return id, nil
}

func startSurf(sticker string) (*xproto.Window, error) {
	// setup and start surf
	stylePath := constructStylePath(sticker)
	surf := exec.Command("surf", "-w", "-b", "-C", stylePath)
	surfOut, err := surf.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = surf.Start()
	if err != nil {
		return nil, err
	}
	// read xprop id from stdout
	scanner := bufio.NewScanner(surfOut)
	scanner.Scan()
	xidStr := scanner.Text()
	xid, err := strconv.ParseUint(xidStr, 10, 32)
	if err != nil {
		return nil, err
	}
	xprid := xproto.Window(xid)
	err = xprop.ChangeProp(X, xprid, 8, "_STICKER", "STRING", []byte((sticker)))
	return &xprid, err
}

func main() {
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer X.Conn().Close()

	if flag.NArg() < 1 {
		log.Fatalf("surfsticker requires a URL as its argument")
	}
	urlStr := flag.Arg(0)
	surfID, err := findRunningSurf(sticker)
	if err != nil {
		log.Fatalf("findRunningSurf: %v", err)
	}

	if surfID == nil {
		surfID, err = startSurf(sticker)
		if err != nil {
			log.Fatalf("startSurf: %v", err)
		}
	}
	log.Printf("surfID is: %d", *surfID)

	err = openURL(*surfID, urlStr)
	if err != nil {
		log.Fatalf("openURL: %v", err)
	}
}
