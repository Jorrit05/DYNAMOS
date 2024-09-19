# Introduction

Most messages in DYNAMOS are definined in [Protocol Buffers](https://protobuf.dev/)

These generic message files need to be compiled into code of the language being targeted, currently Python and Go for DYNAMOS using the `protoc` compiler.

##  Install Protocol Buffers Compiler (protoc)

[Official Docs](https://grpc.io/docs/protoc-installation/)
[Go protoc-gen (plugin required for Go compilation)](https://pkg.go.dev/github.com/golang/protobuf/protoc-gen-go)

## Manual/detailed installation

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
7. Test protoc on the DYNAMOS proto files, for Python and Go.
```sh 
# Go
cd go
make proto

#Python
cd python
make proto
```

There should be no terminal outputs, but the contents within `./go/pkg/proto` might have been updated.
