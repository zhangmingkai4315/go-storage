## go-storage

### 1. Global Environment

```
DATA_SERVER_PORT   storage server listen port default:4000  [data server only]
API_SERVER_PORT    api server listen port default:4000      [api server only]
STORAGE_ROOT       storage folder for save files            [data server only]
STORAGE_MQ_SERVER  rabbitmq server                          
```

### 2. Development Environment Setup

```
docker run -d --hostname storage-rabbit -p 5672:5672 --name storage-rabbit rabbitmq:3

DATA_SERVER_PORT="localhost:4000" STORAGE_ROOT="/tmp" STORAGE_MQ_SERVER="amqp://guest:guest@localhost:5672" go run main.go dataServer

API_SERVER_PORT="localhost:4001" STORAGE_MQ_SERVER="amqp://guest:guest@localhost:5672" go run main.go apiServer
```

### 3. Test Method

#### upload file 

```
curl -v localhost:4001/objects/test -XPUT -d "this is a test object"
```

#### locate file 

```
curl localhost:4001/locate/test
```

#### download file

```
curl -v localhost:4001/objects/test
```
