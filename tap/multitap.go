package tap

import (
	"context"

	"github.com/myhops/httptap"
)

type multiTap []httptap.Tap

func NewMultiTap(taps []httptap.Tap) httptap.Tap {
	return multiTap(taps)
}

func (t multiTap) Serve(ctx context.Context, rr *httptap.RequestResponse) {
	for _, tt := range t {
		tt.Serve(ctx, rr)
	}
}
