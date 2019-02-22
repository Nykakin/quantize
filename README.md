# quantize

This package reimplements hierarchical quantization described in [this tutorial](http://aishack.in/tutorials/dominant-color/) in Go programming language. [Gonum](https://github.com/gonum/gonum) was used instead of OpenCV, then latter replaced with simpler in-home matrix types and [eigenvalue decomposition algorithm](https://github.com/Nykakin/eigenvalues) adapred from Java [JAML](https://math.nist.gov/javanumerics/jama/) package. This allowed to reduce ammount of needed dependencies. The effect and comparission with different Go packages can be found in [this repository](https://github.com/Nykakin/QuantizationTournament). Described eigenvalue method, while correct, in practice appears to be much slower than alternative methods, mostly based on some sort of k-means clustering. Therefore it doesn't really seem to be a good choice for a production code. It does a better job with detecting dominant background colors than some other competitors, though.

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

    quantizer := quantize.NewHierarhicalQuantizer()
    colors, err := quantize.Quantize(img, 5)
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
