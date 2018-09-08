package viewer

import (
	"image"
	"log"
	"syscall"
	"unsafe"

	"github.com/nfnt/resize"
)

var (
	gdi32 = syscall.NewLazyDLL("gdi32.dll")

	pPatBlt        = gdi32.NewProc("PatBlt")
	pStretchDIBits = gdi32.NewProc("StretchDIBits")
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

func getModuleHandle() (syscall.Handle, error) {
	ret, _, err := pGetModuleHandleW.Call(uintptr(0))
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	pCreateWindowExW  = user32.NewProc("CreateWindowExW")
	pDefWindowProcW   = user32.NewProc("DefWindowProcW")
	pDestroyWindow    = user32.NewProc("DestroyWindow")
	pDispatchMessageW = user32.NewProc("DispatchMessageW")
	pGetMessageW      = user32.NewProc("GetMessageW")
	pLoadCursorW      = user32.NewProc("LoadCursorW")
	pPostQuitMessage  = user32.NewProc("PostQuitMessage")
	pRegisterClassExW = user32.NewProc("RegisterClassExW")
	pTranslateMessage = user32.NewProc("TranslateMessage")

	pBeginPaint = user32.NewProc("BeginPaint")
	pEndPaint   = user32.NewProc("EndPaint")

	pGetWindowRect = user32.NewProc("GetWindowRect")
)

const (
	cSW_SHOW        = 5
	cSW_USE_DEFAULT = 0x80000000
)

/* Ternary raster operations */
const (
	SRCCOPY     = 0x00CC0020 /* dest = source                   */
	SRCPAINT    = 0x00EE0086 /* dest = source OR dest           */
	SRCAND      = 0x008800C6 /* dest = source AND dest          */
	SRCINVERT   = 0x00660046 /* dest = source XOR dest          */
	SRCERASE    = 0x00440328 /* dest = source AND (NOT dest )   */
	NOTSRCCOPY  = 0x00330008 /* dest = (NOT source)             */
	NOTSRCERASE = 0x001100A6 /* dest = (NOT src) AND (NOT dest) */
	MERGECOPY   = 0x00C000CA /* dest = (source AND pattern)     */
	MERGEPAINT  = 0x00BB0226 /* dest = (NOT source) OR dest     */
)

const (
	cWS_MAXIMIZE_BOX = 0x00010000
	cWS_MINIMIZEBOX  = 0x00020000
	cWS_THICKFRAME   = 0x00040000
	cWS_SYSMENU      = 0x00080000
	cWS_CAPTION      = 0x00C00000
	cWS_VISIBLE      = 0x10000000

	cWS_OVERLAPPEDWINDOW = 0x00CF0000
)

type winBool uint32

type rect struct {
	left   uint32
	top    uint32
	right  uint32
	bottom uint32
}

type paint struct {
	hdc         syscall.Handle
	erase       winBool
	rc          rect
	restore     winBool
	incUpdate   winBool
	rgbReserved [32]byte
}

const (
	PATCOPY   = 0x00F00021
	PATPAINT  = 0x00FB0A09
	PATINVERT = 0x005A0049
	DSTINVERT = 0x00550009
	BLACKNESS = 0x00000042
	WHITENESS = 0x00FF0062
)

const (
	BI_RGB       = 0
	BI_RLE8      = 1
	BI_RLE4      = 2
	BI_BITFIELDS = 3
	BI_JPEG      = 4
	BI_PNG       = 5
)

//Common 40 byte header
type bitmapinfo struct {
	size           uint32
	width          int32
	height         int32
	planes         uint16
	bitcount       uint16
	compression    uint32
	sizeimage      uint32
	xpelspermeter  int32
	ypelspermeter  int32
	biclrused      uint32
	biclrimportant uint32
	redmask        uint32
	greenmask      uint32
	bluemask       uint32
	alphamask      uint32
	colorSpaceType uint32
	Endpoints      struct {
		red struct {
			x uint32
			y uint32
			z uint32
		}
		green struct {
			x uint32
			y uint32
			z uint32
		}
		blue struct {
			x uint32
			y uint32
			z uint32
		}
	}
	gammaRed    uint32
	gammaGreen  uint32
	gammaBlue   uint32
	intent      uint32
	profileData uint32
	profileSize uint32
	reserved    uint32
}

type bitmapv5header struct {
	bitmapinfo
}

func stretchDIBits(
	hdc syscall.Handle,
	XDest, YDest, nDestWidth, nDestHeight, XSrc, YSrc, nSrcWidth, nSrcHeight int32,
	bits unsafe.Pointer,
	bitsInfo unsafe.Pointer,
	usage uint,
	rop int) (int, error) {
	ret, _, err := pStretchDIBits.Call(
		uintptr(hdc),
		uintptr(XDest),
		uintptr(YDest),
		uintptr(nDestWidth),
		uintptr(nDestHeight),
		uintptr(XSrc),
		uintptr(YSrc),
		uintptr(nSrcWidth),
		uintptr(nSrcHeight),
		uintptr(bits),
		uintptr(bitsInfo),
		uintptr(usage),
		uintptr(rop),
	)
	if ret == 0 {
		return 0, err
	}

	return int(ret), nil
}

func patBlt(hdc syscall.Handle, nXLeft, nYLeft, nWidth, nHeight int, pattern int) (bool, error) {
	ret, _, err := pPatBlt.Call(
		uintptr(hdc),
		uintptr(nXLeft),
		uintptr(nYLeft),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(pattern),
	)
	if int32(ret) == -1 {
		return false, err
	}
	return int32(ret) != 0, nil
}

func beginPaint(hwnd syscall.Handle, pnt *paint) (syscall.Handle, error) {
	ret, _, err := pBeginPaint.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(pnt)),
	)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

