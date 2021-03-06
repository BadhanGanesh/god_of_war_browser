package main

import (
	"flag"
	"log"

	"github.com/mogaika/god_of_war_browser/status"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/vfs"
	"github.com/mogaika/god_of_war_browser/web"

	"github.com/mogaika/god_of_war_browser/drivers/iso"
	"github.com/mogaika/god_of_war_browser/drivers/psarc"
	"github.com/mogaika/god_of_war_browser/drivers/toc"

	_ "github.com/mogaika/god_of_war_browser/pack/txt"
	_ "github.com/mogaika/god_of_war_browser/pack/vag"
	_ "github.com/mogaika/god_of_war_browser/pack/vpk"
	_ "github.com/mogaika/god_of_war_browser/pack/wad"

	_ "github.com/mogaika/god_of_war_browser/pack/wad/anm"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/collision"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/cxt"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/flp"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/gfx"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/inst"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/obj"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/sbk"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/scr"
	_ "github.com/mogaika/god_of_war_browser/pack/wad/txr"
)

func main() {
	var addr, tocpath, dirpath, isopath, psarcpath string
	var gowversion int
	flag.StringVar(&addr, "i", ":8000", "Address of server")
	flag.StringVar(&tocpath, "toc", "", "Path to folder with toc file")
	flag.StringVar(&dirpath, "dir", "", "Path to unpacked wads and other stuff")
	flag.StringVar(&isopath, "iso", "", "Path to iso file")
	flag.StringVar(&psarcpath, "psarc", "", "Path to ps3 psarc file")
	flag.IntVar(&gowversion, "gowversion", 0, "0 - auto, 1 - 'gow1', 2 - 'gow2'")
	flag.Parse()

	var err error
	var rootdir vfs.Directory

	config.SetGOWVersion(config.GOWVersion(gowversion))

	if psarcpath != "" {
		config.SetPlayStationVersion(3)
		f := vfs.NewDirectoryDriverFile(psarcpath)
		if err = f.Open(true); err == nil {
			rootdir, err = psarc.NewPsarcDriver(f)
		}
	} else if isopath != "" {
		f := vfs.NewDirectoryDriverFile(isopath)
		if err = f.Open(false); err == nil {
			var isoDriver *iso.IsoDriver
			if isoDriver, err = iso.NewIsoDriver(f); err == nil {
				rootdir, err = toc.NewTableOfContent(isoDriver)
			}
		}
	} else if tocpath != "" {
		rootdir, err = toc.NewTableOfContent(vfs.NewDirectoryDriver(tocpath))
	} else if dirpath != "" {
		rootdir = vfs.NewDirectoryDriver(dirpath)
		if gowversion == 0 {
			log.Fatalf("You must provide 'gowversion' argument if you use directory driver")
		}
	} else {
		flag.PrintDefaults()
		return
	}

	if err != nil {
		log.Fatalf("Cannot start god of war browser: %v", err)
	}

	status.Info("Starting web server on address '%s'", addr)

	if err := web.StartServer(addr, rootdir, "web"); err != nil {
		log.Fatalf("Cannot start web server: %v", err)
	}

}
