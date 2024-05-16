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
go get github.com/radixdlt/radix-engine-toolkit-go/v2@latest
```

Create `main.go` file with following code:
```
package main

import (
    "github.com/radixdlt/radix-engine-toolkit-go/v2/radix_engine_toolkit_uniffi"
    "fmt"
)

func main() {
    var buildInfo = radix_engine_toolkit_uniffi.GetBuildInformation()
    fmt.Println("RET version:", buildInfo.Version)
}
```

From the latest [release](https://github.com/radixdlt/radix-engine-toolkit-go/releases) download library for your OS. Unpack the library and put it in project directory.

### Linux instructions

Build project specifying library to use:
```
CGO_LDFLAGS="-L<path to directory with library file> -lradix_engine_toolkit_uniffi" go build
```
Run program:
```
LD_LIBRARY_PATH="<path to directory with library file>" ./main
```

### MacOS instructions

> **_NOTE:_**  MacOS support is experimental.

Build project specifying library to use and path to it:
```
CGO_LDFLAGS="-L<path to directory with library file> -lradix_engine_toolkit_uniffi" go build
```
Run program:
```
DYLD_LIBRARY_PATH="<path to directory with library file>" ./main
```
 \
After running our simple program you should see information about `radix-engine-toolkit` version: `RET version: x.y.z`

## License

The Radix Engine Toolkit and Radix Engine Toolkit wrappers binaries are licensed under the [Radix Generic EULA](https://www.radixdlt.com/terms/genericEULA).

The Radix Engine Toolkit and Radix Engine Toolkit wrappers code is released under the [Apache 2.0 license](./LICENSE).


      Copyright 2023 Radix Publishing Ltd

      Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.

      You may obtain a copy of the License at: http://www.apache.org/licenses/LICENSE-2.0

      Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.

      See the License for the specific language governing permissions and limitations under the License.
