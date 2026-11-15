# Auto Tecla

Herramienta simple para **pulsar teclas automáticamente** a intervalos controlados.

Ideal para tareas repetitivas, juegos o automatizaciones ligeras.

## Descargar el .exe (sin instalar Python)

GitHub Actions compila `AutoTecla.exe` en Windows cada vez que cambia el script.

1. Abre [Actions](https://github.com/valen073625/auto-tecla/actions/workflows/build-exe.yml)
2. Entra al último run verde (**Build AutoTecla.exe**)
3. En **Artifacts** descarga **AutoTecla-windows**
4. Descomprime y ejecuta `AutoTecla.exe`

También puedes lanzar el build a mano: pestaña Actions → Build AutoTecla.exe → **Run workflow**.

El .exe es de consola. Windows puede avisar de SmartScreen porque no está firmado: **Más información → Ejecutar de todas formas**.

## Características

- Elige cualquier tecla (letras, números, space, enter, flechas, etc.)
- Intervalo personalizable (en segundos)
- Tiempo de pulsación (hold) configurable
- Inicia / detiene con **F6**
- Interfaz sencilla por consola

## Uso del .exe

1. Abre `AutoTecla.exe`
2. Escribe la tecla (`w`, `space`, `enter`, `f`, `left`...)
3. Intervalo en segundos (ej: `0.5`)
4. Tiempo que se mantiene presionada (hold)
5. Cambia a la ventana donde quieres las pulsaciones
6. **F6** para iniciar / detener
7. **Esc** para salir

## Uso desde Python

```bash
git clone https://github.com/valen073625/auto-tecla.git
cd auto-tecla
pip install -r requirements.txt
python auto_tecla.py
```

## Notas

- En algunos juegos con anti-cheat puede no funcionar o ser detectado.
- El build de Windows usa PyInstaller + `pynput` en `windows-latest`.

## Ejemplo rápido

```
Tecla: space
Intervalo: 0.1
Hold: 0.05
```

Pulsa la barra espaciadora 10 veces por segundo.
