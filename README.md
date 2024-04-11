# Radix Engine Toolkit Go binding
Repository contains the Go package with [Radix Engine Toolkit](https://github.com/radixdlt/radix-engine-toolkit) project Go language binding using Uniffi library.

## How to use
### Prerequisites
Installed Go, minimum version 1.21.

### Demo project
Create new Go project:
```
mkdir project
cd project
go mod init main
```

Add and install `radix_engine_toolkit_go` package dependency:
```
go get github.com/radixdlt/radix-engine-toolkit-go@latest
```

Create `main.go` file with following code:
```
package main

import (
    "github.com/radixdlt/radix-engine-toolkit-go/radix_engine_toolkit_uniffi"
    "fmt"
)

func main() {
    var buildInfo = radix_engine_toolkit_uniffi.GetBuildInformation()
    fmt.Println("RET version:", buildInfo.Version)
}
```

From the latest [release](https://github.com/radixdlt/radix-engine-toolkit-go/releases) download library for your OS.

### Linux instructions
Unpack the library and put it in `/usr/lib` or project directory.

Build project specifying library to use:
```
CGO_LDFLAGS="-lradix_engine_toolkit_uniffi" go build
```
if library is in project directory use command:
```
CGO_LDFLAGS="-L<path to directory with library> -lradix_engine_toolkit_uniffi" go build
```
Run program:
```
./main
```
if library is in project directory use command:
```
LD_LIBRARY_PATH="<path to directory with library>" ./main
```

### MacOS instructions
Unpack the library and put it in the project directory.

Build project specifying library to use and path to it:
```
CGO_LDFLAGS="-L<path to directory with library> -lradix_engine_toolkit_uniffi" go build
```
Run application, if library is in the same directory as executable file:
```
./main
```
If library is in other directory:
```
DYLD_LIBRARY_PATH="<path to directory with library>" ./main
```
 \
After running our simple program you should see information about `radix-engine-toolkit` version: `RET version: x.y.z`
