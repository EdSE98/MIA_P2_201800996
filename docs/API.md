# API REST - MIA_P2_201800996

URL local por defecto:

```text
http://127.0.0.1:8080
```

Las respuestas JSON usan este sobre:

```json
{
  "ok": true,
  "message": "operacion completada",
  "data": {}
}
```

Los errores usan:

```json
{
  "ok": false,
  "error": "descripcion del error"
}
```

## Estado y sesion

### GET `/api/health`

Comprueba que la API esta activa.

```json
{
  "ok": true,
  "message": "MIA Proyecto 2 API running"
}
```

### GET `/api/mounts`

Lista las particiones montadas en el proceso actual.

```json
{
  "ok": true,
  "data": [
    {
      "id": "961A",
      "path": "/tmp/d1.dsk",
      "name": "Part1",
      "partitionType": "P",
      "start": 169,
      "size": 15728640
    }
  ]
}
```

### POST `/api/login`

```json
{
  "id": "961A",
  "user": "root",
  "password": "123"
}
```

Respuesta:

```json
{
  "ok": true,
  "message": "sesion iniciada",
  "data": {
    "active": true,
    "mountedId": "961A",
    "user": "root",
    "uid": 1,
    "group": "root",
    "gid": 1
  }
}
```

### POST `/api/logout`

No requiere body.

```json
{
  "ok": true,
  "message": "sesion cerrada"
}
```

## Discos

### GET `/api/disks`

Lista archivos `.dsk` y `.mia` dentro de `MIA_DISKS_DIR`.

```json
{
  "ok": true,
  "data": [
    {
      "name": "d1.dsk",
      "path": "/home/ubuntu/mia/cali/d1.dsk",
      "size": 20971520
    }
  ]
}
```

### POST `/api/disks`

```json
{
  "size": 20,
  "unit": "M",
  "fit": "FF",
  "path": "/home/ubuntu/mia/cali/d1.dsk"
}
```

Respuesta: `disco creado`.

### DELETE `/api/disks`

```json
{
  "path": "/home/ubuntu/mia/cali/d1.dsk"
}
```

Respuesta: `disco eliminado`.

## Particiones

### GET `/api/partitions?path=/ruta/d1.dsk`

Devuelve primarias, extendidas y logicas leidas del MBR/EBR.

```json
{
  "ok": true,
  "data": [
    {
      "name": "Part1",
      "type": "P",
      "fit": "F",
      "start": 169,
      "size": 15728640,
      "status": "1"
    }
  ]
}
```

### POST `/api/partitions`

```json
{
  "path": "/home/ubuntu/mia/cali/d1.dsk",
  "name": "Part1",
  "size": 15,
  "unit": "M",
  "type": "P",
  "fit": "FF"
}
```

Respuesta: `particion creada`.

### PATCH `/api/partitions/resize`

`add` puede ser positivo o negativo.

```json
{
  "path": "/home/ubuntu/mia/cali/d1.dsk",
  "name": "Part1",
  "add": 1,
  "unit": "M"
}
```

Respuesta: `particion redimensionada`.

### DELETE `/api/partitions`

```json
{
  "path": "/home/ubuntu/mia/cali/d1.dsk",
  "name": "Part1",
  "delete": "fast"
}
```

`delete` acepta `fast` o `full`. Respuesta: `particion eliminada`.

## Mount y formato

### POST `/api/mount`

```json
{
  "path": "/home/ubuntu/mia/cali/d1.dsk",
  "name": "Part1"
}
```

Respuesta:

```json
{
  "ok": true,
  "message": "particion montada",
  "data": {
    "id": "961A",
    "path": "/home/ubuntu/mia/cali/d1.dsk",
    "name": "Part1"
  }
}
```

### POST `/api/unmount`

```json
{
  "id": "961A"
}
```

Respuesta: `particion desmontada`.

### POST `/api/mkfs`

```json
{
  "id": "961A",
  "type": "full"
}
```

