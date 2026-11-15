#!/usr/bin/env python3
"""
Auto Tecla - Pulsación automática de teclas
Presiona F6 para iniciar/detener | Esc para salir
"""

import time
import threading
from pynput.keyboard import Controller, Key, Listener

# Controlador de teclado
keyboard = Controller()

# Estado global
running = False
stop_event = threading.Event()
target_key = None
interval = 0.5
hold_time = 0.05


def get_key_from_name(name: str):
    """Convierte el nombre de tecla a objeto Key o string."""
    name = name.strip().lower()

    special_keys = {
        "space": Key.space,
        "enter": Key.enter,
        "return": Key.enter,
        "tab": Key.tab,
        "esc": Key.esc,
        "escape": Key.esc,
        "shift": Key.shift,
        "ctrl": Key.ctrl,
        "control": Key.ctrl,
        "alt": Key.alt,
        "up": Key.up,
        "down": Key.down,
        "left": Key.left,
        "right": Key.right,
        "backspace": Key.backspace,
        "delete": Key.delete,
        "home": Key.home,
        "end": Key.end,
        "page_up": Key.page_up,
        "page_down": Key.page_down,
        "caps_lock": Key.caps_lock,
        "f1": Key.f1,
        "f2": Key.f2,
        "f3": Key.f3,
        "f4": Key.f4,
        "f5": Key.f5,
        "f6": Key.f6,
        "f7": Key.f7,
        "f8": Key.f8,
        "f9": Key.f9,
        "f10": Key.f10,
        "f11": Key.f11,
        "f12": Key.f12,
    }

    if name in special_keys:
        return special_keys[name]

    # Teclas de una sola letra/número
    if len(name) == 1:
        return name

    print(f"[!] Tecla '{name}' no reconocida. Usando como texto simple.")
    return name


def press_loop():
    """Bucle principal de pulsación."""
    global running
    while not stop_event.is_set():
        if running and target_key is not None:
            try:
                keyboard.press(target_key)
                time.sleep(hold_time)
                keyboard.release(target_key)
            except Exception as e:
                print(f"[Error] {e}")
            time.sleep(max(0.01, interval - hold_time))
        else:
            time.sleep(0.05)


def on_press(key):
    """Maneja teclas de control (F6 y Esc)."""
    global running

    try:
        if key == Key.f6:
            running = not running
            estado = "INICIADO" if running else "DETENIDO"
            print(f"\n>>> Auto Tecla {estado} <<<")
        elif key == Key.esc:
            print("\nSaliendo...")
            stop_event.set()
            return False  # Detiene el listener
    except AttributeError:
        pass


def main():
    global target_key, interval, hold_time

    print("=" * 50)
    print("       AUTO TECLA - Pulsación Automática")
    print("=" * 50)
    print()

    # Configuración
    key_input = input("Tecla a pulsar (ej: w, space, enter, f): ").strip()
    if not key_input:
        print("Debes indicar una tecla.")
        return

    target_key = get_key_from_name(key_input)

    try:
        interval = float(input("Intervalo en segundos (ej: 0.5): ").strip() or "0.5")
        if interval <= 0:
            interval = 0.5
    except ValueError:
        interval = 0.5

    try:
        hold_time = float(input("Tiempo de pulsación/hold (ej: 0.05): ").strip() or "0.05")
        if hold_time < 0:
            hold_time = 0.05
    except ValueError:
        hold_time = 0.05

    print()
    print("-" * 50)
    print(f"Tecla: {key_input}")
    print(f"Intervalo: {interval}s")
    print(f"Hold: {hold_time}s")
    print("-" * 50)
    print()
    print("Controles:")
    print("  • F6  → Iniciar / Detener")
    print("  • Esc → Salir del programa")
    print()
    print("Cambia a la ventana donde quieres las pulsaciones...")
    print("Esperando... (presiona F6 cuando estés listo)")

    # Iniciar hilo de pulsación
    press_thread = threading.Thread(target=press_loop, daemon=True)
    press_thread.start()

    # Listener de teclas de control
    with Listener(on_press=on_press) as listener:
        listener.join()

    print("Programa finalizado.")


if __name__ == "__main__":
    main()
