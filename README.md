## go-storage

### 1. Global Environment

```
DATA_SERVER_PORT   storage server listen port default:4000  [data server only]
API_SERVER_PORT    api server listen port default:4000      [api server only]
STORAGE_ROOT       storage folder for save files            [data server only]
STORAGE_MQ_SERVER  rabbitmq server
STORAGE_ES_SERVER  elasticsearch server for store metadata
```

### 2. Development Environment Setup

```
docker run -d --hostname storage-rabbit -p 5672:5672 --name storage-rabbit rabbitmq:3
docker run -d --name storage-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch

DATA_SERVER_PORT="localhost:4000" STORAGE_ROOT="/tmp" STORAGE_MQ_SERVER="amqp://guest:guest@localhost:5672" go run main.go dataServer

API_SERVER_PORT="localhost:4001" STORAGE_ES_SERVER="localhost:9200" STORAGE_MQ_SERVER="amqp://guest:guest@localhost:5672" go run main.go apiServer
```

### 3. Test Method

#### upload file 

```
curl -v localhost:4001/objects/test -XPUT -d "hello world" -H "Digest: SHA-256=uU0nuZNNPgilLlLX2n2r+sSE7+N6U4DukIj3rOLvzek="
```

#### locate file 

```
curl localhost:4001/locate/Q02XU87fSrjhGuJ26n5AqptzEYmWMGmhNk30VHea6Gk=
```

#### show file version

```
curl -v localhost:4001/versions/test
```
