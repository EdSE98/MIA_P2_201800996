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
- Los archivos soportan 12 bloques directos y un apuntador simple indirecto de hasta 16 bloques adicionales.
- Los directorios usan hasta 12 bloques directos.
- No se implementan apuntadores doble/triple indirectos.
- Los reportes no requieren sesion activa; usan el ID de montaje.

## Scripts incluidos

- `smoke_final_201800996.smia`: flujo compacto completo para validacion final.
- `aVerSiEstoEsDeTuTalla_IDS_201800996.smia`: copia adaptada del script del auxiliar con IDs reales esperados para el carnet `201800996`.
- `cali1_IDS_201800996.smia` y `cali1_IDS_201800996_nopause.smia`: calificacion parte 1 con IDs reales.
- `cali2_continuacion_IDS_201800996.smia` y `cali2_continuacion_IDS_201800996_nopause.smia`: calificacion parte 2 para continuar despues de cali1.
- `cali2_standalone_IDS_201800996.smia` y `cali2_standalone_IDS_201800996_nopause.smia`: calificacion parte 2 con preparacion minima propia.
