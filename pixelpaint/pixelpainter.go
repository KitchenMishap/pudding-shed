package pixelpaint

// Compiler check that implements
var _ IPixelPainter = (*PixelPainter)(nil)

type PixelPainter struct {
	pixels [][]IPixelAccumulator
}

func (pp PixelPainter) Reset(width int, height int, accumulatorFactory IPixelAccumulatorFactory) {

}
