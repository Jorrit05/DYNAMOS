# Development Dependencies Guide
The purpose of this document is to provide a guide of the *technical* prerequisites  to develop in DYNAMOS. Thus, the operating system commands and dependencies are covered, for more detailed guides about developing services, new commands etc, please read the "Development Process Guide" file in this directory

Dependencies that are required on your machine before developing services within DYNAMOS are highlighted below. Necessary configuration will also be covered below.

> [!NOTE]
> The dependencies you'll need depends on what type of service you develop, we currently have GO and Python service, thus these dependencies will be covered.

> [!NOTE]
> Most of the documentation relates to Linux operating system, namely Debian based. For alternative operating systems we do not guarentee that below will work, but links are provided from official sources to help you get it to work with your local setup. 

## Go
Most of the service in DYNAMOS are built with GO, we do not currently have a fixed version that is used, generally we try keep with the latest version. This may change in the future 

### Install GO

The first step is to download the GO binary from [here](https://go.dev/dl/). Download the respective binary for your OS.

Now that you have the binary, follow the guide (here)[https://go.dev/doc/install]. 

For Linux distros, you should first delete any previous version of GO and extract the archive into your `/usr/local/` directory.

Like so:
```sh 
 ➜ sudo rm -rf /usr/local/go &&
   cd ~/Downloads/ &&
   sudo tar -C /usr/local -xzf go*.linux-amd64.tar.gz
```

Make sure to add the path to your go installation to your PATH, this can be done like so:

```sh 
export PATH=$PATH:/usr/local/go/bin
```

You can now check your go version to validate your installation:
```sh 
➜ go version
go version go1.23.0 linux/amd64
```
We have validated that we installed GO version 1.23.0.

###  Install Protocol Buffers Compiler (protoc)
Protobufs play a big role in DYNAMOS, the `protoc` command generates GO code from `.proto` files.   

First, download the protoc release from [here](https://github.com/protocolbuffers/protobuf/releases).
Alternatively, use the following commands:

0. Make sure to update your system and install an unzipping tool (optional)
```sh 
sudo apt update &&
sudo apt install -y unzip
```
1. Set the latest `PROTOC_VERSION` as a variable
```sh 
PROTOC_VERSION=$(curl -s "https://api.github.com/repos/protocolbuffers/protobuf/releases/latest" | grep -Po '"tag_name": "v\K[0-9.]+')
```
In this example, `$PROTOC_VERSION = 28.0`

2. Download the ZIP from github
```sh 
wget -qO protoc.zip https://github.com/protocolbuffers/protobuf/releases/latest/download/protoc-$PROTOC_VERSION-linux-x86_64.zip
```
3. Unzip the release into your `/usr/local`
```sh 
sudo unzip -q protoc.zip bin/protoc -d /usr/local
```
4. Make the bin executable
```sh 
sudo chmod a+x /usr/local/bin/protoc
```
5. Validate the installation
```sh 
protoc --version # libprotoc 28.0
```
6. Delete the protoc zip
```sh 
rm -rf protoc.zip
```
7. Test protoc on the DYNAMOS proto files, from the project root dir, run the following:
```sh 
protoc -I ./proto-files --go_out=./go/pkg/proto --go_opt=paths=source_relative --go-grpc_out=./go/pkg/proto --go-grpc_opt=paths=source_relative ./proto-files/*.proto
```

There should be no terminal outputs, but the contents within `./go/pkg/proto` might have been updated.

## Python
Some DYNAMOS services are based on python. For the microservices based in Python, we have created a `dynamos` python pip library to ease integration. To use this library we recommend taking the following approach:

Firstly, you need to have python installed. Most modern Linux distros have Python3 pre-installed, you can check this by running the following: 

```sh 
python3 --version # Python 3.10.12
```
In the rare case that you do not have Python installed, use the following:
```sh 
sudo apt update
sudo apt install python3
```

### venv
To handle dependencies, we recommend to use venv (you are free to use other dependency managers such as anaconda)

If you're using Python3.4+, `venv` is available directly in python. Else, you can install it with pip, with the following command:
```sh 
# (Optional, since most distros have venv)
pip install virtualenv
```
Create a venv, in this case we will call the directory venv (second argument), but it can be anything you like:
```sh 
python -m venv venv
```
When developing, you can activate your venv with the following command:
```sh 
source venv/bin/activate
```
Keep in mind that the `venv` directory is wherever you created in the previous command.

### Prerequisite pip package 
To be able to use the python Makefile one dependency is required, all other dependencies are handled as requirements in the services themselves.

Activate your venv and run the following:
```sh 
pip install wheel
```

### DYNAMOS Pip package
As previously mentioned, we developed a python library that handles the initialisation and configuration of microservices, to build this locally you must:

1. Change directory to `dynamos-python-lib`
```sh 
cd python/dynamos-python-lib
```
2. Activate venv
```sh 
source venv/bin/activate
```
3. Run pip install:
```sh 
pip install .
```

The output of above command should look something like this:
```
...
Successfully built dynamos
Successfully installed dynamos-0.1
```

## Makefiles
Now that you have all your dependencies in check, we can locally build all images in DYNAMOS.
There are currently three Makefiles in this repo. One in the root directory, on in the `go/` directory and one in the `python/` directory.

### RabbitMQ Makefile (root)

TODO: Make sure this works, and that what im saying is right (@Jorrit?)
The root directory Makefile only has one task: `create-rabbitmq-secret`, this can be used to replace the RabbitMQ token 

This task can be executed with the following:
```sh 
make create-rabbitmq-secret
```

### Building Makefiles 
The GO and Python Makefiles are used to:
A) Prepare the GO/Python packages
B) Compile proto files
C) Locally build Docker containers
D) Push containers to registry (dockerhub)

The newly built docker images will have the following naming convention:
`<dockerhub_account_name>/<service_name>:<branch_name>`

If you want to change the dockerhub account where the images are pushed, edit the `dockerhub_account` variable in the Makefiles.

> [!NOTE]
> If you want to push the docker images to a registry, you must first `docker login` and be associated with a registry.

#### GO Makefile
The GO Makefile supports the following tasks:

1. dynamos
```sh 
make dynamos
```
This task builds the following services:
- Sidecar
- Policy Enforcer
- Orchestrator
- Agent
- API Gateway

2. sql_microservices
```sh 
make sql_microservices
``` 

This task builds the following services:
- SQL algorithm
- SQL Anonymize
- SQL Aggregate
- SQL Test

3. all
```sh
make all
```

This task builds all of the above services in one go.

#### Python Makefile
Currently the python Makefile only supports the `all` task. Change directory to the `python/`
```sh
 
make all
```
This will build the python `sql-query` microservice.


That concludes the development dependency guide! You should be setup to start developing.
If you want more information regarding the development process, for example how to create new 
microservices and commands, please read through the Microservice Development Guide! :)