Respuesta: `particion formateada`.

## Explorador EXT2

### GET `/api/fs/list?id=961A&path=/`

Lista entradas directas de una carpeta.

```json
{
  "ok": true,
  "data": {
    "id": "961A",
    "path": "/",
    "items": [
      {
        "name": "users.txt",
        "path": "/users.txt",
        "type": "file",
        "size": 27,
        "inode": 1,
        "permissions": "664",
        "owner": "root",
        "group": "root"
      }
    ]
  }
}
```

### GET `/api/fs/read?id=961A&path=/users.txt`

Devuelve el contenido completo de un archivo.

```json
{
  "ok": true,
  "data": {
    "id": "961A",
    "path": "/users.txt",
    "name": "users.txt",
    "content": "1,G,root\n1,U,root,root,123\n",
    "size": 27
  }
}
```

### GET `/api/fs/stat?id=961A&path=/users.txt`

Devuelve tipo, inodo, size, permisos, owner y group.

```json
{
  "ok": true,
  "data": {
    "name": "users.txt",
    "type": "file",
    "inode": 1,
    "size": 27,
    "permissions": "664",
    "owner": "root",
    "group": "root"
  }
}
```

## Operaciones EXT2

Todas usan la sesion activa del backend.

### PATCH `/api/fs/edit`

`contenido` es una ruta accesible desde el host del backend.

```json
{
  "path": "/home/docs/a.txt",
  "contenido": "/tmp/nuevo.txt"
}
```

Respuesta: `archivo editado`.

### PATCH `/api/fs/rename`

```json
{
  "path": "/home/docs/a.txt",
  "name": "b1.txt"
}
```

Respuesta: `archivo o carpeta renombrado`.

### DELETE `/api/fs/remove`

```json
{
  "path": "/home/docs/b1.txt"
}
```

Respuesta: `archivo o carpeta eliminado`.

### POST `/api/fs/copy`

```json
{
  "path": "/home/docs/b1.txt",
  "destino": "/home/images"
}
```

Respuesta:

```json
{
  "ok": true,
  "message": "archivo o carpeta copiado",
  "data": {
    "warnings": [
      "sin permiso de lectura: /home/docs/privado.txt"
    ]
  }
}
```

`warnings` se omite cuando todos los descendientes se copian.

### PATCH `/api/fs/move`

```json
{
  "path": "/home/docs/sub",
  "destino": "/home/images"
}
```

Respuesta: `archivo o carpeta movido`.

## Reportes

### POST `/api/reports`

Reportes disponibles: `mbr`, `disk`, `sb`, `inode`, `block`, `bm_inode`,
`bm_block`, `tree`, `ls` y `file`.

```json
{
  "id": "961A",
  "name": "tree",
  "pathFileLs": "/",
  "format": "svg"
}
```

Respuesta:

```json
{
  "ok": true,
  "message": "reporte generado",
  "data": {
    "name": "tree",
    "path": "/home/ubuntu/mia/reportes/tree_961A.svg",
    "url": "/api/report-files/tree_961A.svg",
    "contentType": "image/svg+xml"
  }
}
```

### GET `/api/report-files/{filename}`

Sirve un reporte generado desde `MIA_REPORTS_DIR`.

```text
GET /api/report-files/tree_961A.svg
Content-Type: image/svg+xml
```

El nombre debe ser un archivo simple; rutas absolutas y `..` son rechazados.

## Consola MIA

### POST `/api/commands/execute`

Ejecuta exactamente una linea mediante el parser y dispatcher de la CLI.

```json
{
  "command": "mkdir -p -path=/home/docs"
}
```

Respuesta:

```json
{
  "ok": true,
  "message": "comando ejecutado",
  "data": {
    "command": "mkdir -p -path=/home/docs",
    "output": "Carpeta creada: /home/docs"
  }
}
```

Rechaza comandos vacios, comentarios, `pause`, `exit` y contenido multilinea.
