package arw

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
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
	samples[raw12] = append(samples[raw12], `DSC01373`)
	samples[craw] = append(samples[craw], `1`)
	samples[crawLossless] = append(samples[crawLossless], `DSC01373`)
}

func TestDecodeA7R3(t *testing.T) {
	samplename := samples[raw14][0]
	testARW, err := os.Open(samplename + ".ARW")
	if err != nil {
		t.Error(err)
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

	rendered16bit := readraw14(buf, rw)

	asRGBA := image.NewRGBA(rendered16bit.Rect)
	for y := asRGBA.Rect.Min.Y; y < asRGBA.Rect.Max.Y; y++ {
		for x := asRGBA.Rect.Min.X; x < asRGBA.Rect.Max.X; x++ {
			asRGBA.Set(x, y, rendered16bit.At(x, y))
		}
	}

	const prefix = "8bit-"
	os.Chdir("experiments")

	if false {
		jpgName := prefix + fmt.Sprint(time.Now().Unix()) + ".jpg"
		f, err := os.Create(jpgName)
		if err != nil {
			t.Error(err)
		}

		jpeg.Encode(f, asRGBA, nil)

		f.Close()
	}

	if false {
		f, err := os.Create(prefix + fmt.Sprint(time.Now().Unix()) + ".png")
		if err != nil {
			t.Error(err)
		}

		png.Encode(f, asRGBA)

		f.Close()
	}
	if true {
		display(asRGBA) //For some reason the colours are way blown out. Printing 8 bit to a JPG works fine.
	}
}

func TestViewer(t *testing.T) {
	sampleName := `C:\Users\sjon\4f328002-2680-11e5-8616-c525bf19aff7.jpg`
	sample, err := os.Open(sampleName)
	if err != nil {
		t.Error(err)
	}
	img, _, err := image.Decode(sample)
	if err != nil {
		t.Error(err)
	}

	b := img.Bounds()
	rgba := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r32, g32, b32, a32 := img.At(x, y).RGBA()
			c := color.RGBA{uint8(r32), uint8(g32), uint8(b32), uint8(a32)}
			rgba.SetRGBA(x, y, c)
		}
	}

	display(rgba)
}

func TestMetadata(t *testing.T) {
	samplename := samples[raw14][0]
	testARW, err := os.Open(samplename + ".ARW")
	if err != nil {
		t.Error(err)
	}

	header, err := ParseHeader(testARW)
	meta, err := ExtractMetaData(testARW, int64(header.Offset), 0)
	if err != nil {
		t.Error(err)
	}
	t.Log("0th IFD for primary image data")
	t.Log(meta)

	for _, v := range meta.FIA {
		t.Logf("%+v\n", v)
	}

	for _, fia := range meta.FIA {
		if fia.Tag == SubIFDs {
			t.Log("Reading subIFD located at: ", fia.Offset)
			next, err := ExtractMetaData(testARW, int64(fia.Offset), 0)
			if err != nil {
				t.Error(err)
			}
			t.Log("A subIFD, who knows what we'll find here!")
			t.Log(next)
			for _, v := range next.FIA {
				t.Logf("%+v\n", v)
			}

		}

		if fia.Tag == GPSTag {
			gps, err := ExtractMetaData(testARW, int64(fia.Offset), 0)
			if err != nil {
				t.Error(err)
			}

			t.Log("GPS IFD (GPS Info Tag)")
			t.Log(gps)
		}

		if fia.Tag == ExifTag {
			exif, err := ExtractMetaData(testARW, int64(fia.Offset), 0)
			if err != nil {
				t.Error(err)
			}

			t.Log("Exif IFD (Exif Private Tag)")
			t.Log(exif)
			//Just an attempt at understanding these crazy MakerNotes..
			for i := range exif.FIA {
				if exif.FIA[i].Tag == MakerNote {
					makernote, err := ExtractMetaData(bytes.NewReader(*exif.FIAvals[i].ascii), 0, 0)
					if err != nil || makernote.Count == 0 {
						t.Error(err)
					}

					//t.Log("Really stupid propietary makernote structure.")
					//t.Log(makernote)
					//for _,v := range makernote.FIA {
					//	t.Logf("%+v\n",v)
					//}
				}
			}
		}

		if fia.Tag == DNGPrivateData {
			dng, err := ExtractMetaData(testARW, int64(fia.Offset), 0)
			if err != nil {
				t.Error(err)
			}

			t.Log("DNG IFD (RAW metadata)")
			t.Log(dng)

			for _, v := range dng.FIA {
				t.Logf("%+v\n", v)
			}

			for i := range dng.FIA {
				if dng.FIA[i].Tag == IDC_IFD {
					idc, err := ExtractMetaData(testARW, int64(dng.FIA[i].Offset), 0)
					if err != nil {
						t.Error(err)
					}

					t.Log("IDC IFD (RAW metadata)")
					t.Log(idc)

					for _, v := range idc.FIA {
						t.Logf("%+v\n", v)
					}
				}
			}
		}
	}
	first, err := ExtractMetaData(testARW, int64(meta.Offset), 0)
	if err != nil {
		t.Error(err)
	}

	t.Log("First IFD for thumbnail data")
	t.Log(first)
}

