# Auto Tecla

Herramienta simple para **pulsar teclas automáticamente** a intervalos controlados.

Ideal para tareas repetitivas, juegos o automatizaciones ligeras.

## Características

- Elige cualquier tecla (letras, números, space, enter, flechas, etc.)
- Intervalo personalizable (en segundos)
- Tiempo de pulsación (hold) configurable
- Inicia / detiene con **F6**
- Interfaz sencilla por consola
- Cross-platform (Windows, Linux, macOS)

## Instalación

1. Clona el repositorio:
```bash
git clone https://github.com/valen073625/auto-tecla.git
cd auto-tecla
```

2. Instala las dependencias:
```bash
pip install -r requirements.txt
```

## Uso

```bash
python auto_tecla.py
```

1. Escribe la tecla que quieres que se pulse (ej: `w`, `space`, `enter`, `f`, `left`...)
2. Indica el intervalo en segundos (ej: `0.5` = dos veces por segundo)
3. (Opcional) Tiempo que se mantiene presionada la tecla
4. Presiona **F6** para iniciar / detener la pulsación automática
5. Presiona **Esc** para salir del programa

## Notas

- En algunos juegos con anti-cheat puede no funcionar o ser detectado.
- En Linux puede requerir permisos adicionales según el entorno.
- Usa `pynput` para máxima compatibilidad.

## Ejemplo rápido

```
Tecla: space
Intervalo: 0.1
Hold: 0.05
```
→ Pulsará la barra espaciadora 10 veces por segundo.

---
Hecho con ❤️
