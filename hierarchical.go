package quantize

import (
	"errors"
	"image"
	"image/color"
	"math"
	"sort"

	"gonum.org/v1/gonum/mat"
)

type colorNode struct {
	mean    mat.Dense
	cov     mat.Dense
	classid uint8
	count   uint64

	left  *colorNode
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
	for i := 0; i < count-1; i++ {
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

func convertColor(col color.Color) []float64 {
	r, g, b, _ := col.RGBA()

	return []float64{float64(r) / 65535.0, float64(g) / 65535.0, float64(b) / 65535.0}
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
			if classes[y*bounds.Max.X+x] != node.classid {
				continue
			}
			scaled := mat.NewDense(3, 1, convertColor(img.At(x, y)))

			mean.Add(mean, scaled)
			tmp.Mul(scaled, scaled.T())
			cov.Add(cov, tmp)
			pixcount += 1
		}
	}

	tmp.Mul(mean, mean.T())
	cov.Apply(func(i, j int, v float64) float64 {
		return v - tmp.At(j, i)/float64(pixcount)
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

LOOP:
	for len(queue) > 0 {
		node, queue = queue[len(queue)-1], queue[:len(queue)-1]

		if node.left != nil && node.right != nil {
			queue = append(queue, node.left)
			queue = append(queue, node.right)
			continue
		}

		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if math.IsNaN(node.cov.At(j, i)) {
					continue LOOP
				}
			}
		}

		if !eigen.Factorize(&node.cov, true, true) {
			return nil, errors.New("bad factorization")
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

	eig := mat.NewDense(1, 3, eigen.Vectors().RawRowView(0))
	cmpValue.Mul(eig, &node.mean)

	node.left = &colorNode{classid: newidleft}
	node.right = &colorNode{classid: newidright}

	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			pos := y*bounds.Max.X + x
			if classes[pos] != node.classid {
				continue
			}

			thisValue.Mul(eig, mat.NewDense(3, 1, convertColor(img.At(x, y))))

			if thisValue.At(0, 0) <= cmpValue.At(0, 0) {
				node.left.count++
				classes[pos] = newidleft
			} else {
				node.right.count++
				classes[pos] = newidright
			}
		}
	}

	return nil
}

func getDominantColors(root *colorNode) []color.RGBA {
	ret := []color.RGBA{}
	for _, leave := range getLeaves(root) {
		c := color.RGBA{
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
	sort.Sort(sort.Reverse(ByCount(ret)))
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

type ByCount []*colorNode

func (c ByCount) Len() int           { return len(c) }
func (c ByCount) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCount) Less(i, j int) bool { return c[i].count < c[j].count }
