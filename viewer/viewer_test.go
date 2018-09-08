package viewer

import (
	"image"
	"os"
	"testing"
	"time"
)

func TestViewer(t *testing.T) {
	sampleName := samples[raw14][0]
	sample, err := os.Open(sampleName + ".ARW")
	if err != nil {
		t.Error(err)
	}

	rw, err := extractDetails(sample)
	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, rw.length)
	sample.ReadAt(buf, int64(rw.offset))

	var rendered16bit *RGB14

	switch rw.rawType {
	case raw14:
		rendered16bit = readRaw14(buf, rw)
	case craw:
		rendered16bit = readCRAW(buf, rw)
	default:
		t.Error("Unhanded RAW type:", rw.rawType)
	}

	asRGBA := image.NewRGBA(rendered16bit.Rect)
	for y := asRGBA.Rect.Min.Y; y < asRGBA.Rect.Max.Y; y++ {
		for x := asRGBA.Rect.Min.X; x < asRGBA.Rect.Max.X; x++ {
			asRGBA.Set(x, y, rendered16bit.At(x, y))
		}
	}

	display(asRGBA, sampleName, rw.lensModel, rw.focalLength, rw.aperture, int(rw.iso), time.Duration(rw.shutter*float32(time.Second)))
}
