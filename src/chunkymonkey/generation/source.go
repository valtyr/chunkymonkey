package generation


type Source interface {
	At2d(x, y float64) float64
	MeanMagnitude() float64
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


type Mult struct {
	A, B Source
}

func (gen *Mult) At2d(x, y float64) float64 {
	return gen.A.At2d(x, y) * gen.B.At2d(x, y)
}

func (gen *Mult) MeanMagnitude() float64 {
	return gen.A.MeanMagnitude() * gen.B.MeanMagnitude()
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
