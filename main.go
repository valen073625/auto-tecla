package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procSendInput        = user32.NewProc("SendInput")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

const (
	inputKeyboard  = 1
	keyeventfKeyup = 0x0002
	vkF6           = 0x75
	vkEscape       = 0x1B
)

type keybdInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

// Windows INPUT is 40 bytes on amd64 (4 type + 4 pad + 32-byte union).
type winInput struct {
	inputType uint32
	_         [4]byte
	ki        keybdInput
	_         [12]byte
}

func sendKey(vk uint16, up bool) {
	flags := uint32(0)
	if up {
		flags = keyeventfKeyup
	}
	inp := winInput{
		inputType: inputKeyboard,
		ki: keybdInput{
			wVk:     vk,
			dwFlags: flags,
		},
	}
	procSendInput.Call(1, uintptr(unsafe.Pointer(&inp)), unsafe.Sizeof(inp))
}

func keyDown(vk int) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

func parseKey(name string) (uint16, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	special := map[string]uint16{
		"space": 0x20, "enter": 0x0D, "return": 0x0D, "tab": 0x09,
		"esc": 0x1B, "escape": 0x1B, "shift": 0x10, "ctrl": 0x11, "control": 0x11, "alt": 0x12,
		"up": 0x26, "down": 0x28, "left": 0x25, "right": 0x27,
		"backspace": 0x08, "delete": 0x2E, "home": 0x24, "end": 0x23,
		"pageup": 0x21, "page_up": 0x21, "pagedown": 0x22, "page_down": 0x22,
		"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73, "f5": 0x74, "f6": 0x75,
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

func readLine(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

func main() {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("       AUTO TECLA - Pulsación automática")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	in := bufio.NewReader(os.Stdin)
	keyName := readLine(in, "Tecla a pulsar (ej: w, space, enter, f): ")
	vk, ok := parseKey(keyName)
	if !ok {
		fmt.Println("Tecla no reconocida.")
		return
	}

	interval := 0.5
	if s := readLine(in, "Intervalo en segundos (ej: 0.5): "); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
			interval = v
		}
	}
	hold := 0.05
	if s := readLine(in, "Tiempo de pulsación/hold (ej: 0.05): "); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil && v >= 0 {
			hold = v
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Tecla: %s\nIntervalo: %gs\nHold: %gs\n", keyName, interval, hold)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()
	fmt.Println("Controles:")
	fmt.Println("  F6  -> Iniciar / Detener")
	fmt.Println("  Esc -> Salir")
	fmt.Println()
	fmt.Println("Cambia a la ventana donde quieres las pulsaciones.")
	fmt.Println("Esperando... (presiona F6 cuando estés listo)")

	running := false
	prevF6 := false
	prevEsc := false
	next := time.Now()

	for {
		f6 := keyDown(vkF6)
		if f6 && !prevF6 {
			running = !running
			if running {
				fmt.Println("\n>>> Auto Tecla INICIADO <<<")
				next = time.Now()
			} else {
				fmt.Println("\n>>> Auto Tecla DETENIDO <<<")
			}
		}
		prevF6 = f6

		esc := keyDown(vkEscape)
		if esc && !prevEsc {
			fmt.Println("\nSaliendo...")
			return
		}
		prevEsc = esc

		if running && !time.Now().Before(next) {
			sendKey(vk, false)
			if hold > 0 {
				time.Sleep(time.Duration(hold * float64(time.Second)))
			}
			sendKey(vk, true)
			gap := interval - hold
			if gap < 0.01 {
				gap = 0.01
			}
			next = time.Now().Add(time.Duration(gap * float64(time.Second)))
		}

		time.Sleep(15 * time.Millisecond)
	}
}
