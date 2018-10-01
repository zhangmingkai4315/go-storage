## go-storage

### 1. Environment Setting 

```
STORAGE_PORT  storage server listen port default:4000
STORAGE_ROOT  storage folder for save files
STORAGE_MQ_SERVER  rabbitmq server
```

### 2. Development Environment

```
docker run -d --hostname storage-rabbit -p 5672:5672 --name storage-rabbit rabbitmq:3

STORAGE_PORT="localhost:4000" STORAGE_ROOT="/tmp" STORAGE_MQ_SERVER="amqp://guest:guest@localhost:5672 go run main.go"
```