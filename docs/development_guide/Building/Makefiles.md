
[Makefiles](https://opensource.com/article/18/8/what-how-makefile) are a useful tool in automating repeating tasks or building software. At the moment there is a Makefile in the `go/` directory for building all services in Go, and one in the `python/` directory for all Python services.

The newly built docker images will have the following naming convention:
`<dockerhub_account_name>/<service_name>:<branch_name>`

> [!NOTE]
> If you want to push the docker images to a registry, you must first `docker login` and be associated with a registry. At the moment a maximum of three people can be added to the `dynamos1` registry.

# GO Makefile

The folder name of a service need to match the Makefile target name:
```shell
make agent

# A folder with the name 'agent' has to be present
```

## Steps

1. Copy the `Dockerfile`,  `go.mod`, `go.sum`, `pkg` folder (Go Library) into the target service
2. Build docker image
3. Push to registry with tag of current GIT branch name
4. Remove all copied files

## Targets

The number of services built can be specified with `targets`. 

### Proto

Generate proto files.
```sh 
make proto
```

### Build a single service:
```sh 
make sql-algorithm
```

This will only work if this is defined as a target somewhere at the top of the file, in our case it is defined in the `microservices` target:

```Makefile
sql_microservices := sql-algorithm sql-anonymize sql-aggregate sql-test
```


### Build a collection of services

Create a target with a set of logically grouped services that you would like to build. At the moment these are available:

```Makefile
sql_microservices := sql-algorithm sql-anonymize sql-aggregate sql-test
dynamos := sidecar policy-enforcer orchestrator agent api-gateway
```

So:

```sh 
make dynamos
```

Will build all related services in order. 
### All

```sh
make all
```

This task builds all  targets.

# Python Makefile
Change directory to the `python/`

The folder name of a service need to match the Makefile target name:
```shell
make sql-query

# A folder with the name 'sql-query' has to be present
```

## Steps

1. Copy the `Dockerfile`, into the target service
2. Build the `dynamos`pip package and copy to target service
3. copy the generated python proto files to target service
4. Build docker image
5. Push to registry with tag of current GIT branch name
6. Remove all copied files

## Targets
See [[#GO Makefile]] for more details how targets work. 

Available targets:

```Makefile
targets := sql-query
```

### Proto

Generate proto files.
```sh 
make proto
```
