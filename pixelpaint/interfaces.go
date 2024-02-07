package pixelpaint

// Partial (ie weighted) contributions are made to a pixel's 32 bit colour.
// You can then read out the accumulated colour bytes.
// Note that different implementations may mix colours in subtley different ways.
type IPixelAccumulator interface {
	Reset()
	RgboIn(weight float64, red byte, green byte, blue byte, octarine byte)
	WeightOut() float64
	RgboOut() (red byte, green byte, blue byte, octarine byte)
}

type IPixelAccumulatorFactory interface {
	NewAccumulator() IPixelAccumulator
}

type IPixelPainter interface {
	Reset(width int, height int, accumulatorFactory IPixelAccumulatorFactory)
	// Takes co-ordinates which go between 0.0 and 1.0
	PaintRect(left float64, top float64, right float64, bottom float64, red byte, green byte, blue byte, octarine byte)
}
