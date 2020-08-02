package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
)

var X *xgbutil.XUtil

func init() {
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
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
	surf := exec.Command("surf", "-w", "-b")
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
	if len(os.Args) != 2 {
		log.Fatal("requires URL as an argument")
	}
	url := os.Args[1]
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
