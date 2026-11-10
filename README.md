# KPTUI

Gestor de bases KeePass para la terminal, escrito en Go. Abre archivos `.kdbx`,
muestra los grupos como un árbol y deja crear, editar y borrar entradas.

## Cómo se usa

Necesita Go 1.26 o superior.

```bash
go build -o kptui.exe .   # Windows
go run .                  # o directamente
go test ./...             # los tests
```

La app busca los archivos `.kdbx` **en la carpeta desde la que la ejecutás**, así
que hay que correrla desde la carpeta donde están tus bases (o copiarlas ahí).

En la primera pantalla: `Enter` abre la base y pide la contraseña maestra,
`Ctrl+N` (o `Enter` sobre "Create new database") crea una base nueva, `q` sale.

## Teclas

### Navegador (grupos, entradas y detalle)

| Tecla | Qué hace |
|---|---|
| `/` | Barra de comandos (se escribe el comando, `Enter` lo ejecuta, `Esc` cancela) |
| `Ctrl+S` | Guarda la base |
| `Tab` | Va de los grupos a las entradas |
| `⇧Tab` / `Esc` | Vuelve de las entradas a los grupos |
| `q` / `Ctrl+C` | Sale (pregunta antes si hay cambios sin guardar) |

### Grupos (panel izquierdo)

| Tecla | Qué hace |
|---|---|
| `↑` `↓` / `k` `j` | Mueve el cursor |
| `Enter` | Pasa a las entradas del grupo |
| `n` | Nuevo grupo |
| `r` | Renombra el grupo |
| `d` | Borra el grupo (pregunta y dice cuántos grupos y entradas tiene) |
| `i` | Cambia el ícono |

### Entradas (panel derecho)

| Tecla | Qué hace |
|---|---|
| `Enter` | Abre el detalle |
| `n` | Nueva entrada |
| `d` | Borra la entrada (pregunta) |
| `i` | Cambia el ícono |
| `Esc` | Limpia la búsqueda, o vuelve a los grupos |

### Detalle de una entrada

| Tecla | Qué hace |
|---|---|
| `c` | Copia el password al portapapeles |
| `e` | Edita la entrada |
| `i` | Cambia el ícono |
| `p` | Muestra u oculta el password |
| `Esc` | Vuelve |

### Formularios, confirmaciones e íconos

| Tecla | Qué hace |
|---|---|
| `Tab` / `⇧Tab` | Siguiente o anterior campo |
| `Enter` | Guarda o acepta |
| `Esc` | Cancela |
| `Ctrl+P` | Muestra u oculta la contraseña (desbloqueo y base nueva) |
| `Y` / `N` | Confirma o rechaza un borrado |
| `S` / `D` / `C` | Guardar y salir, descartar o cancelar (al salir con cambios) |

## Comandos

Se escriben con `/` en la barra de arriba:

| Comando | Qué hace |
|---|---|
| `/search <texto>` | Busca entradas en todos los grupos |
| `/copy` | Copia el password de la entrada seleccionada |
| `/save` | Guarda la base |
| `/new-entry` | Nueva entrada en el grupo seleccionado |
| `/new-group` | Nuevo grupo dentro del grupo seleccionado |
| `/rename` | Renombra el grupo seleccionado (con el foco en los grupos) |
| `/edit` | Edita la entrada seleccionada |
| `/delete` | Borra el grupo o la entrada según el foco (pregunta) |
| `/icon` | Cambia el ícono |
| `/lock` | Cierra la base y vuelve a la pantalla de desbloqueo |
| `/quit` | Sale (pregunta antes si hay cambios sin guardar) |

## Detalles que conviene saber

- La base se escribe primero en un archivo temporal y después se renombra sobre
  el original, así una save interrumpida no la destruye.
- En algunos terminales de Windows `Ctrl+S` congela la salida (el terminal manda
  XOFF): usá `/save` en ese caso.
- Los nombres de grupo muy largos se cortan con `…` en el árbol para que no se
  rompan los recuadros.
