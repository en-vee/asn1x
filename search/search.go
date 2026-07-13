package search

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/en-vee/asn1x"
	"github.com/en-vee/asn1x/cdrfile"
	"github.com/en-vee/asn1x/decode"
)

// Config configures schema loading and CDR file framing for search.
type Config struct {
	SchemaPath      string
	RootType        string
	DecodeSpecsPath string
	FileHeader      bool
	CDRHeader       bool
	Gzip            bool
}

// Searcher decodes and searches BER-encoded CDR files.
type Searcher struct {
	dec        *asn1x.Decoder
	rootType   string
	fileHeader bool
	cdrHeader  bool
	gzip       bool
}

// GrepOptions configures a search run.
type GrepOptions struct {
	Path           string
	JSONPath       string
	Limit          int
	IncludeRecords bool
}

// Match is a CDR record that satisfied the path filter.
type Match struct {
	File      string `json:"file"`
	RecordNum int    `json:"record_number"`
	Record    any    `json:"record,omitempty"`
}

// NewSearcher loads schema and decode specs and returns a reusable searcher.
func NewSearcher(cfg Config) (*Searcher, error) {
	if cfg.SchemaPath == "" {
		return nil, fmt.Errorf("schema path is required")
	}
	if cfg.RootType == "" {
		return nil, fmt.Errorf("root type is required")
	}

	schemaFile, err := os.Open(cfg.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("open schema: %w", err)
	}
	defer schemaFile.Close()

	mod, err := asn1x.Parse(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	var fieldSpecs map[string]string
	if cfg.DecodeSpecsPath != "" {
		fieldSpecs, err = asn1x.LoadFieldSpecsFile(cfg.DecodeSpecsPath)
		if err != nil {
			return nil, fmt.Errorf("load decode specs: %w", err)
		}
	}

	return &Searcher{
		dec:        asn1x.NewDecoderWithOptions(mod, asn1x.DecodeOptions{FieldSpecs: fieldSpecs}),
		rootType:   cfg.RootType,
		fileHeader: cfg.FileHeader,
		cdrHeader:  cfg.CDRHeader,
		gzip:       cfg.Gzip,
	}, nil
}

// Grep searches files under path for records matching jsonPath.
// Path may be a single file or a directory.
func (s *Searcher) Grep(opts GrepOptions) ([]Match, error) {
	filter, err := decode.ParsePathFilter(opts.JSONPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(opts.Path)
	if err != nil {
		return nil, err
	}

	matches := make([]Match, 0)
	if !info.IsDir() {
		fileMatches, err := s.grepFile(filter, opts.Path, opts.Limit, opts.IncludeRecords)
		if err != nil {
			return nil, err
		}
		return append(matches, fileMatches...), nil
	}

	err = filepath.WalkDir(opts.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		fileMatches, err := s.grepFile(filter, path, opts.Limit, opts.IncludeRecords)
		if err != nil {
			return err
		}
		matches = append(matches, fileMatches...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (s *Searcher) grepFile(filter decode.PathFilter, path string, limit int, includeRecords bool) ([]Match, error) {
	data, err := readFileContent(path, s.gzip)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	fileReader := cdrfile.NewReader(data, cdrfile.Options{
		HasFileHeader: s.fileHeader,
		HasCDRHeader:  s.cdrHeader,
	})

	if s.fileHeader {
		if _, err := fileReader.ReadFileHeader(); err != nil {
			return nil, nil
		}
	}

	matches := make([]Match, 0)
	recordNum := 0
	for fileReader.Remaining() > 0 {
		if limit > 0 && recordNum >= limit {
			break
		}

		_, cdrData, err := fileReader.NextRecord()
		if err != nil {
			break
		}
		recordNum++

		val, err := s.dec.DecodeBytes(s.rootType, cdrData)
		if err != nil {
			continue
		}
		if !filter.Match(val) {
			continue
		}

		match := Match{
			File:      path,
			RecordNum: recordNum,
		}
		if includeRecords {
			match.Record = val
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func readFileContent(path string, gzipEnabled bool) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if !gzipEnabled {
		return io.ReadAll(f)
	}
	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()
	return io.ReadAll(gr)
}
