# go-whosonfirst-geojson-fearturecollection

Tools for working with GeoJSON FeatureCollections and Who's On First documents.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.7 so let's just assume you need [Go 1.10](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

Error handling has been removed, in the examples below, for the sake of brevity.

### Simple

Create a GeoJSON `FeatureCollection` from a single GeoJSON file, writing the results to `STDOUT`.

```
package main

import(
	"github.com/whosonfirst/go-whosonfirst-geojson-featurecollection/encode"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
)

func main() {

	f, _ := feature.LoadFeatureFromFile("example.geojson")

	opts, _ := encode.DefaultEncoderOptions()
	enc, _ := encode.NewEncoder(opts)

	enc.WriteFeature(f)
	enc.Close()
}	
```

### Fancy

Create a GeoJSON `FeatureCollection` from all the GeoJSON files processed by a `go-whosonfirst-index` index, optionally writing the results to a file. This example uses the `Listen()` which return a channel for sending features (to encode) to and a "done" channel to signal when there are no more features to encode.

The "done" channel will trigger the encoder's `Close()` method which, in turn, will invoke the underlying writer's `Close()` method. That might be a little too much magic for anyone's good so this behaviour might change. The `Listen()` method might also get a different (better?) name...

```
package main

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-geojson-featurecollection/encode"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"io"
	"log"
)

func main() {

	var mode = flag.String("mode", "repo", "...")
	var out = flag.String("out", "", "...")	
	var spr = flag.Bool("spr", false, "Encode features as a \"standard places response\" (SPR)")

	flag.Parse()

	opts, _ := encode.DefaultEncoderOptions()
	opts.SPR = *spr

	if *out != "" {
		fh, _ := os.OpenFile(*out, os.O_RDWR|os.O_CREATE, 0644)
	   	opts.Writer = fh
	}
	
	enc, _ := encode.NewEncoder(opts)

	// _, _ are an error channel and any error that may have been
	// triggered when invoking the Listen() method, respectively
	
	feature_ch, done_ch, _, _ := enc.Listen()

	f := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		ok, _ := utils.IsPrincipalWOFRecord(fh, ctx)

		if !ok {
			return nil
		}

		f, _ := feature.LoadFeatureFromReader(fh)

		feature_ch <- f
		return nil
	}

	i, _ := index.NewIndexer(*mode, f)

	for _, path := range flag.Args() {
		i.IndexPath(path)
	}

	done_ch <- true
}
```

## Options

### EncoderOptions

```
type EncoderOptions struct {
	SPR    bool
	Writer io.Writer
}
```

## Tools

### wof-encode-featurecollection

```
$> ./bin/wof-encode-featurecollection -h
Usage of ./bin/wof-encode-featurecollection:
  -has-property properties.{PROPERTY}={VALUE}
    	Ensure that only features matching properties.{PROPERTY}={VALUE} are included. This flag can be passed multiple times.
  -mode string
    	A valid go-whosonfirst-index mode. Valid modes are: directory,feature,feature-collection,files,geojson-ls,meta,path,repo,sqlite (default "repo")
  -out string
    	Write results to this path. If empty results are written to STDOUT.
  -spr
    	Encode features as a "standard places response" (SPR)
```

For example:

```
$> ./bin/wof-encode-featurecollection -has-property 'sfomuseum:placetype=gate' /usr/local/data/sfomuseum-data-architecture \
   	| jq '[.["features"][]["properties"]["sfomuseum:placetype"]] | unique'

[
  "gate"
]
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-geojson-v2
* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-whosonfirst-index