package main

import (
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
	wmDestroy = 0x0002
	wmCommand = 0x0111
	wmTimer   = 0x0113
	wmSetFont = 0x0030

	wsOverlappedWindow = 0x00CF0000
	wsVisible          = 0x10000000
	wsChild            = 0x40000000
	wsTabstop          = 0x00010000
	wsBorder           = 0x00800000
	esAutoHScroll      = 0x0080
	wsExClientEdge     = 0x00000200

	swShow     = 5
	idcArrow   = 32512
	colorBtn   = 16
	mbIconWarn = 0x00000030

	idKey      = 101
	idInterval = 102
	idHold     = 103
	idStart    = 104
	idStatus   = 105
	timerID    = 1

	keyeventfKeyup = 0x0002
	vkF6           = 0x75
	vkEscape       = 0x1B
)

type wndClassEx struct {
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

type point struct{ x, y int32 }

type msg struct {
	hwnd    syscall.Handle
	message uint32
	_       uint32
	wparam  uintptr
	lparam  uintptr
	time    uint32
	_       uint32
	pt      point
}

var (
	hwndMain     syscall.Handle
	hwndKey      syscall.Handle
	hwndInterval syscall.Handle
	hwndHold     syscall.Handle
	hwndStart    syscall.Handle
	hwndStatus   syscall.Handle
	font         uintptr

	running bool
	prevF6  bool
	prevEsc bool
	count   int
	nextAt  time.Time
	target  uint16
	hold    time.Duration
	gap     time.Duration

	wndProcCB = syscall.NewCallback(wndProc)
)

func utf16(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		p, _ = syscall.UTF16PtrFromString("")
	}
	return p
}

func textOf(hwnd syscall.Handle) string {
	buf := make([]uint16, 64)
	procGetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func setText(hwnd syscall.Handle, s string) {
	procSetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16(s))))
}

func alert(text string) {
	procMessageBox.Call(uintptr(hwndMain), uintptr(unsafe.Pointer(utf16(text))), uintptr(unsafe.Pointer(utf16("Auto Tecla"))), mbIconWarn)
}

func keyDown(vk int) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return uint16(r)&0x8000 != 0
}

func tap(vk uint16, d time.Duration) {
	scan, _, _ := procMapVirtualKey.Call(uintptr(vk), 0)
	procKeybdEvent.Call(uintptr(vk), scan, 0, 0)
	if d > 0 {
		time.Sleep(d)
	}
	procKeybdEvent.Call(uintptr(vk), scan, keyeventfKeyup, 0)
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

func setStatus(s string) { setText(hwndStatus, s) }

func stop() {
	running = false
	setText(hwndStart, "Iniciar")
	setStatus("Detenido. Pulsaciones: " + strconv.Itoa(count))
}

func start() {
	vk, ok := parseKey(textOf(hwndKey))
	if !ok {
		alert("Escribe una tecla válida: w, space, enter, left...")
		return
	}
	target = vk
	hold = parseSeconds(textOf(hwndHold), 0.03)
	interval := parseSeconds(textOf(hwndInterval), 0.5)
	if interval < 10*time.Millisecond {
		interval = 10 * time.Millisecond
	}
	gap = interval - hold
	if gap < 10*time.Millisecond {
		gap = 10 * time.Millisecond
	}
	count = 0
	running = true
	nextAt = time.Now()
	setText(hwndStart, "Detener")
	setStatus("En marcha. La tecla va a la ventana enfocada.")
}

func toggle() {
	if running {
		stop()
	} else {
		start()
	}
}

func onTimer() {
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

	if running && !time.Now().Before(nextAt) {
		tap(target, hold)
		count++
		setStatus("En marcha. Pulsaciones: " + strconv.Itoa(count))
		nextAt = time.Now().Add(gap)
	}
}

func wndProc(hwnd syscall.Handle, message uint32, wparam, lparam uintptr) uintptr {
	switch message {
	case wmCommand:
		if int(wparam&0xFFFF) == idStart {
			toggle()
		}
		return 0
	case wmTimer:
		onTimer()
		return 0
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(message), wparam, lparam)
	return r
}

func createControl(class, text string, style, ex uint32, x, y, w, h int, id int) syscall.Handle {
	hwnd, _, _ := procCreateWindowEx.Call(
		uintptr(ex),
		uintptr(unsafe.Pointer(utf16(class))),
		uintptr(unsafe.Pointer(utf16(text))),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(hwndMain),
		uintptr(id),
		0, 0,
	)
	if font != 0 && hwnd != 0 {
		procSendMessage.Call(hwnd, wmSetFont, font, 1)
	}
	return syscall.Handle(hwnd)
}

func main() {
	inst, _, _ := procGetModuleHandle.Call(0)
	className := utf16("AutoTeclaWindow")
	cursor, _, _ := procLoadCursor.Call(0, idcArrow)

	wc := wndClassEx{
		size:       uint32(unsafe.Sizeof(wndClassEx{})),
		wndProc:    wndProcCB,
		instance:   syscall.Handle(inst),
		cursor:     syscall.Handle(cursor),
		background: syscall.Handle(colorBtn),
		className:  className,
	}
	if atom, _, _ := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		alert("No se pudo crear la ventana.")
		return
	}

	hwnd, _, _ := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16("Auto Tecla"))),
		wsOverlappedWindow|wsVisible,
		uintptr(200), uintptr(160), uintptr(460), uintptr(340),
		0, 0, inst, 0,
	)
	hwndMain = syscall.Handle(hwnd)

	font, _, _ = procCreateFont.Call(
		uintptr(int64(-16)), 0, 0, 0, 400, 0, 0, 0,
		0, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(utf16("Segoe UI"))),
	)

	label := wsChild | wsVisible
	edit := wsChild | wsVisible | wsTabstop | wsBorder | esAutoHScroll
	createControl("STATIC", "Tecla", label, 0, 24, 28, 100, 22, 0)
	hwndKey = createControl("EDIT", "space", edit, wsExClientEdge, 140, 24, 280, 26, idKey)
	createControl("STATIC", "Cada (segundos)", label, 0, 24, 68, 110, 22, 0)
	hwndInterval = createControl("EDIT", "0.5", edit, wsExClientEdge, 140, 64, 280, 26, idInterval)
	createControl("STATIC", "Mantener (seg)", label, 0, 24, 108, 110, 22, 0)
	hwndHold = createControl("EDIT", "0.03", edit, wsExClientEdge, 140, 104, 280, 26, idHold)
	hwndStart = createControl("BUTTON", "Iniciar", wsChild|wsVisible|wsTabstop, 0, 24, 156, 396, 36, idStart)
	hwndStatus = createControl("STATIC", "Detenido. F6 inicia o para. Cierra la ventana para salir.", label, 0, 24, 210, 400, 48, idStatus)

	procSetTimer.Call(uintptr(hwndMain), timerID, 15, 0)
	procShowWindow.Call(uintptr(hwndMain), swShow)
	procUpdateWindow.Call(uintptr(hwndMain))

	var m msg
	for {
		r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}
