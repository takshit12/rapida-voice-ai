// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package types

import (
	"fmt"
	"time"

	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type Metric struct {
	Name        string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value       string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
}

func (d *Metric) Type() string {
	return "metric"
}

func (x *Metric) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Metric) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Metric) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Metric) ToProto() *protos.Metric {
	out := &protos.Metric{}
	_ = utils.Cast(x, out)
	return out
}

type Metrics []*Metric

func (m Metrics) ToProto() []*protos.Metric {
	out := make([]*protos.Metric, len(m))
	_ = utils.Cast(m, &out)
	return out
}

func ToMetric(mtr *protos.Metric) *Metric {
	out := &Metric{}
	err := utils.Cast(mtr, out)
	if err != nil {
		return nil
	}
	return out
}

func ToMetrics(mtr []*protos.Metric) []*Metric {
	out := make([]*Metric, len(mtr))
	for idx, k := range mtr {
		out[idx] = ToMetric(k)
	}
	return out
}

func NewMetric(name string, val string, description *string) *Metric {
	met := &Metric{
		Name:  name,
		Value: val,
	}
	if description != nil {
		met.Description = *description
	}
	return met
}

func NewTimeTakenMetric(duration time.Duration) *Metric {
	return NewMetric(type_enums.TIME_TAKEN.String(), fmt.Sprintf("%d", duration), utils.Ptr("Time taken for given task"))
}

func NewInputTokenMetric(count int) *Metric {
	return NewMetric(type_enums.INPUT_TOKEN.String(), fmt.Sprintf("%d", count), utils.Ptr("Number of input tokens"))
}

func NewOutputTokenMetric(count int) *Metric {
	return NewMetric(type_enums.OUTPUT_TOKEN.String(), fmt.Sprintf("%d", count), utils.Ptr("Number of output tokens"))
}

func NewTotalTokenMetric(count int) *Metric {
	return NewMetric(type_enums.TOTAL_TOKEN.String(), fmt.Sprintf("%d", count), utils.Ptr("Total number of tokens"))
}

func NewInputCostMetric(cost float64) *Metric {
	return NewMetric(type_enums.INPUT_COST.String(), fmt.Sprintf("%.6f", cost), utils.Ptr("Cost for input tokens"))
}

func NewOutputCostMetric(cost float64) *Metric {
	return NewMetric(type_enums.OUTPUT_COST.String(), fmt.Sprintf("%.6f", cost), utils.Ptr("Cost for output tokens"))
}

func NewTotalCostMetric(cost float64) *Metric {
	return NewMetric(type_enums.COST.String(), fmt.Sprintf("%.6f", cost), utils.Ptr("Total cost for the operation"))
}

func NewStatusMetric(status type_enums.RecordState) *Metric {
	return NewMetric(type_enums.STATUS.String(), status.String(), utils.Ptr("Status of the operation"))
}
