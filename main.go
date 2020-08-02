package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"unicode"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
)

var X *xgbutil.XUtil
var style string

func init() {
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&style, "style", "default", "surf stylesheet, no extension")
	flag.Parse()
}

func constructStylePath(s string) (string, error) {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return "", fmt.Errorf("%s cannot contain punctuation")
		}
	}
	return fmt.Sprintf("~/.surf/styles/%s.css", s), nil
}

func openURL(w xproto.Window, u string) error {
	return xprop.ChangeProp(X, w, 8, "_SURF_GO", "STRING", []byte(u))
}

func findRunningSurf() (*xproto.Window, error) {
	var id *xproto.Window
	clientids, err := ewmh.ClientListGet(X)
	if err != nil {
		return id, err
	}
	for _, cid := range clientids {
		prop, _ := xprop.GetProperty(X, cid, "_SURF_URI")
		if prop != nil {
			return &cid, nil
		}
	}
	return id, nil
}

func startSurf() (*xproto.Window, error) {
	// setup and start surf
	safeStyle, err := constructStylePath(style)
	if err != nil {
		return nil, err
	}
	surf := exec.Command("surf", "-w", "-b", "-C", safeStyle)
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
	return &xprid, nil
}

func main() {
	if flag.NArg() != 1 {
		log.Fatalf("surfsticker requires a URL as its argument")
	}
	url := flag.Arg(0)
	surfID, err := findRunningSurf()
	if surfID == nil {
		surfID, err = startSurf()
		if err != nil {
			log.Fatalf("startSurf: %v", err)
		}
	}
	log.Printf("surfID is: %d", *surfID)
	err = openURL(*surfID, url)
	if err != nil {
		log.Fatalf("openURL: %v", err)
	}
}
