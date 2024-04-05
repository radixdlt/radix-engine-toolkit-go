# Radix Engine Toolkit Go binding
Repository contains Go package with Radix Engine Toolkit project Go language binding using Uniffi library.

## How to use
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

Create `main.go` file with content:
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

From latest [release](https://github.com/radixdlt/radix-engine-toolkit-go/releases) download library for you OS. And put it your project directory or in system library directory (for linux use `/usr/lib`).

Build project specyfing library to use and run it (`main` executable file will be created):
```
CGO_LDFLAGS="-lradix_engine_toolkit_uniffi" go build
./main
```
If you put radix_engine_toolkit_uniffi library in your project directory specify also library search path:
```
CGO_LDFLAGS="-L<path to directory with library> -lradix_engine_toolkit_uniffi" go build
LD_LIBRARY_PATH="<path to directory with library>" ./main
```
After running our simple program you should see information about `radix-engine-toolkit` version: `RET version: x.y.z`
