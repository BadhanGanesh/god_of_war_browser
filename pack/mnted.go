package pack

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/toc"
	"github.com/mogaika/god_of_war_browser/utils"
)

type TocDriver struct {
	Files     map[string]*toc.File
	Streams   [toc.PARTS_COUNT]*os.File
	Directory string
	Cache     *InstanceCache
}

func (p *TocDriver) GetFileNamesList() []string {
	return getFileNamesListFromTocMap(p.Files)
}

func getFileNamesListFromTocMap(files map[string]*toc.File) []string {
	result := make([]string, len(files))
	i := 0
	for name := range files {
		result[i] = name
		i++
	}
	return result
}

func (p *TocDriver) tocGetFileName() string {
	return filepath.Join(p.Directory, toc.GetTocFileName())
}

func (p *TocDriver) partGetFileName(packNumber int) string {
	return filepath.Join(p.Directory, toc.GenPartFileName(packNumber))
}

func (p *TocDriver) prepareStream(packNumber int) error {
	if p.Streams[packNumber] == nil {
		if f, err := os.Open(p.partGetFileName(packNumber)); err != nil {
			return err
		} else {
			p.Streams[packNumber] = f
		}
	}
	return nil
}

func (p *TocDriver) closeStreams() {
	for i, f := range p.Streams {
		if f != nil {
			f.Close()
		}
		p.Streams[i] = nil
	}
}

func (p *TocDriver) getFile(fileName string) (*toc.File, error) {
	if f, exists := p.Files[fileName]; exists {
		return f, nil
	} else {
		return nil, fmt.Errorf("Cannot find '%s' file in pack", fileName)
	}
}

func (p *TocDriver) GetFile(fileName string) (PackFile, error) {
	return p.getFile(fileName)
}

func (p *TocDriver) GetFileReader(fileName string) (PackFile, *io.SectionReader, error) {
	if f, err := p.getFile(fileName); err == nil {
		for packNumber := range p.Streams {
			for _, enc := range f.Encounters {
				if enc.Pack == packNumber {
					if err := p.prepareStream(packNumber); err != nil {
						log.Printf("WARNING: Cannot open pack stream %d: %v", packNumber, err)
					}
					return f, io.NewSectionReader(p.Streams[packNumber], enc.Start, f.Size()), nil
				}
			}
		}
		return f, nil, fmt.Errorf("Cannot open stream for '%s'", fileName)
	} else {
		return nil, nil, err
	}
}

func (p *TocDriver) GetInstance(fileName string) (interface{}, error) {
	return defaultGetInstanceCachedHandler(p, p.Cache, fileName)
}

func (p *TocDriver) UpdateFile(fileName string, in *io.SectionReader) error {
	defer p.parseTocFile()

	f, err := p.getFile(fileName)
	if err != nil {
		return err
	}
	p.closeStreams()

	var fParts [toc.PARTS_COUNT]*os.File
	var partWriters [toc.PARTS_COUNT]utils.ReaderWriterAt
	defer func() {
		for _, part := range fParts {
			if part != nil {
				part.Close()
			}
		}
	}()
	for iPart := range fParts {
		if part, err := os.OpenFile(p.partGetFileName(iPart), os.O_RDWR, 0666); err == nil {
			fParts[iPart] = part
			partWriters[iPart] = utils.NewReaderWriterAtFromFile(part)
		} else {
			return fmt.Errorf("Cannot open '%s' for writing: %v", p.partGetFileName(iPart), err)
		}
	}

	fToc, err := os.OpenFile(p.tocGetFileName(), os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("Cannot open tocfile '%s' for writing: %v", p.tocGetFileName(), err)
	}
	defer fToc.Close()

	ftocoriginal, _ := ioutil.ReadAll(fToc)
	fToc.Seek(0, os.SEEK_SET)

	err = toc.UpdateFile(bytes.NewReader(ftocoriginal), fToc, partWriters, f, in)

	p.Cache = &InstanceCache{}

	return err
}

func (p *TocDriver) parseTocFile() error {
	if tocStream, err := os.Open(p.tocGetFileName()); err == nil {
		defer tocStream.Close()
		log.Printf("[pack] Parsing toc '%s'", p.tocGetFileName())
		p.Files, _, err = toc.ParseFiles(tocStream)
		return err
	} else {
		return err
	}
}

func NewPackFromToc(gamePath string) (*TocDriver, error) {
	p := &TocDriver{
		Directory: gamePath,
		Cache:     &InstanceCache{},
	}

	return p, p.parseTocFile()
}