func endPaint(hwnd syscall.Handle, pnt *paint) {
	pEndPaint.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(pnt)),
	)
}

func getWindowRect(hwnd syscall.Handle, r *rect) (bool, error) {
	ret, _, err := pGetWindowRect.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(r)),
	)
	if ret == 0 {
		return false, err
	} else {
		return true, nil
	}
}

func createWindow(className, windowName string, style uint32, x, y, width, height int64, parent, menu, instance syscall.Handle) (syscall.Handle, error) {
	ret, _, err := pCreateWindowExW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowName))),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		uintptr(0),
	)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

const (
	cWM_DESTROY           = 0x0002
	cWM_CLOSE             = 0x0010
	cWM_PAINT             = 0x000F
	cWM_WINDOWPOSCHANGING = 0x0046
)

func defWindowProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := pDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		uintptr(wparam),
		uintptr(lparam),
	)
	return uintptr(ret)
}

func destroyWindow(hwnd syscall.Handle) error {
	ret, _, err := pDestroyWindow.Call(uintptr(hwnd))
	if ret == 0 {
		return err
	}
	return nil
}

type point struct {
	x, y int32
}

type message struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

func dispatchMessage(msg *message) {
	pDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

func getMessage(msg *message, hwnd syscall.Handle, msgFilterMin, msgFilterMax uint32) (bool, error) {
	ret, _, err := pGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	if int32(ret) == -1 {
		return false, err
	}
	return int32(ret) != 0, nil
}

const (
	cIDC_ARROW = 32512
)

func loadCursorResource(cursorName uint32) (syscall.Handle, error) {
	ret, _, err := pLoadCursorW.Call(
		uintptr(0),
		uintptr(uint16(cursorName)),
	)
	if ret == 0 {
		return 0, err
	}
	return syscall.Handle(ret), nil
}

func postQuitMessage(exitCode int32) {
	pPostQuitMessage.Call(uintptr(exitCode))
}

const (
	cCOLOR_WINDOW = 5
)

const (
	VREDRAW         = 0x0001
	HREDRAW         = 0x0002
	DBLCLKS         = 0x0008
	OWNDC           = 0x0020
	CLASSDC         = 0x0040
	PARENTDC        = 0x0080
	NOCLOSE         = 0x0200
	SAVEBITS        = 0x0800
	BYTEALIGNCLIENT = 0x1000
	BYTEALIGNWINDOW = 0x2000
	GLOBALCLASS     = 0x4000
)

type tWNDCLASSEXW struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   syscall.Handle
	icon       syscall.Handle
	cursor     syscall.Handle
	background syscall.Handle
	menuName   *uint16
	className  *uint16
	iconSm     syscall.Handle
}

type tWINDOWPOS struct {
	hwnd            syscall.Handle
	hwndInsertAfter syscall.Handle
	x               int
	y               int
	cx              int
	cy              int
	flags           uint
}

func registerClassEx(wcx *tWNDCLASSEXW) (uint16, error) {
	ret, _, err := pRegisterClassExW.Call(
		uintptr(unsafe.Pointer(wcx)),
	)
	if ret == 0 {
		return 0, err
	}
	return uint16(ret), nil
}

func translateMessage(msg *message) {
	pTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
}

var displayBuffer *image.RGBA

