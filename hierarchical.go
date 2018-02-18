package quantize

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/Nykakin/eigenvalues"
)

type colorNode struct {
	mean    vec3x1
	cov     mat3x3
	classid uint8
	count   uint64

	left  *colorNode
	right *colorNode
}

func newColorNode(classid uint8) *colorNode {
	return &colorNode{
		classid: classid,
		mean:    newVec3x1(),
		cov:     newMat3x3(),
	}
}

type Quantizer interface {
	Quantize(img image.Image, count int) ([]color.RGBA, error)
}

type hierarhicalQuantizer struct {
	tmp3x3 mat3x3
	tmp3x1 vec3x1
	tmp1x3 vec1x3
}

func NewHierarhicalQuantizer() hierarhicalQuantizer {
	return hierarhicalQuantizer{
		tmp3x3: newMat3x3(),
		tmp3x1: newVec3x1(),
		tmp1x3: newVec1x3(),
	}
}

func (hq hierarhicalQuantizer) Quantize(img image.Image, count int) ([]color.RGBA, error) {
	bounds := img.Bounds()
	pixelCount := bounds.Max.X * bounds.Max.Y

	classes := make([]uint8, pixelCount, pixelCount)
	for i := range classes {
		classes[i] = 1
	}
	root := newColorNode(1)

	hq.getClassMeanCov(img, classes, root)
	for i := 0; i < count-1; i++ {
		next, err := hq.getMaxEigenvalueNode(root)
		if err != nil {
			return nil, err
		}
		err = hq.partitionClass(img, classes, getNextClassid(root), next)
		if err != nil {
			return nil, err
		}
		hq.getClassMeanCov(img, classes, next.left)
		hq.getClassMeanCov(img, classes, next.right)
	}
	return getDominantColors(root), nil
}

func convertColor(col color.Color) (color []float64, isTransparent bool) {
	r, g, b, a := col.RGBA()
	// TODO: handle transparency more smartly
	if a == 0 {
		return nil, true
	}

	return []float64{float64(r) / 65535.0, float64(g) / 65535.0, float64(b) / 65535.0}, false
}

func (hq hierarhicalQuantizer) getClassMeanCov(img image.Image, classes []uint8, node *colorNode) {
	bounds := img.Bounds()

	node.mean.set(0)
	node.cov.set(0)
	pixcount := 0

	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			if classes[y*bounds.Max.X+x] != node.classid {
				continue
			}

			color, isTransparent := convertColor(img.At(x, y))
			if isTransparent {
				continue
			}
			hq.tmp3x1.setVec(color)
			node.mean.add(node.mean, hq.tmp3x1)
			hq.tmp1x3.t(hq.tmp3x1)
			hq.tmp3x3.mul(hq.tmp3x1, hq.tmp1x3)
			node.cov.add(node.cov, hq.tmp3x3)
			pixcount += 1
		}
	}

	hq.tmp1x3.t(node.mean)
	hq.tmp3x3.mul(node.mean, hq.tmp1x3)
	node.cov.apply(func(i, j int, v float64) float64 {
		return v - hq.tmp3x3.at(j, i)/float64(pixcount)
	}, node.cov)
	node.mean.setVec([]float64{
		node.mean.atVec(0) / float64(pixcount),
		node.mean.atVec(1) / float64(pixcount),
		node.mean.atVec(2) / float64(pixcount),
	})
}

func (hq hierarhicalQuantizer) getMaxEigenvalueNode(current *colorNode) (*colorNode, error) {
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
				if math.IsNaN(node.cov.at(j, i)) {
					continue LOOP
				}
			}
		}

		r := eigenvalues.NewEigenvalueDecomposition(node.cov)
		val := r.EigenvaluesReal()[0]

		if val > maxEigen {
			maxEigen = val
			ret = node
		}
	}
	return ret, nil
}

func (hq hierarhicalQuantizer) partitionClass(img image.Image, classes []uint8, nextid uint8, node *colorNode) error {
	cmpValue := newMat3x3()
	thisValue := newMat3x3()

	bounds := img.Bounds()
	newidleft := nextid
	newidright := nextid + 1

	r := eigenvalues.NewEigenvalueDecomposition(node.cov)
	eig := newVec3x1()
	eig.setVec(r.Eigenvector()[0])
	cmpValue.mul(eig, node.mean)

	node.left = newColorNode(newidleft)
	node.right = newColorNode(newidright)

	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			pos := y*bounds.Max.X + x
			if classes[pos] != node.classid {
				continue
			}

			color, isTransparent := convertColor(img.At(x, y))
			if isTransparent {
				continue
			}
			hq.tmp3x1.setVec(color)
			thisValue.mul(eig, hq.tmp3x1)

			if thisValue.at(0, 0) <= cmpValue.at(0, 0) {
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
			uint8(leave.mean.atVec(0) * float64(255.0)),
			uint8(leave.mean.atVec(1) * float64(255.0)),
			uint8(leave.mean.atVec(2) * float64(255.0)),
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
