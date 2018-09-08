package arw

import (
	"fmt"
	"image/png"
	"os"
	"testing"
	"time"
)

const testFileLocation = "samples"

var samples map[sonyRawFile][]string

func init() {
	os.Chdir(testFileLocation)

	samples = make(map[sonyRawFile][]string)

	samples[raw14] = append(samples[raw14], `Y-a7r-iii-DSC00024`)
	samples[raw14] = append(samples[raw14], `4379231197`)
	samples[raw14] = append(samples[raw14], `4538279284`)
	samples[raw14] = append(samples[raw14], `5132423552`)

	samples[raw12] = append(samples[raw12], `DSC01373`)

	samples[craw] = append(samples[craw], `1`)
}

func TestDecodeA7R3(t *testing.T) {
	samplename := samples[raw14][1]
	testARW, err := os.Open(samplename + ".ARW")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	rw, err := extractDetails(testARW)
	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, rw.length)
	testARW.ReadAt(buf, int64(rw.offset))

	if rw.rawType != raw14 {
		t.Error("Not yet implemented type:", rw.rawType)
	}

	start := time.Now()
	t.Log(rw.gammaCurve)
	readRaw14(buf, rw)
	t.Log("processing duration:", time.Now().Sub(start))
}

func TestProcessedPNG(t *testing.T) {
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
	if rw.rawType == raw14 {
		rendered16bit = readRaw14(buf, rw)
	}
	if rw.rawType == craw {
		rendered16bit = readCRAW(buf, rw)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(wd)

	_, err = os.Create("./experiments/bar.png")
	if err != nil {
		t.Error(err)
	}

	const prefix = `16bitPNG`
	f, err := os.Create("experiments/" + prefix + fmt.Sprint(time.Now().Unix()) + ".png")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	png.Encode(f, rendered16bit)
}
