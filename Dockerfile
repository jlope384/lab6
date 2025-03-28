FROM golang:1.24.1

WORKDIR /app

# Copiar solo los archivos de módulo primero
COPY go.mod go.sum ./

# Descargar dependencias (se cachea este paso si no cambian go.mod/go.sum)
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Sincronizar dependencias y compilar
RUN go mod tidy && go build -o laLigaAPI

EXPOSE 8080
CMD ["./laLigaAPI"]