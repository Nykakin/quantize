package quantize

import (
    "errors"
    "image"
    "image/color"

    "gonum.org/v1/gonum/mat"
)

type colorNode struct {
    mean mat.Dense
    cov mat.Dense 
    classid uint8

    left *colorNode
    right *colorNode
}

func DominantColors(img image.Image, count int) ([]color.RGBA, error) {
    bounds := img.Bounds()
    pixelCount := bounds.Max.X * bounds.Max.Y

    classes := make([]uint8, pixelCount, pixelCount)
    for i := range classes {
        classes[i] = 1
    }
    root := &colorNode{classid: 1}

    getClassMeanCov(img, classes, root)
    for i := 0; i < count - 1; i++ {
        next, err := getMaxEigenvalueNode(root)
        if err != nil {
            return nil, err
        }
        err = partitionClass(img, classes, getNextClassid(root), next)
        if err != nil {
            return nil, err
        }
        getClassMeanCov(img, classes, next.left)
        getClassMeanCov(img, classes, next.right)
    }

    return getDominantColors(root), nil
}

func getClassMeanCov(img image.Image, classes []uint8, node *colorNode) {
    bounds := img.Bounds()
    tmp := mat.NewDense(3, 3, nil)

    mean := mat.NewDense(3, 1, []float64{0, 0, 0})
    cov := mat.NewDense(3, 3, []float64{
        0, 0, 0,
        0, 0, 0,
        0, 0, 0,
    })
    pixcount := 0
    for y := 0; y < bounds.Max.Y; y++ {
        for x := 0; x < bounds.Max.X; x++ {
            if classes[y * bounds.Max.X + x] != node.classid {
                continue
            }
            r, g, b, _ := img.At(x, y).RGBA()
            scaled := mat.NewDense(3, 1, []float64{
                float64(r) / float64(255.0),
                float64(g) / float64(255.0),
                float64(b) / float64(255.0),
            })

            mean.Add(mean, scaled)
            tmp.Mul(scaled, scaled.T())
            cov.Add(cov, tmp)
            pixcount += 1
        }
    }

    tmp.Mul(mean, mean.T())
    cov.Apply(func(i, j int, v float64) float64 {
        return v / float64(pixcount)
    }, cov)
    mean.Apply(func(i, j int, v float64) float64 {
        return v / float64(pixcount)
    }, mean)

    node.mean.Clone(mean)
    node.cov.Clone(cov)
}

func getMaxEigenvalueNode(current *colorNode) (*colorNode, error) {
    var eigen mat.Eigen

    maxEigen := float64(-1)
    queue := []*colorNode{current}
    var node *colorNode
    ret := current
    if current.left == nil && current.right == nil {
        return current, nil
    }
    for len(queue) > 0 {
        node, queue = queue[len(queue)-1], queue[:len(queue)-1]

        if node.left != nil && node.right != nil {
            queue = append(queue, node.left)
            queue = append(queue, node.right)
            continue
        }

        if !eigen.Factorize(&node.cov, true, true) {
            return nil, errors.New("bad factorization")
        }

        eigenvalues := eigen.Values(nil)
        maxValue := real(eigenvalues[0])
        for _, i := range eigenvalues[1:] {
            if real(i) > maxValue {
                maxValue = real(i)
            }
        }

        val := real(eigen.Values(nil)[0])
        if val > maxEigen {
            maxEigen = val
            ret = node
        }
    }
    return ret, nil
}

func partitionClass(img image.Image, classes []uint8, nextid uint8, node *colorNode) error {
    var eigen mat.Eigen
    var cmpValue mat.Dense
    var thisValue mat.Dense

    bounds := img.Bounds()
    newidleft := nextid
    newidright := nextid + 1

    if !eigen.Factorize(&node.cov, true, true) {
        return errors.New("bad factorization")
    }

    eigenvalues := eigen.Values(nil)
    maxEigenvalue := real(eigenvalues[0])
    maxEigenvalueIndex := 0
    if real(eigenvalues[1]) > maxEigenvalue {
        maxEigenvalue = real(eigenvalues[1])
        maxEigenvalueIndex = 1
    }
    if real(eigenvalues[2]) > maxEigenvalue {
        maxEigenvalueIndex = 2
    }

    eig := mat.NewDense(1, 3, eigen.Vectors().RawRowView(maxEigenvalueIndex))
    cmpValue.Mul(eig, &node.mean)

    node.left = &colorNode{classid: newidleft}
    node.right = &colorNode{classid: newidright}

    for y := 0; y < bounds.Max.Y; y++ {
        for x := 0; x < bounds.Max.X; x++ {
            pos := y * bounds.Max.X + x
            if classes[pos] != node.classid {
                continue
            }

            r, g, b, _ := img.At(x, y).RGBA()
            thisValue.Mul(
                eig,
                mat.NewDense(3, 1, []float64{
                    float64(r) / float64(255.0),
                    float64(g) / float64(255.0),
                    float64(b) / float64(255.0),
                }),
            )

            if thisValue.At(0, 0) <= cmpValue.At(0, 0) {
                classes[pos] = newidleft
            } else {
                classes[pos] = newidright
            }
        }
    }

    return nil
}

func getDominantColors(root *colorNode) []color.RGBA {
    ret := []color.RGBA{}
    for _, leave := range getLeaves(root) {
        c := color.RGBA {
            uint8(leave.mean.At(0, 0) * float64(255.0)),
            uint8(leave.mean.At(1, 0) * float64(255.0)),
            uint8(leave.mean.At(2, 0) * float64(255.0)),
            255,
        }
        ret = append(ret, c)
    }
    return ret
}

func getLeaves(root *colorNode) []*colorNode {
    ret := []*colorNode{}
    queue := []*colorNode{root}
    var current *colorNode
    for len(queue) > 0 {
        current, queue = queue[len(queue)-1], queue[:len(queue)-1]
        if current.left != nil && current.right != nil {
            queue = append(queue, current.left)
            queue = append(queue, current.right)
            continue
        }
        ret = append(ret, current)
    }
    return ret
}

func getNextClassid(root *colorNode) uint8 {
    maxid := uint8(0)

    queue := []*colorNode{root}
    var current *colorNode
    for len(queue) > 0 {
        current, queue = queue[len(queue)-1], queue[:len(queue)-1]

        if current.classid > maxid {
            maxid = current.classid
        }
        if current.left != nil {
            queue = append(queue, current.left)
        }
        if current.right != nil {
            queue = append(queue, current.right)
        }
    }

    return maxid + 1
}