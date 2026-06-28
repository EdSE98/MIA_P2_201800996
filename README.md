# MIA_P2_201800996

Proyecto Fase 2 de Manejo e Implementacion de Archivos. Incluye una CLI en Go,
una API REST y un frontend React para administrar discos virtuales, particiones
y un sistema de archivos EXT2 persistido en archivos `.mia` o `.dsk`.

El carnet del proyecto es `201800996`; los IDs de montaje usan el prefijo `96`.

## Componentes

- CLI: parser case-insensitive, scripts y comandos MIA.
- Backend: API REST en Go preparada para EC2.
- Frontend: React, TypeScript y Vite, preparado para build estatico en S3.
- Persistencia: MBR, EBR y EXT2 escritos directamente en archivos binarios.
- Reportes: DOT y renderizado mediante Graphviz.

## Requisitos

- Go 1.22 o posterior.
- Graphviz (`dot`) para renderizar SVG, PNG, JPG y PDF.
- Node.js 20 o posterior y npm.

Comprobar instalaciones:

```bash
go version
dot -V
node --version
npm --version
```

## Ejecucion local

### CLI

```bash
go run .
```

Ejecutar un script:

```bash
go run . -script=archivo.smia
```

### Backend REST

Las variables de `.env.example` son una referencia; Go no carga el archivo
automaticamente. Se deben exportar en el shell o configurar en el servicio.

```bash
export MIA_API_ADDR=127.0.0.1:8080
export MIA_DISKS_DIR=/home/eduardo/mia/cali
export MIA_REPORTS_DIR=/home/eduardo/parte2/reportes/api
go run ./cmd/server
```

Si `MIA_API_ADDR` no se define, el servidor usa `127.0.0.1:${PORT}` y `PORT`
usa `8080` por defecto.

### Frontend

```bash
cd web
npm install
cp .env.example .env
npm run dev
```

Abrir `http://127.0.0.1:5173`.

Build de produccion:

```bash
cd web
npm run build
```

Los archivos estaticos quedan en `web/dist/`.

## Variables de entorno

Backend:

| Variable | Proposito | Valor local sugerido |
| --- | --- | --- |
| `MIA_API_ADDR` | Direccion y puerto HTTP | `127.0.0.1:8080` |
| `MIA_DISKS_DIR` | Carpeta base para listar discos | `/home/eduardo/mia/cali` |
| `MIA_REPORTS_DIR` | Carpeta de reportes de la API | `/home/eduardo/parte2/reportes/api` |
| `PORT` | Puerto compatible si no hay `MIA_API_ADDR` | `8080` |

Frontend:

| Variable | Proposito | Valor local sugerido |
| --- | --- | --- |
| `VITE_API_BASE_URL` | URL publica del backend | `http://127.0.0.1:8080` |

## Funcionalidad

- Discos: `mkdisk`, `rmdisk`.
- Particiones: creacion, `fdisk -add`, `fdisk -delete`, mount y unmount.
- EXT2: mkfs, login, usuarios, grupos, carpetas y archivos.
- Operaciones: `edit`, `rename`, `remove`, `copy` y `move`.
- Reportes: mbr, disk, sb, inode, block, bitmaps, file, ls y tree.
- Web: explorador EXT2, contenido de archivos, reportes y consola MIA.

## Flujo recomendado

1. Crear o seleccionar un disco.
2. Crear una particion primaria.
3. Montar la particion y conservar el ID generado, por ejemplo `961A`.
4. Formatear con mkfs.
5. Iniciar sesion como `root` con password `123`.
6. Crear carpetas y archivos desde la consola web:

```text
mkdir -p -path=/home/docs
mkfile -path=/home/docs/a.txt -size=20
cat -file=/home/docs/a.txt
```

7. Navegar desde `/` y abrir el archivo.
8. Generar reportes `disk`, `tree`, `inode`, `block` y bitmaps.
9. Probar `edit`, `rename`, `copy`, `move` y `remove`.
10. Desmontar antes de redimensionar o eliminar una particion.

## Validacion

Ejecutar todo:

```bash
./scripts/validate.sh
```

O manualmente:

```bash
GOCACHE=/tmp/go-build-cache go test ./...
GOCACHE=/tmp/go-build-cache go vet ./...
cd web && npm run build
```

## Documentacion

- [API REST](docs/API.md)
- [Despliegue en AWS](docs/AWS_DEPLOY.md)

## Limitaciones controladas

- Nombres EXT2 de maximo 12 bytes.
- Archivos con 12 bloques directos y 16 mediante indirecto simple.
- Directorios con hasta 12 bloques directos.
- Sin apuntadores doble o triple indirectos.
- `move` requiere una ranura libre existente en el directorio destino.
- El estado de mounts y sesion vive en RAM del proceso backend.

## Entrega

No se deben versionar:

- Discos `.mia` o `.dsk`.
- Reportes generados.
- `web/node_modules/` ni `web/dist/`.
- Archivos `.env` con configuracion local.
- Binarios compilados.

No modificar ni incluir como codigo propio la carpeta de referencia
`MIA_Junio2026-main/`.
