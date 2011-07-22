package physics

import (
	"math"
	"testing"

	gomock "gomock.googlecode.com/hg/gomock"

	. "chunkymonkey/types"
)

func almostEqual(v1, v2 float64) bool {
	return math.Fabs(v1-v2) < 1e-10
}

func Test_calcNextBlockDt(t *testing.T) {
	type Test struct {
		p        AbsCoord
		v        AbsVelocityCoord
		expected TickTime
	}

	tests := []Test{
		// Degenerate case of zero velocity
		Test{0.0, 0.0, math.MaxFloat64},
		Test{0.5, 0.0, math.MaxFloat64},
		Test{-0.5, 0.0, math.MaxFloat64},

		// Not sure if the apparent disparity between these two cases matters,
		// given the starting position is on a block boundary.
		Test{0.0, 0.1, 10.0},
		Test{0.0, -0.1, 0.0},

		// +ve pos, +ve vel
		Test{1.0, 0.1, 10.0},
		Test{1.0, 0.5, 2.0},
		Test{1.0, 1.0, 1.0},
		Test{20.0, 0.1, 10.0},
		Test{20.0, 0.5, 2.0},
		Test{20.0, 1.0, 1.0},

		// +ve pos, -ve vel
		Test{0.9, -0.1, 9.0},
		Test{0.9, -0.5, 1.8},
		Test{0.9, -1.0, 0.9},
		Test{19.9, -0.1, 9.0},
		Test{19.9, -0.5, 1.8},
		Test{19.9, -1.0, 0.9},

		// -ve pos, -ve vel
		Test{-1.0, -0.1, 10.0},
		Test{-1.0, -0.5, 2.0},
		Test{-1.0, -1.0, 1.0},
		Test{-20.0, -0.1, 10.0},
		Test{-20.0, -0.5, 2.0},
		Test{-20.0, -1.0, 1.0},

		// -ve pos, +ve vel
		Test{-0.9, 0.1, 9.0},
		Test{-0.9, 0.5, 1.8},
		Test{-0.9, 1.0, 0.9},
		Test{-19.9, 0.1, 9.0},
		Test{-19.9, 0.5, 1.8},
		Test{-19.9, 1.0, 0.9},

		// Crossing p=0
		Test{-0.5, 1.0, 0.5},
		Test{0.5, -1.0, 0.5},
	}

	for _, r := range tests {
		result := calcNextBlockDt(r.p, r.v)
		if !almostEqual(float64(r.expected), float64(result)) {
			t.Errorf("calcNextBlockDt(%g, %g) expected %g got %g",
				r.p, r.v, r.expected, result)
		}
	}
}

func Test_getBlockAxisMove(t *testing.T) {
	type Test struct {
		xDt, yDt, zDt    TickTime
		expBlockAxisMove blockAxisMove
		expDt            TickTime
	}

	tests := []Test{
		// In these first tests it doesn't really matter which axis is the
		// answer, but the code has a bias for the X axis so we roll with that.
		Test{0.0, 0.0, 0.0, blockAxisMoveX, 0.0},
		Test{0.5, 0.5, 0.5, blockAxisMoveX, 0.5},
		Test{1.0, 1.0, 1.0, blockAxisMoveX, 1.0},
		Test{1.5, 1.5, 1.5, blockAxisMoveX, 1.5},

		// Hit block boundary on X axis first
		Test{0.5, 0.9, 0.9, blockAxisMoveX, 0.5},
		Test{0.5, 10.0, 20.0, blockAxisMoveX, 0.5},
		Test{0.1, 20.0, 10.0, blockAxisMoveX, 0.1},

		// Hit block boundary on Y axis first
		Test{0.9, 0.5, 0.9, blockAxisMoveY, 0.5},
		Test{10.0, 0.5, 20.0, blockAxisMoveY, 0.5},
		Test{20.0, 0.1, 10.0, blockAxisMoveY, 0.1},

		// Hit block boundary on Z axis first
		Test{0.9, 0.9, 0.5, blockAxisMoveZ, 0.5},
		Test{10.0, 20.0, 0.5, blockAxisMoveZ, 0.5},
		Test{20.0, 10.0, 0.1, blockAxisMoveZ, 0.1},
	}

	for _, r := range tests {
		resultBlockAxisMove, resultDt := getBlockAxisMove(r.xDt, r.yDt, r.zDt)
		if r.expBlockAxisMove != resultBlockAxisMove || !almostEqual(float64(r.expDt), float64(resultDt)) {
			t.Errorf(
				"getBlockAxisMove(%g, %g, %g) expected %d,%g got %d,%g",
				r.xDt, r.yDt, r.zDt,
				r.expBlockAxisMove, r.expDt,
				resultBlockAxisMove, resultDt)
		}
	}
}

func Test_VelocityFromLook(t *testing.T) {
	type Test struct {
		look     LookDegrees
		momentum float64
		want     AbsVelocity
	}
	tests := []Test{}

	// TODO: Add some unit tests that accurately reflect the desired
	// behaviour.
	for _, test := range tests {
		v := VelocityFromLook(test.look, test.momentum)
		if v.X != test.want.X || v.Y != test.want.Y || v.Z != test.want.Z {
			t.Errorf("VelocityFromLook, wanted %+v, got %+v", test.want, v)
		}
	}
}

func tickFixtures(t *testing.T) (mockCtrl *gomock.Controller, mockBlockQuerier *MockIBlockQuerier, pointObj *PointObject) {
	mockCtrl = gomock.NewController(t)
	mockBlockQuerier = NewMockIBlockQuerier(mockCtrl)
	pointObj = new(PointObject)
	return
}

func Test_PointObject_Tick_FallsToImmediateSurface(t *testing.T) {
	mockCtrl, mockBlockQuerier, pointObj := tickFixtures(t)
	defer mockCtrl.Finish()

	pointObj.Init(
		&AbsXyz{0.5, 100.1, 0.5},
		&AbsVelocity{},
	)

	mockBlockQuerier.EXPECT().BlockQuery(BlockXyz{0, 99, 0}).Return(true, true)
	pointObj.Tick(mockBlockQuerier)

	expectedBlockPos := BlockXyz{0, 100, 0}
	if !pointObj.position.ToBlockXyz().Equals(expectedBlockPos) {
		t.Errorf("Expected object to end at %#v but was at %#v", expectedBlockPos, pointObj.position)
	}
}
