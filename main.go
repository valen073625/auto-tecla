package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procKeybdEvent       = user32.NewProc("keybd_event")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
	procMapVirtualKey    = user32.NewProc("MapVirtualKeyW")
)

const (
	keyeventfKeyup = 0x0002
	vkF6           = 0x75
	vkEscape       = 0x1B
)

func tap(vk uint16, hold time.Duration) {
	scan, _, _ := procMapVirtualKey.Call(uintptr(vk), 0)
	procKeybdEvent.Call(uintptr(vk), scan, 0, 0)
	if hold > 0 {
		time.Sleep(hold)
	}
	procKeybdEvent.Call(uintptr(vk), scan, keyeventfKeyup, 0)
}

func keyDown(vk int) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return uint16(r)&0x8000 != 0
}

func parseKey(name string) (uint16, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	special := map[string]uint16{
		"space": 0x20, "espacio": 0x20, "enter": 0x0D, "return": 0x0D, "intro": 0x0D, "tab": 0x09,
		"esc": 0x1B, "escape": 0x1B, "shift": 0x10, "ctrl": 0x11, "control": 0x11, "alt": 0x12,
		"up": 0x26, "arriba": 0x26, "down": 0x28, "abajo": 0x28, "left": 0x25, "izquierda": 0x25,
		"right": 0x27, "derecha": 0x27,
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
	fmt.Println("AUTO TECLA")
	fmt.Println("Pulsa teclas en la ventana que tengas enfocada.")
	fmt.Println()

	in := bufio.NewReader(os.Stdin)
	keyName := readLine(in, "Tecla (w, space, enter, f...): ")
	vk, ok := parseKey(keyName)
	if !ok {
		fmt.Println("Tecla no reconocida.")
		fmt.Println("Pulsa Enter para cerrar.")
		readLine(in, "")
		return
	}

	interval := 0.5
	if s := readLine(in, "Intervalo en segundos (0.5): "); s != "" {
		s = strings.ReplaceAll(s, ",", ".")
		if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
			interval = v
		}
	}
	holdSec := 0.03
	if s := readLine(in, "Hold en segundos (0.03): "); s != "" {
		s = strings.ReplaceAll(s, ",", ".")
		if v, err := strconv.ParseFloat(s, 64); err == nil && v >= 0 {
			holdSec = v
		}
	}
	hold := time.Duration(holdSec * float64(time.Second))

	fmt.Println()
	fmt.Printf("Tecla %s | cada %g s | hold %g s\n", keyName, interval, holdSec)
	fmt.Println("1. Deja esta ventana abierta")
	fmt.Println("2. Haz clic en el Bloc de notas (u otra ventana)")
	fmt.Println("3. Pulsa F6 para empezar o parar")
	fmt.Println("4. Pulsa Esc para salir")

	running := false
	prevF6 := false
	prevEsc := false
	var count int
	next := time.Now()

	for {
		f6 := keyDown(vkF6)
		if f6 && !prevF6 {
			running = !running
			if running {
				fmt.Println("\nINICIADO. La tecla va a la ventana enfocada.")
				next = time.Now()
			} else {
				fmt.Printf("\nDETENIDO. Pulsaciones: %d\n", count)
			}
		}
		prevF6 = f6

		if keyDown(vkEscape) && !prevEsc {
			fmt.Println("\nSalir.")
			return
		}
		prevEsc = keyDown(vkEscape)

		if running && !time.Now().Before(next) {
			tap(vk, hold)
			count++
			fmt.Printf("\rPulsaciones: %d   ", count)
			gap := time.Duration(interval*float64(time.Second)) - hold
			if gap < 10*time.Millisecond {
				gap = 10 * time.Millisecond
			}
			next = time.Now().Add(gap)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
