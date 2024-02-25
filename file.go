package main

import (
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	mbOk    = 0x00000000
	mbYesNo = 0x00000004

	mbDefaultIcon1 = 0x00000000
	mbDefaultIcon2 = 0x00000100

	mbIconInfo     = 0x00000040
	mbIconWarning  = 0x00000030
	mbIconError    = 0x00000010
	mbIconQuestion = 0x00000020

	idOk  = 1
	idYes = 6

	swShow       = 5
	swShowNormal = 1
	swUseDefault = 0x80000000

	swpNoZOrder = 0x0004
	swpNoSize   = 0x0001

	smCxScreen = 0
	smCyScreen = 1

	wsThickFrame       = 0x00040000
	wsSysMenu          = 0x00080000
	wsBorder           = 0x00800000
	wsCaption          = 0x00C00000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsMaximizeBox      = 0x00010000
	wsMinimizeBox      = 0x00020000
	wsTabStop          = 0x00010000
	wsGroup            = 0x00020000
	wsOverlappedWindow = 0x00CF0000
	wsExClientEdge     = 0x00000200

	wmCreate     = 0x0001
	wmDestroy    = 0x0002
	wmClose      = 0x0010
	wmCommand    = 0x0111
	wmSetFont    = 0x0030
	wmKeydown    = 0x0100
	wmInitDialog = 0x0110

	ofnAllowMultiSelect = 0x00000200
	ofnExplorer         = 0x00080000
	ofnFileMustExist    = 0x00001000
	ofnHideReadOnly     = 0x00000004
	ofnOverwriteprompt  = 0x00000002

	esPassword    = 0x0020
	esAutoVScroll = 0x0040
	esAutoHScroll = 0x0080

	bifEditBox        = 0x00000010
	bifNewDialogStyle = 0x00000040

	ccRgbInit  = 0x00000001
	ccFullOpen = 0x00000002

	lbAddString   = 0x0180
	lbGetCurSel   = 0x0188
	lbGetSelCount = 0x0190
	lbGetSelItems = 0x0191
	lbGetItemData = 0x0199
	lbSetItemData = 0x019A

	lbSeparator = "LB_SEP"

	lbsExtendedsel = 0x0800

	dtsUpdown          = 0x0001
	dtsShowNone        = 0x0002
	dtsShortDateFormat = 0x0000
	dtsLongDateFormat  = 0x0004

	dtmFirst         = 0x1000
	dtmGetSystemTime = dtmFirst + 1
	dtmSetSystemTime = dtmFirst + 2

	gdtError = -1
	gdtValid = 0
	gdtNone  = 1

	vkEscape               = 0x1B
	enUpdate               = 0x0400
	bsPushButton           = 0
	colorWindow            = 5
	spiGetNonClientMetrics = 0x0029
	gwlStyle               = -16
	maxPath                = 260
)

// OPENFILENAMEA structure
// https://learn.microsoft.com/tr-tr/windows/win32/api/commdlg/ns-commdlg-openfilenamea?redirectedfrom=MSDN
type openfilenameA struct {
	lStructSize       uint32
	hwndOwner         syscall.Handle
	hInstance         syscall.Handle
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          syscall.Handle
	lpTemplateName    *uint16
	pvReserved        unsafe.Pointer
	dwReserved        uint32
	flagsEx           uint32
}

func utf16PtrFromStrnig(s string) *uint16 {
	b := utf16.Encode([]rune(s))
	return &b[0]
}

var comdlg32 *syscall.LazyDLL = syscall.NewLazyDLL("comdlg32.dll")
var getOpenFileNameW *syscall.LazyProc = comdlg32.NewProc("GetOpenFileNameW")

func getOpenFileName(lpofn *openfilenameA) bool {
	ret, _, _ := getOpenFileNameW.Call(uintptr(unsafe.Pointer(lpofn)), 0, 0)
	return ret != 0
}

func stringFromUtf16Ptr(p *uint16) string {
	b := *(*[maxPath]uint16)(unsafe.Pointer(p))
	r := utf16.Decode(b[:])
	return strings.Trim(string(r), "\x00")
}

// FileMulti displays a file dialog that allows for selecting multiple files. It returns the selected
// files, a bool for success, and an error if it was unable to display the dialog. Filter is a string
// that determines which files should be available for selection in the dialog. Separate multiple file
// extensions by spaces and use "*.extension" format for cross-platform compatibility, e.g. "*.png *.jpg".
// A blank string for the filter will display all file types.
func FileMultiSelect(title, filter string) ([]string, bool, error) {
	out, ok := fileDialog(title, filter, true)

	files := make([]string, 0)

	if !ok {
		return files, ok, nil
	}

	l := strings.Split(out, "\x00")
	if len(l) > 1 {
		for _, p := range l[1:] {
			files = append(files, filepath.Join(l[0], p))
		}
	} else {
		files = append(files, out)
	}

	return files, ok, nil
}

func fileDialog(title string, filter string, multi bool) (string, bool) {
	var ofn openfilenameA
	buf := make([]uint16, maxPath*10)

	t, _ := syscall.UTF16PtrFromString(title)

	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.lpstrTitle = t
	ofn.lpstrFile = &buf[0]
	ofn.nMaxFile = uint32(len(buf))

	if filter != "" {
		ofn.lpstrFilter = utf16PtrFromStrnig(filter)
	}

	flags := ofnExplorer | ofnFileMustExist | ofnHideReadOnly
	if multi {
		flags |= ofnAllowMultiSelect
	}
	ofn.flags = uint32(flags)

	if getOpenFileName(&ofn) {
		return stringFromUtf16Ptr(ofn.lpstrFile), true
	}

	return "", false

}
