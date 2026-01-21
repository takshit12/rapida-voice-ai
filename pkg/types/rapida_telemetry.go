// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package types

type Telemetry interface {
	Type() string
}

func GetDifferentTelemetry(mtr []Telemetry) ([]*Event, []*Metric, []*Metadata) {
	evnts := make([]*Event, 0)
	mtrs := make([]*Metric, 0)
	meta := make([]*Metadata, 0)
	for _, telemetry := range mtr {
		switch t := telemetry.(type) {
		case *Metric:
			mtrs = append(mtrs, t)
		case *Event:
			evnts = append(evnts, t)
		case *Metadata:
			meta = append(meta, t)
		}
	}
	return evnts, mtrs, meta
}
