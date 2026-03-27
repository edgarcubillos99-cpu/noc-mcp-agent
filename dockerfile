# Etapa 1: Construcción (Builder)
FROM golang:1.22-alpine AS builder

# Instalar libcap para manipulación de capacidades en la etapa de build
RUN apk add --no-cache libcap

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Compilar el binario optimizado para Linux
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /noc-mcp-agent ./cmd/mcp

# Etapa 2: Ejecución (Producción)
FROM alpine:latest

# Instalar SOLO las herramientas de red necesarias
RUN apk add --no-cache \
    iputils \
    traceroute \
    nmap \
    libcap \
    tzdata

# Crear un usuario sin privilegios para mitigar vulnerabilidades de escape
RUN addgroup -S nocgroup && adduser -S nocuser -G nocgroup

# Otorgar capacidades de Raw Sockets a los binarios para que el usuario no-root pueda usarlos
# Esto es vital para entornos de telecomunicaciones donde el Zero Trust es ley.
RUN setcap cap_net_raw=ep /bin/ping && \
    setcap cap_net_raw=ep /usr/bin/nmap || true

WORKDIR /home/nocuser
COPY --from=builder /noc-mcp-agent ./

# Restringir permisos de ejecución al usuario
RUN chown nocuser:nocgroup /home/nocuser/noc-mcp-agent && \
    chmod 500 /home/nocuser/noc-mcp-agent

# Abandonar privilegios de root
USER nocuser

# Iniciar el agente MCP
ENTRYPOINT ["./noc-mcp-agent"]