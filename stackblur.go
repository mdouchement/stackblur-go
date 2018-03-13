// Go implementation of StackBlur algorithm described here:
// http://incubator.quasimondo.com/processing/fast_blur_deluxe.php

package stackblur

import (
	"image"
	"image/color"
	"image/draw"
)

// blurStack is a linked list containing the color value and a pointer to the next struct.
type blurStack struct {
	r, g, b, a uint32
	next       *blurStack
}

// NewBlurStack is a constructor function returning a new struct of type blurStack.
func (bs *blurStack) NewBlurStack() *blurStack {
	return &blurStack{bs.r, bs.g, bs.b, bs.a, bs.next}
}

// Process takes an image as parameter and returns it's blurred version by applying the blur radius.
func Process(src image.Image, radius int, done chan struct{}) image.Image {
	var stackEnd, stackIn, stackOut *blurStack
	var width, height = src.Bounds().Dx(), src.Bounds().Dy()
	var (
		div, widthMinus1, heightMinus1, radiusPlus1, sumFactor, p int
		rSum, gSum, bSum, aSum,
		rOutSum, gOutSum, bOutSum, aOutSum,
		rInSum, gInSum, bInSum, aInSum,
		pr, pg, pb, pa uint32
	)

	// Copy image
	img := image.NewNRGBA(src.Bounds())
	draw.Draw(img, img.Bounds(), src, src.Bounds().Min, draw.Src)

	div = radius + radius + 1
	widthMinus1 = width - 1
	heightMinus1 = height - 1
	radiusPlus1 = radius + 1
	sumFactor = radiusPlus1 * (radiusPlus1 + 1) / 2
	divsum := uint32((div + 1) >> 1)
	divsum *= divsum

	bs := blurStack{}
	stackStart := bs.NewBlurStack()
	stack := stackStart

	for i := 1; i < div; i++ {
		stack.next = bs.NewBlurStack()
		stack = stack.next
		if i == radiusPlus1 {
			stackEnd = stack
		}
	}
	stack.next = stackStart

	for y := 0; y < height; y++ {
		rInSum, gInSum, bInSum, aInSum, rSum, gSum, bSum, aSum = 0, 0, 0, 0, 0, 0, 0, 0

		r, g, b, a := img.At(0, y).RGBA()
		pr = to8b(r)
		pg = to8b(g)
		pb = to8b(b)
		pa = to8b(a)

		rOutSum = uint32(radiusPlus1) * pr
		gOutSum = uint32(radiusPlus1) * pg
		bOutSum = uint32(radiusPlus1) * pb
		aOutSum = uint32(radiusPlus1) * pa

		rSum += uint32(sumFactor) * pr
		gSum += uint32(sumFactor) * pg
		bSum += uint32(sumFactor) * pb
		aSum += uint32(sumFactor) * pa

		stack = stackStart

		for i := 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa
			stack = stack.next
		}

		for i := 1; i < radiusPlus1; i++ {
			p = i
			if widthMinus1 < i {
				p = widthMinus1
			}

			r, g, b, a := img.At(p, y).RGBA()
			pr = to8b(r)
			pg = to8b(g)
			pb = to8b(b)
			pa = to8b(a)

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			rSum += stack.r * uint32(radiusPlus1-i)
			gSum += stack.g * uint32(radiusPlus1-i)
			bSum += stack.b * uint32(radiusPlus1-i)
			aSum += stack.a * uint32(radiusPlus1-i)

			rInSum += pr
			gInSum += pg
			bInSum += pb
			aInSum += pa

			stack = stack.next
		}
		stackIn = stackStart
		stackOut = stackEnd

		for x := 0; x < width; x++ {
			img.Set(x, y, color.NRGBA{
				R: uint8(rSum / divsum),
				G: uint8(gSum / divsum),
				B: uint8(bSum / divsum),
				A: uint8(aSum / divsum),
			})

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum
			aSum -= aOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b
			aOutSum -= stackIn.a

			p = x + radius + 1

			if p > widthMinus1 {
				p = widthMinus1
			}

			r, g, b, a := img.At(p, y).RGBA()
			stackIn.r = to8b(r)
			stackIn.g = to8b(g)
			stackIn.b = to8b(b)
			stackIn.a = to8b(a)

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b
			aInSum += stackIn.a

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum
			aSum += aInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb
			aOutSum += pa

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb
			aInSum -= pa

			stackOut = stackOut.next
		}
	}

	for x := 0; x < width; x++ {
		rInSum, gInSum, bInSum, aInSum, rSum, gSum, bSum, aSum = 0, 0, 0, 0, 0, 0, 0, 0

		r, g, b, a := img.At(x, 0).RGBA()
		pr = to8b(r)
		pg = to8b(g)
		pb = to8b(b)
		pa = to8b(a)

		rOutSum = uint32(radiusPlus1) * pr
		gOutSum = uint32(radiusPlus1) * pg
		bOutSum = uint32(radiusPlus1) * pb
		aOutSum = uint32(radiusPlus1) * pa

		rSum += uint32(sumFactor) * pr
		gSum += uint32(sumFactor) * pg
		bSum += uint32(sumFactor) * pb
		aSum += uint32(sumFactor) * pa

		stack = stackStart

		for i := 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa
			stack = stack.next
		}

		for i := 1; i <= radius; i++ {
			r, g, b, a := img.At(x, i).RGBA()
			pr = to8b(r)
			pg = to8b(g)
			pb = to8b(b)
			pa = to8b(a)

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			rSum += stack.r * uint32(radiusPlus1-i)
			gSum += stack.g * uint32(radiusPlus1-i)
			bSum += stack.b * uint32(radiusPlus1-i)
			aSum += stack.a * uint32(radiusPlus1-i)

			rInSum += pr
			gInSum += pg
			bInSum += pb
			aInSum += pa

			stack = stack.next
		}

		stackIn = stackStart
		stackOut = stackEnd

		for y := 0; y < height; y++ {
			img.Set(x, y, color.NRGBA{
				R: uint8(rSum / divsum),
				G: uint8(gSum / divsum),
				B: uint8(bSum / divsum),
				A: uint8(aSum / divsum),
			})

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum
			aSum -= aOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b
			aOutSum -= stackIn.a

			p = y + radiusPlus1

			if p > heightMinus1 {
				p = heightMinus1
			}
			r, g, b, a := img.At(x, p).RGBA()
			stackIn.r = to8b(r)
			stackIn.g = to8b(g)
			stackIn.b = to8b(b)
			stackIn.a = to8b(a)

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b
			aInSum += stackIn.a

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum
			aSum += aInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb
			aOutSum += pa

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb
			aInSum -= pa

			stackOut = stackOut.next
		}
	}
	done <- struct{}{}
	return img
}

func to8b(c uint32) uint32 {
	return uint32(uint8(c)) // Go on 0..255 values
}
