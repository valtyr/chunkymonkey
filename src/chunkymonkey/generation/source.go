package generation

import (
	"math"
)


type Source interface {
	At2d(x, y float64) float64
	MeanMagnitude() float64
}


type Const float64

func (gen Const) At2d(x, y float64) float64 {
	return float64(gen)
}

func (gen Const) MeanMagnitude() float64 {
	return float64(gen)
}


type Offset struct {
	Dx, Dy float64
	Source Source
}

func (gen *Offset) At2d(x, y float64) float64 {
	return gen.Source.At2d(x+gen.Dx, y+gen.Dy)
}

func (gen *Offset) MeanMagnitude() float64 {
	return gen.Source.MeanMagnitude()
}


type Turbulence struct {
	Dx, Dy Source
	Factor float64
	Source Source
}

func (gen *Turbulence) At2d(x, y float64) float64 {
	dx := gen.Dx.At2d(x, y) * gen.Factor
	dy := gen.Dy.At2d(x, y) * gen.Factor
	return gen.Source.At2d(x+dx, y+dy)
}

func (gen *Turbulence) MeanMagnitude() float64 {
	return gen.Source.MeanMagnitude()
}


type Scale struct {
	Wavelength float64
	Amplitude  float64
	Source     Source
}

func (gen *Scale) At2d(x, y float64) float64 {
	return gen.Source.At2d(x/gen.Wavelength, y/gen.Wavelength) * gen.Amplitude
}

func (gen *Scale) MeanMagnitude() float64 {
	return gen.Amplitude * gen.Source.MeanMagnitude()
}


type Pow struct {
	Source Source
	Power  Source
}

func (gen *Pow) At2d(x, y float64) float64 {
	v := gen.Source.At2d(x, y)
	power := gen.Power.At2d(x, y)
	if v < 0 {
		// Retain sign, and don't break on negatives raised to fractional powers.
		// Not strictly a power function, but serves the purpose.
		return -math.Pow(-v, power)
	}
	return math.Pow(v, power)
}

func (gen *Pow) MeanMagnitude() float64 {
	// TODO Remove all MeanMagnitude.
	return 0.5
}


type Mult struct {
	A, B Source
}

func (gen *Mult) At2d(x, y float64) float64 {
	return gen.A.At2d(x, y) * gen.B.At2d(x, y)
}

func (gen *Mult) MeanMagnitude() float64 {
	return gen.A.MeanMagnitude() * gen.B.MeanMagnitude()
}


type Add struct {
	Source Source
	Value  float64
}

func (gen *Add) At2d(x, y float64) float64 {
	return gen.Source.At2d(x, y) + gen.Value
}

func (gen *Add) MeanMagnitude() float64 {
	return gen.Source.MeanMagnitude() + gen.Value
}


type SumStack struct {
	Inputs []Source
}

func NewSumStack(inputs []Source) *SumStack {
	return &SumStack{
		Inputs: inputs,
	}
}

func (gen *SumStack) At2d(x, y float64) float64 {
	var accum float64
	for _, input := range gen.Inputs {
		accum += input.At2d(x, y)
	}
	return accum
}

func (gen *SumStack) MeanMagnitude() float64 {
	var meanAccum float64
	for _, input := range gen.Inputs {
		meanAccum += input.MeanMagnitude()
	}
	return meanAccum
}