func TestNestedHeader(t *testing.T) {
	samplename := samples[raw14][0]
	testARW, err := os.Open(samplename + ".ARW")
	if err != nil {
		t.Error(err)
	}

	meta, err := ExtractMetaData(testARW, 52082, 0)
	if err != nil {
		t.Error(err)
	}
	for _, v := range meta.FIA {
		t.Logf("%+v\n", v)
	}

	var sr2offset uint32
	var sr2length uint32
	var sr2key [4]byte
	for i := range meta.FIA {
		if meta.FIA[i].Tag == SR2SubIFDOffset {
			offset := meta.FIA[i].Offset
			sr2offset = offset
		}
		if meta.FIA[i].Tag == SR2SubIFDLength {
			sr2length = meta.FIA[i].Offset
		}
		if meta.FIA[i].Tag == SR2SubIFDKey {
			key := meta.FIA[i].Offset*0x0edd + 1
			sr2key[3] = byte((key >> 24) & 0xff)
			sr2key[2] = byte((key >> 16) & 0xff)
			sr2key[1] = byte((key >> 8) & 0xff)
			sr2key[0] = byte((key) & 0xff)
		}
	}

	t.Logf("SR2len: %v SR2off: %v SR2key: %v\n", sr2length, sr2offset, sr2key)

	buf := DecryptSR2(testARW, sr2offset, sr2length)
	f, _ := ioutil.TempFile(os.TempDir(), "SR2")
	f.Write(buf)

	br := bytes.NewReader(buf)

	meta, err = ExtractMetaData(br, 0, 0)
	if err != nil {
		t.Error(err)
	}
	t.Log(meta)

	for _, v := range meta.FIA {
		t.Logf("%+v\n", v)
	}
}

func TestJPEGDecode(t *testing.T) {
	testARW, err := os.Open("2.ARW")
	if err != nil {
		t.Error(err)
	}
	header, err := ParseHeader(testARW)
	meta, err := ExtractMetaData(testARW, int64(header.Offset), 0)
	if err != nil {
		t.Error(err)
	}

	var jpegOffset uint32
	var jpegLength uint32
	for i := range meta.FIA {
		switch meta.FIA[i].Tag {
		case JPEGInterchangeFormat:
			jpegOffset = meta.FIA[i].Offset
		case JPEGInterchangeFormatLength:
			jpegLength = meta.FIA[i].Offset
		}
	}
	jpg, err := ExtractThumbnail(testARW, jpegOffset, jpegLength)
	if err != nil {
		t.Error(err)
	}
	reader := bytes.NewReader(jpg)
	img, err := jpeg.Decode(reader)
	if err != nil {
		t.Error(err)
	}

	out, err := os.Create(fmt.Sprint(time.Now().Unix(), "reencoded", ".jpg"))
	if err != nil {
		t.Error(err)
	}
	jpeg.Encode(out, img, nil)
}

func TestJPEG(t *testing.T) {
	testARW, err := os.Open("1.ARW")
	if err != nil {
		t.Error(err)
	}

	header, err := ParseHeader(testARW)
	meta, err := ExtractMetaData(testARW, int64(header.Offset), 0)
	if err != nil {
		t.Error(err)
	}

	var jpegOffset uint32
	var jpegLength uint32
	for i := range meta.FIA {
		switch meta.FIA[i].Tag {
		case JPEGInterchangeFormat:
			jpegOffset = meta.FIA[i].Offset
		case JPEGInterchangeFormatLength:
			jpegLength = meta.FIA[i].Offset
		}
	}

	t.Log("JPEG start:", jpegOffset, " JPEG size:", jpegLength)
	jpg := make([]byte, jpegLength)
	testARW.ReadAt(jpg, int64(jpegOffset))
	out, err := os.Create(fmt.Sprint(time.Now().Unix(), "raw", ".jpg"))
	out.Write(jpg)
}
