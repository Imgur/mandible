# Aspell library bindings for Go

GNU Aspell is a spell checking tool written in C/C++. This package provides simplified Aspell bindings for Go.
It uses UTF-8 by default and encapsulates some Aspell internals.

## Getting started

First make sure aspell library and headers are installed on your system. On Debian/Ubuntu you could install it this way:

```
sudo apt-get install aspell libaspell-dev
```

It you need some more dictionaries you can install them like this:

```
sudo apt-get install aspell-ua aspell-se
```

Then you can install the package using the Go tool:

```
go get github.com/trustmaster/go-aspell
```

## Usage

Here is a simple spell checker program using the aspell package:

```go
package main

import (
	"github.com/trustmaster/go-aspell"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Get a word from cmd line arguments
	if len(os.Args) != 2 {
		fmt.Print("Usage: aspell_example word\n")
		return
	}
	word := os.Args[1]

	// Initialize the speller
	speller, err := aspell.NewSpeller(map[string]string{
		"lang": "en_US",
	})
	if err != nil {
		fmt.Errorf("Error: %s", err.Error())
		return
	}
	defer speller.Delete()

	// Check and suggest
	if speller.Check(word) {
		fmt.Print("OK\n")
	} else {
		fmt.Printf("Incorrect word, suggestions: %s\n", strings.Join(speller.Suggest(word), ", "))
	}
}
```

For more information see [aspell_test.go](https://github.com/trustmaster/go-aspell/blob/master/aspell_test.go) file and use the godoc tool:

```
godoc github.com/trustmaster/go-aspell
```

## License

Copyright (c) 2012, Vladimir Sibirov
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
