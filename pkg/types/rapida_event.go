// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Event struct {
	EventType string                 `protobuf:"bytes,1,opt,name=eventType,proto3" json:"eventType,omitempty"`
	Payload   map[string]interface{} `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (d *Event) SetValue(src interface{}) error {
	switch v := src.(type) {
	case string:
		d.Payload = map[string]interface{}{"value": v}
	case []byte:
		d.Payload = map[string]interface{}{"value": string(v)}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		d.Payload = map[string]interface{}{"value": fmt.Sprintf("%d", v)}
	case float32, float64:
		d.Payload = map[string]interface{}{"value": fmt.Sprintf("%f", v)}
	case bool:
		d.Payload = map[string]interface{}{"value": strconv.FormatBool(v)}
	case nil:
		d.Payload = map[string]interface{}{}
	default:
		// Use JSON marshalling for all other types
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal Payload: %w", err)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return fmt.Errorf("failed to unmarshal into InterfaceMap: %w", err)
		}
		d.Payload = result
	}
	return nil
}

func (d *Event) Type() string {
	return "event"
}

func NewEvent(k string, v interface{}) *Event {
	md := &Event{
		EventType: k,
	}
	md.SetValue(v)
	return md
}
