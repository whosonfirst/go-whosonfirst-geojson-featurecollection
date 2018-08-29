package encode

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"io"
	"os"
	"sync"
)

type EncoderOptions struct {
	SPR    bool
	Writer io.WriteCloser
}

type Encoder struct {
	mu    *sync.RWMutex
	opts  *EncoderOptions
	count int
}

func DefaultEncoderOptions() (*EncoderOptions, error) {

	fh := os.Stdout

	o := EncoderOptions{
		SPR:    true,
		Writer: fh,
	}

	return &o, nil
}

func NewEncoder(opts *EncoderOptions) (*Encoder, error) {

	mu := new(sync.RWMutex)

	e := &Encoder{
		mu:    mu,
		opts:  opts,
		count: 0,
	}

	err := e.writeHeader()

	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Encoder) Listen() (chan geojson.Feature, chan bool, chan error, error) {

	feature_ch := make(chan geojson.Feature)
	done_ch := make(chan bool)
	err_ch := make(chan error)

	go func() {

		for {
			select {

			case f := <-feature_ch:

				err := e.WriteFeature(f)

				if err != nil {
					err_ch <- err
				}

			case <-done_ch:

				err := e.Close()

				if err != nil {
					err_ch <- err
				}

				return
			default:
				// pass
			}
		}
	}()

	return feature_ch, done_ch, err_ch, nil
}

func (e *Encoder) writeHeader() error {
	return e.writeString(`{"type":"FeatureCollection","features":[`)
}

func (e *Encoder) writeFooter() error {
	return e.writeString(`]}`)
}

func (e *Encoder) writeString(body string) error {
	return e.write([]byte(body))
}

func (e *Encoder) write(body []byte) error {
	_, err := e.opts.Writer.Write(body)
	return err
}

func (e *Encoder) Close() error {
	e.writeFooter()
	return e.opts.Writer.Close()
}

func (e *Encoder) WriteFeature(f geojson.Feature) error {

	var body []byte

	if e.opts.SPR {

		spr, err := f.SPR()

		if err != nil {
			return err
		}

		body, err = json.Marshal(spr)

		if err != nil {
			return err
		}

	} else {

		var stub interface{}
		err := json.Unmarshal(f.Bytes(), &stub)

		if err != nil {
			return err
		}

		body, err = json.Marshal(stub)

		if err != nil {
			return err
		}
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.count += 1

	if e.count > 1 {

		err := e.writeString(`,`)

		if err != nil {
			return err
		}
	}

	err := e.write(body)

	if err != nil {
		return err
	}

	return nil
}
