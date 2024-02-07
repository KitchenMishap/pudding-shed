package pixelpaint

// Compiler check that implements
var _ IPixelAccumulatorFactory = (*BasicAccumulator32Factory)(nil)

type BasicAccumulator32Factory struct {
}

func (baf *BasicAccumulator32Factory) NewAccumulator() *BasicAccumulator32 {
	result := BasicAccumulator32{}
	result.Reset()
	return &result
}

// Compiler check that implements
var _ IPixelAccumulator = (*BasicAccumulator32)(nil)

type BasicAccumulator32 struct {
	weight   float32
	red      float32
	green    float32
	blue     float32
	octarine float32
}

func (ba *BasicAccumulator32) Reset() {
	ba.weight = 0.0
	ba.red = 0.0
	ba.green = 0.0
	ba.blue = 0.0
	ba.octarine = 0.0
}

func (ba *BasicAccumulator32) RgboIn(weight float64, red byte, green byte, blue byte, octarine byte) {
	ba.red = ba.weight*ba.red + float32(weight)*float32(red)
	ba.green = ba.weight*ba.green + float32(weight)*float32(green)
	ba.blue = ba.weight*ba.blue + float32(weight)*float32(blue)
	ba.octarine = ba.weight*ba.octarine + float32(weight)*float32(octarine)
	ba.weight += float32(weight)
}

func (ba *BasicAccumulator32) WeightOut() float64 {
	return float64(ba.weight)
}

func (ba *BasicAccumulator32) RgboOut() (red byte, green byte, blue byte, octarine byte) {
	return byte(ba.red), byte(ba.green), byte(ba.blue), byte(ba.octarine)
}
