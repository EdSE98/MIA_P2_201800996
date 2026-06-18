# MIA_P1_201800996

Proyecto 1 de Manejo e Implementacion de Archivos. La aplicacion es una CLI en Go que simula discos, particiones y un sistema de archivos EXT2 sobre archivos binarios `.mia` o `.dsk`.

## Requisitos

- Go.
- Graphviz opcional para renderizar reportes a `jpg`, `png`, `pdf` o `svg`. Si `dot` no esta instalado, el programa conserva el archivo `.dot`.

## Ejecucion

Modo interactivo:

```bash
go run .
```

Modo script:

```bash
go run . -script=archivo.smia
go run . -script=aVerSiEstoEsDeTuTalla_IDS_201800996.smia
```

## Tests

```bash
go test ./...
```

Si el cache default de Go no es escribible:

```bash
GOCACHE=/tmp/go-build-cache go test ./...
```

## IDs de montaje

Carnet: `201800996`.

Prefijo de IDs: `96`.

Ejemplos:

- `961A`: primera particion montada del primer disco.
- `962A`: primera particion montada del segundo disco.
- `963A`: primera particion montada del tercer disco.

Los IDs se aceptan sin distinguir mayusculas/minusculas.

## Comandos implementados

- Discos y particiones: `mkdisk`, `rmdisk`, `fdisk`, `mount`, `unmount`.
- Formato EXT2: `mkfs`.
- Sesion: `login`, `logout`.
- Usuarios y grupos: `mkgrp`, `rmgrp`, `mkusr`, `rmusr`, `chgrp`.
- Carpetas y archivos: `mkdir`, `mkfile`, `cat`.
- Reportes: `mbr`, `disk`, `sb`, `inode`, `block`, `bm_inode`, `bm_block`, `bm_bloc`, `file`, `ls`, `tree`.

## Limitaciones controladas

- Los nombres de archivo/carpeta dentro del EXT2 tienen maximo 12 bytes porque `Content.BName` es `[12]byte`.
- Archivos y directorios usan hasta 12 bloques directos.
- Los apuntadores indirectos estan definidos en las estructuras, pero no se crean todavia en las operaciones de archivos/directorios.
- Los reportes no requieren sesion activa; usan el ID de montaje.

## Scripts incluidos

- `smoke_final_201800996.smia`: flujo compacto completo para validacion final.
- `aVerSiEstoEsDeTuTalla_IDS_201800996.smia`: copia adaptada del script del auxiliar con IDs reales esperados para el carnet `201800996`.
