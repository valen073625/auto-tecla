package main

import (
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	procRegisterClassEx  = user32.NewProc("RegisterClassExW")
	procCreateWindowEx   = user32.NewProc("CreateWindowExW")
	procDefWindowProc    = user32.NewProc("DefWindowProcW")
	procGetMessage       = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procLoadCursor       = user32.NewProc("LoadCursorW")
	procSetWindowText    = user32.NewProc("SetWindowTextW")
	procGetWindowText    = user32.NewProc("GetWindowTextW")
	procSendMessage      = user32.NewProc("SendMessageW")
	procSetTimer         = user32.NewProc("SetTimer")
	procMessageBox       = user32.NewProc("MessageBoxW")
	procGetModuleHandle  = kernel32.NewProc("GetModuleHandleW")
	procCreateFont       = gdi32.NewProc("CreateFontW")
	procKeybdEvent       = user32.NewProc("keybd_event")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
	procMapVirtualKey    = user32.NewProc("MapVirtualKeyW")
)

const (
	wmDestroy uintptr = 0x0002
	wmCommand uintptr = 0x0111
	wmTimer   uintptr = 0x0113
	wmSetFont uintptr = 0x0030

	wsOverlappedWindow uintptr = 0x00CF0000
	wsChild            uintptr = 0x40000000
	wsVisible          uintptr = 0x10000000
	wsTabstop          uintptr = 0x00010000
	wsBorder           uintptr = 0x00800000
	esAutoHScroll      uintptr = 0x0080
	wsExClientEdge     uintptr = 0x00000200

	swShow     uintptr = 5
	idcArrow   uintptr = 32512
	colorBtn   uintptr = 16
	mbIconWarn uintptr = 0x30

	idKey      = 101
	idInterval = 102
	idHold     = 103
	idStart    = 104

	timerID uintptr = 1

	keyeventfKeyup uintptr = 0x0002
	vkF6           = 0x75
	vkEscape       = 0x1B
)

type wndClassEx struct {
	cbSize    uint32
	style     uint32
	lpfn      uintptr
	clsExtra  int32
	wndExtra  int32
	hInst     uintptr
	hIcon     uintptr
	hCursor   uintptr
	hbr       uintptr
	menuName  *uint16
	className *uint16
	hIconSm   uintptr
}

type point struct{ x, y int32 }

type winMsg struct {
	hwnd    uintptr
	message uint32
	_       uint32
	wparam  uintptr
	lparam  uintptr
	time    uint32
	_       uint32
	pt      point
}

var (
	hwndMain     uintptr
	hwndKey      uintptr
	hwndInterval uintptr
	hwndHold     uintptr
	hwndStart    uintptr
	hwndStatus   uintptr

	className = wide("AutoTeclaWindow")
	fontName  = wide("Segoe UI")

	running   bool
	prevF6    bool
	prevEsc   bool
	count     int
	nextAt    time.Time
	releaseAt time.Time
	keyHeld   bool
	target    uint16
	holdFor   time.Duration
	gapFor    time.Duration
)

func wide(s string) []uint16 {
	b, err := syscall.UTF16FromString(s)
	if err != nil || len(b) == 0 {
		return []uint16{0}
	}
	return b
}

func ptr(b []uint16) uintptr {
	return uintptr(unsafe.Pointer(&b[0]))
}

func alert(text string) {
	body := wide(text)
	title := wide("Auto Tecla")
	procMessageBox.Call(hwndMain, ptr(body), ptr(title), mbIconWarn)
	runtime.KeepAlive(body)
	runtime.KeepAlive(title)
}

func setText(hwnd uintptr, s string) {
	if hwnd == 0 {
		return
	}
	b := wide(s)
	procSetWindowText.Call(hwnd, ptr(b))
	runtime.KeepAlive(b)
}

func textOf(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	buf := make([]uint16, 128)
	procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	runtime.KeepAlive(buf)
	return syscall.UTF16ToString(buf)
}

func keyDown(vk int) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return uint16(r)&0x8000 != 0
}

func keyEvent(vk uint16, up bool) {
	scan, _, _ := procMapVirtualKey.Call(uintptr(vk), 0)
	flags := uintptr(0)
	if up {
		flags = keyeventfKeyup
	}
	procKeybdEvent.Call(uintptr(vk), scan, flags, 0)
}

func releaseKey() {
	if keyHeld {
		keyEvent(target, true)
		keyHeld = false
	}
}

func parseKey(name string) (uint16, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	special := map[string]uint16{
		"space": 0x20, "espacio": 0x20, "enter": 0x0D, "return": 0x0D, "intro": 0x0D, "tab": 0x09,
		"shift": 0x10, "ctrl": 0x11, "control": 0x11, "alt": 0x12,
		"up": 0x26, "arriba": 0x26, "down": 0x28, "abajo": 0x28,
		"left": 0x25, "izquierda": 0x25, "right": 0x27, "derecha": 0x27,
		"backspace": 0x08, "delete": 0x2E, "home": 0x24, "end": 0x23,
		"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73, "f5": 0x74,
		"f7": 0x76, "f8": 0x77, "f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
	}
	if vk, ok := special[name]; ok {
		return vk, true
	}
	if len(name) == 1 {
		c := name[0]
		if c >= 'a' && c <= 'z' {
			return uint16(c - 'a' + 'A'), true
		}
		if c >= '0' && c <= '9' {
			return uint16(c), true
		}
	}
	return 0, false
}

func parseSeconds(s string, fallback float64) time.Duration {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v < 0 {
		v = fallback
	}
	return time.Duration(v * float64(time.Second))
}

