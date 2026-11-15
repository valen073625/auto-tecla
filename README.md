# Auto Tecla

Pulsación automática de teclas. Programa nativo en **Go** para Windows. No usa Python.

## Descargar el .exe

GitHub Actions compila `AutoTecla.exe` en Windows.

1. Abre [Actions](https://github.com/valen073625/auto-tecla/actions/workflows/build-exe.yml)
2. Entra al último run verde (**Build AutoTecla.exe**)
3. En **Artifacts** descarga **AutoTecla-windows**
4. Descomprime y ejecuta `AutoTecla.exe`

También puedes lanzarlo a mano: Actions → Build AutoTecla.exe → **Run workflow**.

Windows puede avisar de SmartScreen porque no está firmado: **Más información → Ejecutar de todas formas**.

## Uso

1. Abre `AutoTecla.exe`
2. Escribe la tecla (`w`, `space`, `enter`, `f`, `left`...)
3. Intervalo en segundos (ej: `0.5`)
4. Tiempo que se mantiene presionada (hold)
5. Cambia a la ventana donde quieres las pulsaciones
6. **F6** para iniciar / detener
7. **Esc** para salir

## Compilar en local (opcional)

Hace falta Go, no Python:

```bash
go build -ldflags="-s -w" -o AutoTecla.exe .
```

En Windows el programa usa `SendInput` de user32.

## Notas

- En algunos juegos con anti-cheat puede no funcionar o ser detectado.
