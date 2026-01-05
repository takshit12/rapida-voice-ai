// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package gorm_models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	return db
}

func TestTimeWrapper_MarshalJSON(t *testing.T) {
	now := time.Now()
	tw := TimeWrapper(now)

	data, err := tw.MarshalJSON()
	assert.NoError(t, err)

	var ts timestamppb.Timestamp
	err = json.Unmarshal(data, &ts)
	assert.NoError(t, err)

	assert.True(t, ts.AsTime().Sub(now) < time.Second) // Close enough
}

func TestTimeWrapper_UnmarshalJSON(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)
	data, err := json.Marshal(ts)
	assert.NoError(t, err)

	var tw TimeWrapper
	err = tw.UnmarshalJSON(data)
	assert.NoError(t, err)

	assert.True(t, time.Time(tw).Sub(now) < time.Second)
}

func TestTimeWrapper_Value(t *testing.T) {
	now := time.Now()
	tw := TimeWrapper(now)

	value, err := tw.Value()
	assert.NoError(t, err)
	assert.Equal(t, now, value)
}

func TestAudited_BeforeCreate(t *testing.T) {
	// Test with zero CreatedDate and zero Id
	audited := &Audited{}
	err := audited.BeforeCreate(nil) // DB not used in method
	assert.NoError(t, err)
	assert.False(t, time.Time(audited.CreatedDate).IsZero())
	assert.True(t, audited.Id > 0)

	// Test with existing CreatedDate
	existingTime := time.Now().Add(-time.Hour)
	audited2 := &Audited{
		CreatedDate: TimeWrapper(existingTime),
		Id:          123,
	}
	err = audited2.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, TimeWrapper(existingTime), audited2.CreatedDate)
	assert.Equal(t, uint64(123), audited2.Id)
}

func TestAudited_BeforeUpdate(t *testing.T) {
	audited := &Audited{}
	err := audited.BeforeUpdate(nil)
	assert.NoError(t, err)
	assert.False(t, time.Time(audited.UpdatedDate).IsZero())
}

func TestMutable_JSONMarshaling(t *testing.T) {
	mutable := Mutable{
		Status:    "ACTIVE",
		CreatedBy: 123,
		UpdatedBy: 456,
	}

	data, err := json.Marshal(mutable)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"status":"ACTIVE"`)
	assert.Contains(t, string(data), `"createdBy":123`)
	assert.Contains(t, string(data), `"updatedBy":456`)
}

func TestOrganizational_JSONMarshaling(t *testing.T) {
	org := Organizational{
		ProjectId:      789,
		OrganizationId: 101112,
	}

	data, err := json.Marshal(org)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"projectId":789`)
	assert.Contains(t, string(data), `"organizationId":101112`)
}

func TestTimeWrapper_JSONRoundTrip(t *testing.T) {
	original := TimeWrapper(time.Now())

	// Marshal
	data, err := json.Marshal(original)
	assert.NoError(t, err)

	// Unmarshal
	var unmarshaled TimeWrapper
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	// Should be approximately equal
	assert.True(t, time.Time(unmarshaled).Sub(time.Time(original)) < time.Second)
}

func TestAudited_JSONMarshaling(t *testing.T) {
	now := time.Now()
	audited := Audited{
		Id:          12345,
		CreatedDate: TimeWrapper(now),
		UpdatedDate: TimeWrapper(now.Add(time.Hour)),
	}

	data, err := json.Marshal(audited)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, float64(12345), result["id"])
	assert.Contains(t, result, "createdDate")
	assert.Contains(t, result, "updatedDate")
}
