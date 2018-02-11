# quantize

> This package is not yet production ready

This package reimplements hierarchical quantization described in [this tutorial](http://aishack.in/tutorials/dominant-color/) in Go programming language. [Gonum](https://github.com/gonum/gonum) library was used for matrix computations instead of OpenCV.

```go
package main

import (
    "fmt"

    "image"
    _ "image/jpeg"
    "os"

    "github.com/Nykakin/quantize"
)   

func main() {
    f, err := os.Open("test.jpg")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    img, _, err := image.Decode(f)
    if err != nil {
        panic(err)
    }

    colors, err := quantize.DominantColors(img, 5)
    if err != nil {
        panic(err)
    }    
    for _, c := range colors {
        fmt.Printf("#%.2X%.2X%.2X\n", c.R, c.G, c.B)
    }
}
```