func stop() {
	releaseKey()
	running = false
	setText(hwndStart, "Iniciar")
	setText(hwndStatus, "Detenido. Pulsaciones: "+strconv.Itoa(count))
}

func start() {
	vk, ok := parseKey(textOf(hwndKey))
	if !ok {
		alert("Escribe una tecla válida: w, space, enter, left...")
		return
	}
	target = vk
	holdFor = parseSeconds(textOf(hwndHold), 0.03)
	interval := parseSeconds(textOf(hwndInterval), 0.5)
	if interval < 15*time.Millisecond {
		interval = 15 * time.Millisecond
	}
	gapFor = interval
	if holdFor >= interval {
		holdFor = interval / 2
	}
	count = 0
	running = true
	nextAt = time.Now()
	setText(hwndStart, "Detener")
	setText(hwndStatus, "En marcha. La tecla va a la ventana enfocada.")
}

func toggle() {
	if running {
		stop()
	} else {
		start()
	}
}

func onTimer() {
	defer func() { recover() }()
	now := time.Now()
	if keyHeld && !now.Before(releaseAt) {
		releaseKey()
		nextAt = now.Add(gapFor - holdFor)
		if !nextAt.After(now) {
			nextAt = now.Add(15 * time.Millisecond)
		}
	}
	f6 := keyDown(vkF6)
	if f6 && !prevF6 {
		toggle()
	}
	prevF6 = f6
	esc := keyDown(vkEscape)
	if esc && !prevEsc && running {
		stop()
	}
	prevEsc = esc
	if running && !keyHeld && !now.Before(nextAt) {
		keyEvent(target, false)
		keyHeld = true
		releaseAt = now.Add(holdFor)
		count++
		setText(hwndStatus, "En marcha. Pulsaciones: "+strconv.Itoa(count))
	}
}

func wndProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	switch msg {
	case wmCommand:
		if int(wparam&0xFFFF) == idStart {
			toggle()
			return 0
		}
	case wmTimer:
		onTimer()
		return 0
	case wmDestroy:
		releaseKey()
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProc.Call(hwnd, msg, wparam, lparam)
	return r
}

func createControl(class, text string, style, ex, x, y, w, h, id uintptr) uintptr {
	cls := wide(class)
	txt := wide(text)
	hwnd, _, _ := procCreateWindowEx.Call(ex, ptr(cls), ptr(txt), style, x, y, w, h, hwndMain, id, 0, 0)
	runtime.KeepAlive(cls)
	runtime.KeepAlive(txt)
	return hwnd
}

func main() {
	runtime.LockOSThread()
	cb := syscall.NewCallback(wndProc)

	inst, _, _ := procGetModuleHandle.Call(0)
	cursor, _, _ := procLoadCursor.Call(0, idcArrow)
	wc := wndClassEx{
		cbSize:    uint32(unsafe.Sizeof(wndClassEx{})),
		lpfn:      cb,
		hInst:     inst,
		hCursor:   cursor,
		hbr:       colorBtn,
		className: &className[0],
	}
	atom, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	runtime.KeepAlive(&wc)
	if atom == 0 {
		alert("No se pudo registrar la ventana: " + err.Error())
		return
	}

	title := wide("Auto Tecla")
	hwnd, _, err := procCreateWindowEx.Call(
		0,
		ptr(className),
		ptr(title),
		wsOverlappedWindow|wsVisible,
		200, 160, 460, 340,
		0, 0, inst, 0,
	)
	runtime.KeepAlive(title)
	if hwnd == 0 {
		alert("No se pudo abrir la ventana: " + err.Error())
		return
	}
	hwndMain = hwnd

	height := int64(-16)
	font, _, _ := procCreateFont.Call(uintptr(height), 0, 0, 0, 400, 0, 0, 0, 0, 0, 0, 5, 0, ptr(fontName))
	runtime.KeepAlive(fontName)

	label := wsChild | wsVisible
	edit := wsChild | wsVisible | wsTabstop | wsBorder | esAutoHScroll
	btn := wsChild | wsVisible | wsTabstop
	applyFont := func(h uintptr) {
		if h != 0 && font != 0 {
			procSendMessage.Call(h, wmSetFont, font, 1)
		}
	}
	applyFont(createControl("STATIC", "Tecla", label, 0, 24, 28, 120, 22, 0))
	hwndKey = createControl("EDIT", "space", edit, wsExClientEdge, 150, 24, 270, 26, idKey)
	applyFont(hwndKey)
	applyFont(createControl("STATIC", "Cada (segundos)", label, 0, 24, 68, 120, 22, 0))
	hwndInterval = createControl("EDIT", "0.5", edit, wsExClientEdge, 150, 64, 270, 26, idInterval)
	applyFont(hwndInterval)
	applyFont(createControl("STATIC", "Mantener (seg)", label, 0, 24, 108, 120, 22, 0))
	hwndHold = createControl("EDIT", "0.03", edit, wsExClientEdge, 150, 104, 270, 26, idHold)
	applyFont(hwndHold)
	hwndStart = createControl("BUTTON", "Iniciar", btn, 0, 24, 156, 396, 36, idStart)
	applyFont(hwndStart)
	hwndStatus = createControl("STATIC", "Detenido. F6 inicia o para. Cierra la ventana para salir.", label, 0, 24, 210, 400, 48, 0)
	applyFont(hwndStatus)

	procSetTimer.Call(hwndMain, timerID, 15, 0)
	procShowWindow.Call(hwndMain, swShow)
	procUpdateWindow.Call(hwndMain)

	var m winMsg
	for {
		r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}
