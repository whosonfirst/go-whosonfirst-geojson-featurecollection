package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-geojson-featurecollection/encode"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	geojson_utils "github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-index"
	index_utils "github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/warning"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	var has_properties flags.KeyValueArgs
	flag.Var(&has_properties, "has-property", "Ensure that only features matching `properties.{PROPERTY}={VALUE}` are included. This flag can be passed multiple times.")

	modes := index.Modes()
	str_modes := strings.Join(modes, ",")

	desc := fmt.Sprintf("A valid go-whosonfirst-index mode. Valid modes are: %s", str_modes)

	var mode = flag.String("mode", "repo", desc)
	var out = flag.String("out", "", "Write results to this path. If empty results are written to STDOUT.")
	var spr = flag.Bool("spr", false, "Encode features as a \"standard places response\" (SPR)")

	flag.Parse()

	opts, err := encode.DefaultEncoderOptions()

	if err != nil {
		log.Fatal(err)
	}

	if *out != "" {

		abs_path, err := filepath.Abs(*out)

		if err != nil {
			log.Fatal(err)
		}

		fh, err := os.OpenFile(abs_path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Fatal(err)
		}

		opts.Writer = fh
	}

	opts.SPR = *spr

	enc, err := encode.NewEncoder(opts)

	if err != nil {
		log.Fatal(err)
	}

	feature_ch, done_ch, _, err := enc.Listen()

	f := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		ok, err := index_utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil && !warning.IsWarning(err) {
			return err
		}

		for _, kv := range has_properties {

			path := kv.Key
			value := kv.Value

			if !strings.HasPrefix(path, "properties.") {
				path = fmt.Sprintf("properties.%s", path)
			}

			possible := []string{path}

			prop := geojson_utils.StringProperty(f.Bytes(), possible, "")

			if prop != value {
				return nil
			}
		}

		feature_ch <- f
		return nil
	}

	i, err := index.NewIndexer(*mode, f)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range flag.Args() {

		err := i.IndexPath(path)

		if err != nil {
			log.Fatal(err)
		}
	}

	done_ch <- true
}
