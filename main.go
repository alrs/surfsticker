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
	//	"github.com/davecgh/go-spew/spew"
)

var X *xgbutil.XUtil
var sticker string

func init() {
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&sticker, "sticker", "default", "single surf window")
	flag.Parse()
	err = validateSticker(sticker)
	if err != nil {
		log.Fatal(err)
	}
}

func validateSticker(s string) error {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return fmt.Errorf("%s cannot contain punctuation")
		}
	}
	return nil
}

func constructStylePath(s string) string {
	return fmt.Sprintf("~/.surf/styles/%s.css", s)
}

func openURL(w xproto.Window, u string) error {
	return xprop.ChangeProp(X, w, 8, "_SURF_GO", "STRING", []byte(u))
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
	if flag.NArg() != 1 {
		log.Fatalf("surfsticker requires a URL as its argument")
	}
	url := flag.Arg(0)
	surfID, err := findRunningSurf(sticker)
	if surfID == nil {
		surfID, err = startSurf(sticker)
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
