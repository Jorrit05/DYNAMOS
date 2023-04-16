docker stack deploy -c stack/logging.yaml -c stack/rabbitmq.yaml -c stack/etcd.yaml  core

docker service ps --no-trunc <ID>

docker network create --driver overlay core_network
docker network create --driver overlay unl_1
docker network create --driver overlay unl_2
docker network create --driver overlay third_party

openssl rand -base64 12 | docker secret create db_root_password -
openssl rand -base64 12 | docker secret create db_dba_password -
openssl rand -base64 12 | docker secret create rabbitmq_user -
(Get hashed pw by logging into rabbit container and "rabbitmqctl hash_password <PW>", I think there was another way through the api/definitions. But forgot..
Perhaps starting the service, creating the user manually, copying the hash from the api/definitions.. big brain time)

docker exec -it $(docker ps -f name=apps_db -q) mysql -u root -p
docker exec -it $(docker ps -f name=apps_db -q) mongo -u root -p example
docker exec -it $(docker ps -f name=service_service -q) /bin/sh

docker exec -it $(docker ps -f name=apps_db -q) cat /run/secrets/db_root_password

docker exec -it $(docker ps -f name=mongo -q) cat /run/secrets/db_root_password
docker exec -it $(docker ps -f name=apps_randomize_service -q) cat /run/secrets/rabbitmq_user

docker logs --since 5s $(docker ps -q --filter "ancestor=grafana/loki:2.8.0" --filter "status=restarting")



docker run --rm -d --name mongo \
        -e MONGO_INITDB_ROOT_USERNAME=root \
        -e MONGO_INITDB_ROOT_PASSWORD=example \
        -p 27017:27017 \
        mongo:4

{
    "query" : "SELECT `first_name`, `last_name`, `sex`, `person_id` FROM `person` LIMIT 2"
}

# Logs

docker volume create --name=service_logs


# MONGO

db.auth("root", passwordPrompt() )

# GoLang

go mod init
go get github.com/Jorrit05/GoLib@7f4fdc0293d3af27b39f3a7f811322bcd3e6b9dc


# ETCD

- Leader:
etcdctl --endpoints=http://etcd1:2379,http://etcd2:2379,http://etcd3:2379 endpoint status --write-out=table

- all container IP:
docker container ls --filter "name=etcd_cluster" --format "{{.ID}}" | xargs -n1 docker container inspect --format "{{.Name}} {{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}"