func display(img *image.RGBA, name string) {
	displayBuffer = img
	className := "testClass"

	instance, err := getModuleHandle()
	if err != nil {
		log.Println(err)
		return
	}

	cursor, err := loadCursorResource(cIDC_ARROW)
	if err != nil {
		log.Println(err)
		return
	}

	fn := func(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
		switch msg {
		case cWM_CLOSE:
			destroyWindow(hwnd)
		case cWM_DESTROY:
			postQuitMessage(0)
			//case cWM_WINDOWPOSCHANGING:
			//	var t *tWINDOWPOS
			//	t = (*tWINDOWPOS)(unsafe.Pointer(lparam))
			//	log.Println(t)
		case cWM_PAINT:

			var p paint
			deviceContext, err := beginPaint(hwnd, &p)
			if err != nil {
				panic(err)
			}

			var screen rect
			getWindowRect(hwnd, &screen)

			height := int(screen.bottom - screen.top)
			width := int(screen.right - screen.left)

			if width != displayBuffer.Rect.Dx() || height != displayBuffer.Rect.Dy() {
				resized := resize.Resize(uint(width), uint(height), img, resize.NearestNeighbor)
				displayBuffer = resized.(*image.RGBA)
			}

			//log.Println("Planning on redering:",x,y,height,width)
			var binfo bitmapv5header

			binfo.height = -int32(displayBuffer.Rect.Dy()) //Negative height in BMP means Windows will interpret it as having a top left origin
			binfo.width = int32(displayBuffer.Rect.Dx())
			binfo.planes = 1
			binfo.bitcount = 32
			binfo.compression = BI_BITFIELDS
			binfo.redmask = 0x000000FF
			binfo.greenmask = 0x0000FF00
			binfo.bluemask = 0x00FF0000
			binfo.alphamask = 0xFF000000
			binfo.size = uint32(unsafe.Sizeof(binfo))

			//TODO(sjon): figure out proper origin from which to draw the buffer to be scaled, also a proper size would help
			//This code is currently only useful for displaying the initial picture.
			_, err = stretchDIBits(deviceContext,
				0, 0, int32(width), int32(height),
				0, 0, binfo.width, -binfo.height,
				unsafe.Pointer(&displayBuffer.Pix[0]), unsafe.Pointer(&binfo), 0, SRCCOPY)
			if err != nil {
				panic(err)
			}
			endPaint(hwnd, &p)

			////We kinda want a bmp copy to test our sanity!
			//f,err := os.Create("blitted.bmp")
			//if err != nil {
			//	panic(err)
			//}
			//binary.Write(f,binary.LittleEndian, uint16(0x4d42))
			//binary.Write(f,binary.LittleEndian, uint32(14+int(unsafe.Sizeof(binfo))+len(img.Pix)))
			//binary.Write(f,binary.LittleEndian,uint16(0x0000))
			//binary.Write(f,binary.LittleEndian,uint16(0x0000))
			//binary.Write(f,binary.LittleEndian,uint32(unsafe.Sizeof(binfo)+14))
			//
			//binary.Write(f,binary.LittleEndian,binfo)
			//f.WriteAt(img.Pix,int64(unsafe.Sizeof(binfo)+14))
			//f.Close()
			//log.Println(time.Now().Local(), "drawing done")

		default:
			ret := defWindowProc(hwnd, msg, wparam, lparam)
			return ret
		}
		return 0
	}

	wcx := tWNDCLASSEXW{
		style:      HREDRAW | VREDRAW,
		wndProc:    syscall.NewCallback(fn),
		instance:   instance,
		cursor:     cursor,
		background: cCOLOR_WINDOW + 1,
		className:  syscall.StringToUTF16Ptr(className),
	}
	wcx.size = uint32(unsafe.Sizeof(wcx))

	if _, err = registerClassEx(&wcx); err != nil {
		log.Println(err)
		return
	}

	_, err = createWindow(
		className,
		name,
		cWS_VISIBLE|cWS_OVERLAPPEDWINDOW,
		cSW_USE_DEFAULT,
		cSW_USE_DEFAULT,
		int64(img.Rect.Dx())/2, //TODO(sjon): We need to set width and height based on image or screen size. 80% screen might be good?
		int64(img.Rect.Dy())/2,
		0,
		0,
		instance,
	)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		msg := message{}
		gotMessage, err := getMessage(&msg, 0, 0, 0)
		if err != nil {
			log.Println(err)
			return
		}
		if gotMessage {
			translateMessage(&msg)
			dispatchMessage(&msg)
		} else {
			break
		}
	}
}
