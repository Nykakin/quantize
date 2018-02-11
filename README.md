# quantize

> This package is not yet production ready

This package reimplements hierarchical quantization described in [this tutorial](http://aishack.in/tutorials/dominant-color/) in Go programming language. [Gonum](https://github.com/gonum/gonum) library was used for matrix computations instead of OpenCV.

```go
package main

import (
    "image"
    "image/color"
    _ "image/jpeg"
    _ "image/png"
    "os"

    "github.com/Nykakin/quantize"
    "github.com/joshdk/preview"
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

    palette := make([]color.Color, len(colors))
    for index, clr := range colors {
    	palette[index] = clr
    }

    // Display our new palette
    preview.Show(palette)
}
```
