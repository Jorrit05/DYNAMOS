# Introduction
Most of the service in DYNAMOS are built with GO, we do not currently have a fixed version that is used, generally we try keep with the latest version. This may change in the future 

### Install GO

[Official Installation Guide](https://go.dev/doc/install)
[Go Release page](https://go.dev/dl/)
## Manual/detailed installation

The first step is to download the GO binary from [here](https://go.dev/dl/). Download the respective binary for your OS.

Now that you have the binary, follow the guide [[https://go.dev/doc/install]]. 

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

