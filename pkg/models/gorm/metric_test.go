// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package gorm_models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	name := "test_metric"
	value := "42"
	description := "A test metric"

	metric := NewMetric(name, value, description)

	assert.NotNil(t, metric)
	assert.Equal(t, name, metric.Name)
	assert.Equal(t, value, metric.Value)
	assert.Equal(t, description, metric.Description)
}

func TestMetric_JSONMarshaling(t *testing.T) {
	metric := Metric{
		Name:        "cpu_usage",
		Value:       "85.5",
		Description: "CPU usage percentage",
	}

	data, err := json.Marshal(metric)
	assert.NoError(t, err)

	expected := `{"name":"cpu_usage","value":"85.5","description":"CPU usage percentage"}`
	assert.JSONEq(t, expected, string(data))
}

func TestMetric_JSONUnmarshaling(t *testing.T) {
	jsonStr := `{"name":"memory_usage","value":"1024","description":"Memory usage in MB"}`

	var metric Metric
	err := json.Unmarshal([]byte(jsonStr), &metric)
	assert.NoError(t, err)

	assert.Equal(t, "memory_usage", metric.Name)
	assert.Equal(t, "1024", metric.Value)
	assert.Equal(t, "Memory usage in MB", metric.Description)
}

func TestMetric_EdgeCases(t *testing.T) {
	t.Run("empty strings", func(t *testing.T) {
		metric := NewMetric("", "", "")
		assert.Equal(t, "", metric.Name)
		assert.Equal(t, "", metric.Value)
		assert.Equal(t, "", metric.Description)
	})

	t.Run("special characters", func(t *testing.T) {
		name := "metric/with/slashes"
		value := "value with spaces & symbols: @#$%"
		description := "Description with\nnewlines\tand\ttabs"

		metric := NewMetric(name, value, description)
		assert.Equal(t, name, metric.Name)
		assert.Equal(t, value, metric.Value)
		assert.Equal(t, description, metric.Description)
	})

	t.Run("long strings", func(t *testing.T) {
		longName := string(make([]byte, 1000)) // 1000 chars
		for i := range longName {
			longName = longName[:i] + "a" + longName[i+1:]
		}
		longValue := string(make([]byte, 1000))
		for i := range longValue {
			longValue = longValue[:i] + "b" + longValue[i+1:]
		}
		longDesc := string(make([]byte, 1000))
		for i := range longDesc {
			longDesc = longDesc[:i] + "c" + longDesc[i+1:]
		}

		metric := NewMetric(longName, longValue, longDesc)
		assert.Equal(t, longName, metric.Name)
		assert.Equal(t, longValue, metric.Value)
		assert.Equal(t, longDesc, metric.Description)
	})
}
