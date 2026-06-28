# Despliegue AWS - MIA_P2_201800996

Arquitectura propuesta:

- EC2 Linux: API Go, Graphviz, discos y reportes.
- S3 Static Website Hosting: frontend compilado.

## Backend en EC2

### 1. Crear instancia

Usar Ubuntu Server o Amazon Linux. En el Security Group permitir:

- SSH `22` solo desde la IP administrativa.
- API `8080` desde la IP de pruebas o desde el origen del frontend.

Para una entrega academica puede abrirse temporalmente `8080`, pero no es
recomendable publicar la consola de comandos sin autenticacion adicional.

### 2. Instalar requisitos

Ubuntu:

```bash
sudo apt update
sudo apt install -y graphviz git
```

Instalar una version soportada de Go desde `https://go.dev/dl/` o mediante el
gestor aprobado para la instancia. Verificar:

```bash
go version
dot -V
```

### 3. Clonar y compilar

```bash
git clone URL_DEL_REPOSITORIO
cd MIA_P1_201800996
go build -o mia-p2-api ./cmd/server
```

Crear almacenamiento:

```bash
sudo mkdir -p /home/ubuntu/mia/cali
sudo mkdir -p /home/ubuntu/mia/reportes
sudo chown -R ubuntu:ubuntu /home/ubuntu/mia
```

### 4. Ejecutar

```bash
export MIA_API_ADDR=0.0.0.0:8080
export MIA_DISKS_DIR=/home/ubuntu/mia/cali
export MIA_REPORTS_DIR=/home/ubuntu/mia/reportes
./mia-p2-api
```

Comprobar:

```bash
curl http://127.0.0.1:8080/api/health
curl http://IP_PUBLICA_EC2:8080/api/health
```

### 5. Servicio systemd

Crear `/etc/systemd/system/mia-p2-api.service`:

```ini
[Unit]
Description=MIA Proyecto 2 API
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/MIA_P1_201800996
Environment=MIA_API_ADDR=0.0.0.0:8080
Environment=MIA_DISKS_DIR=/home/ubuntu/mia/cali
Environment=MIA_REPORTS_DIR=/home/ubuntu/mia/reportes
ExecStart=/home/ubuntu/MIA_P1_201800996/mia-p2-api
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Activar:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now mia-p2-api
sudo systemctl status mia-p2-api
sudo journalctl -u mia-p2-api -f
```

## Frontend en S3

### 1. Configurar API publica

Crear `web/.env.production`:

```env
VITE_API_BASE_URL=http://IP_PUBLICA_EC2:8080
```

Si el sitio S3 se sirve mediante HTTPS, el navegador bloqueara una API HTTP
por contenido mixto. Para produccion se recomienda exponer EC2 con HTTPS
mediante Nginx/ALB y un dominio, y usar esa URL en `VITE_API_BASE_URL`.

### 2. Compilar

```bash
cd web
npm ci
npm run build
```

El artefacto queda en `web/dist/`.

### 3. Crear bucket y publicar

Crear un bucket S3 y habilitar Static Website Hosting:

- Index document: `index.html`.
- Error document: `index.html`.

Con AWS CLI:

```bash
aws s3 sync web/dist/ s3://NOMBRE_BUCKET --delete
```

Para acceso publico mediante website hosting se debe desactivar el bloqueo
publico que corresponda y aplicar una policy de solo lectura:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicRead",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::NOMBRE_BUCKET/*"
    }
  ]
}
```

Como alternativa recomendada, mantener el bucket privado y usar CloudFront.

## CORS y red

La API permite durante desarrollo:

```text
Access-Control-Allow-Origin: *
GET, POST, PATCH, DELETE, OPTIONS
Content-Type
```

Para una publicacion real conviene restringir el origen al dominio S3 o
CloudFront y limitar el Security Group.

## Persistencia y actualizaciones

- Los mounts y la sesion son memoria del proceso; se pierden al reiniciar EC2.
- Los discos y reportes persisten en las carpetas configuradas.
- No almacenar discos dentro del repositorio.
- Respaldar `/home/ubuntu/mia/cali` antes de actualizar.

Actualizar backend:

```bash
git pull
go test ./...
go build -o mia-p2-api ./cmd/server
sudo systemctl restart mia-p2-api
```

Actualizar frontend:

```bash
cd web
npm ci
npm run build
aws s3 sync dist/ s3://NOMBRE_BUCKET --delete
```

## Checklist AWS

- [ ] API responde `/api/health` desde Internet.
- [ ] Graphviz genera un reporte SVG.
- [ ] EC2 puede escribir discos y reportes.
- [ ] Security Group no expone SSH globalmente.
- [ ] `VITE_API_BASE_URL` apunta a la API publica.
- [ ] No existe bloqueo de mixed content HTTP/HTTPS.
- [ ] S3 entrega `index.html` y assets.
- [ ] Login, explorador y reportes funcionan desde el frontend publicado.
