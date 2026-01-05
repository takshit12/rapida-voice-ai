// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package gorm_generator

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	// Test bit length constants
	assert.Equal(t, uint8(41), TimestampLength)
	assert.Equal(t, uint8(10), MachineIDLength)
	assert.Equal(t, uint8(12), SequenceLength)

	// Test max value constants
	assert.Equal(t, uint16(4095), MaxSequence)           // 2^12 - 1
	assert.Equal(t, uint64(2199023255551), MaxTimestamp) // 2^41 - 1
	assert.Equal(t, uint16(1023), MaxMachineID)          // 2^10 - 1

	// Test move length constants
	assert.Equal(t, SequenceLength, machineIDMoveLength)
	assert.Equal(t, MachineIDLength+SequenceLength, timestampMoveLength)
}

func TestAtomicResolver_BasicFunctionality(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0

	ms := int64(1000)
	seq, err := AtomicResolver(ms)

	assert.NoError(t, err)
	assert.Equal(t, uint16(0), seq) // First sequence should be 0
}

func TestAtomicResolver_SequenceIncrement(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0

	ms := int64(1000)

	// Generate multiple sequences for same timestamp
	seq1, err1 := AtomicResolver(ms)
	assert.NoError(t, err1)
	assert.Equal(t, uint16(0), seq1)

	seq2, err2 := AtomicResolver(ms)
	assert.NoError(t, err2)
	assert.Equal(t, uint16(1), seq2)

	seq3, err3 := AtomicResolver(ms)
	assert.NoError(t, err3)
	assert.Equal(t, uint16(2), seq3)
}

func TestAtomicResolver_SequenceOverflow(t *testing.T) {
	// Reset global state
	lastTime = 1000               // Same as ms
	lastSeq = uint32(MaxSequence) // Set to max

	ms := int64(1000)

	// Next sequence should overflow and return MaxSequence
	seq, err := AtomicResolver(ms)
	assert.NoError(t, err)
	assert.Equal(t, MaxSequence, seq)
}

func TestAtomicResolver_TimeTravel(t *testing.T) {
	// Reset global state
	lastTime = 2000
	lastSeq = 0

	ms := int64(1000) // Earlier time

	seq, err := AtomicResolver(ms)
	assert.NoError(t, err)
	assert.Equal(t, MaxSequence, seq) // Should return MaxSequence for time travel
}

func TestID_BasicGeneration(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	id := ID()
	assert.NotZero(t, id)

	// ID should be greater than 0
	assert.Greater(t, id, uint64(0))
}

func TestNextID_BasicGeneration(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	id, err := NextID()
	assert.NoError(t, err)
	assert.NotZero(t, id)
	assert.Greater(t, id, uint64(0))
}

func TestNextID_UniqueGeneration(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	// Generate multiple IDs
	ids := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		id, err := NextID()
		assert.NoError(t, err)
		assert.NotZero(t, id)
		assert.False(t, ids[id], "ID collision detected: %d", id)
		ids[id] = true
	}

	assert.Len(t, ids, 100)
}

func TestNextID_WithMachineID(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0

	// Set machine ID
	SetMachineID(5)
	defer func() { machineID = 0 }() // Reset after test

	id, err := NextID()
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// Parse ID and check machine ID
	sid := ParseID(id)
	assert.Equal(t, uint64(5), sid.MachineID)
}

func TestSetMachineID_ValidRange(t *testing.T) {
	// Test valid machine IDs
	validIDs := []uint16{0, 1, 100, 500, 1023}
	for _, id := range validIDs {
		assert.NotPanics(t, func() { SetMachineID(id) })
		machineID = 0 // Reset
	}
}

func TestSetMachineID_InvalidRange(t *testing.T) {
	// Test invalid machine IDs
	invalidIDs := []uint16{1024, 2000, 65535}
	for _, id := range invalidIDs {
		assert.Panics(t, func() { SetMachineID(id) })
	}
}

func TestSetSequenceResolver(t *testing.T) {
	// Reset resolver
	resolver = nil

	// Test setting nil resolver (should not change)
	SetSequenceResolver(nil)
	assert.Nil(t, resolver)

	// Test setting custom resolver
	customResolver := func(ms int64) (uint16, error) {
		return 42, nil
	}
	SetSequenceResolver(customResolver)

	// Verify resolver is set
	actualResolver := callSequenceResolver()
	seq, err := actualResolver(1000)
	assert.NoError(t, err)
	assert.Equal(t, uint16(42), seq)

	// Reset
	resolver = nil
}

func TestParseID_BasicParsing(t *testing.T) {
	// Create a known ID structure
	expectedTimestamp := uint64(123456)
	expectedMachineID := uint64(42)
	expectedSequence := uint64(789)

	// Manually construct ID
	id := (expectedTimestamp << uint64(timestampMoveLength)) |
		(expectedMachineID << uint64(machineIDMoveLength)) |
		expectedSequence

	sid := ParseID(id)

	assert.Equal(t, id, sid.ID)
	assert.Equal(t, expectedSequence, sid.Sequence)
	assert.Equal(t, expectedMachineID, sid.MachineID)
	assert.Equal(t, expectedTimestamp, sid.Timestamp)
}

func TestParseID_RoundTrip(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	// Generate an ID
	originalID, err := NextID()
	assert.NoError(t, err)

	// Parse it
	sid := ParseID(originalID)

	// Verify the parsed components make sense
	assert.Equal(t, originalID, sid.ID)
	assert.LessOrEqual(t, sid.Sequence, uint64(MaxSequence))
	assert.LessOrEqual(t, sid.MachineID, uint64(MaxMachineID))
	assert.Greater(t, sid.Timestamp, uint64(0))
}

func TestSID_GenerateTime(t *testing.T) {
	// Create a SID with known timestamp
	sid := SID{
		Timestamp: 1000, // 1000ms after start time
	}

	generatedTime := sid.GenerateTime()
	expectedTime := startTime.Add(1000 * time.Millisecond)

	assert.Equal(t, expectedTime.UTC(), generatedTime)
}

func TestElapsedTime(t *testing.T) {
	testTime := time.Date(2008, 11, 10, 23, 0, 1, 0, time.UTC) // 1 second after start
	ms := testTime.UTC().UnixNano() / 1e6

	elapsed := elapsedTime(ms, startTime)
	assert.Equal(t, int64(1000), elapsed) // 1 second = 1000ms
}

func TestCurrentMillis(t *testing.T) {
	ms := currentMillis()
	assert.Greater(t, ms, int64(0))

	// Should be close to current time
	now := time.Now().UTC().UnixNano() / 1e6
	assert.InDelta(t, now, ms, 100) // Within 100ms
}

func TestWaitForNextMillis(t *testing.T) {
	last := int64(1000)
	next := waitForNextMillis(last)

	// Should return a time greater than last
	assert.Greater(t, next, last)
}

func TestNextID_SequenceOverflowHandling(t *testing.T) {
	callCount := 0
	// Set up a resolver that returns MaxSequence first, then 0
	overflowResolver := func(ms int64) (uint16, error) {
		callCount++
		if callCount == 1 {
			return MaxSequence, nil
		}
		return 0, nil
	}
	SetSequenceResolver(overflowResolver)
	defer func() { resolver = nil }()

	// This should handle overflow by waiting for next millisecond
	id, err := NextID()
	assert.NoError(t, err)
	assert.NotZero(t, id)
	assert.GreaterOrEqual(t, callCount, 2) // Should have been called at least twice
}

func TestNextID_TimeOverflow(t *testing.T) {
	// Save original start time
	originalStartTime := startTime
	defer func() { startTime = originalStartTime }()

	// Set start time to year 2100 (far future)
	startTime = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

	// This should fail due to time overflow
	id, err := NextID()
	assert.Error(t, err)
	assert.Zero(t, id)
	assert.Contains(t, err.Error(), "maximum life cycle")
}

func TestNextID_NegativeElapsedTime(t *testing.T) {
	// Save original start time
	originalStartTime := startTime
	defer func() { startTime = originalStartTime }()

	// Set start time to future
	startTime = time.Now().Add(24 * time.Hour)

	// This should fail due to negative elapsed time
	id, err := NextID()
	assert.Error(t, err)
	assert.Zero(t, id)
	assert.Contains(t, err.Error(), "maximum life cycle")
}

func TestID_IgnoresErrors(t *testing.T) {
	// Set up a resolver that returns an error
	errorResolver := func(ms int64) (uint16, error) {
		return 0, errors.New("test error")
	}
	SetSequenceResolver(errorResolver)
	defer func() { resolver = nil }()

	// ID() should ignore errors and return 0
	id := ID()
	assert.Zero(t, id)
}

func TestCallSequenceResolver(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0

	// Test default resolver
	resolver = nil
	actualResolver := callSequenceResolver()
	assert.NotNil(t, actualResolver)

	// Should return AtomicResolver by default
	seq, err := actualResolver(1000)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0), seq)
}

func TestThreadSafety_ID(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	var wg sync.WaitGroup
	ids := make(chan uint64, 100)

	// Generate IDs from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				id := ID()
				ids <- id
			}
		}()
	}

	// Wait for all goroutines
	go func() {
		wg.Wait()
		close(ids)
	}()

	// Collect results
	idMap := make(map[uint64]bool)
	count := 0
	for id := range ids {
		assert.NotZero(t, id)
		assert.False(t, idMap[id], "Duplicate ID found: %d", id)
		idMap[id] = true
		count++
	}

	assert.Equal(t, 100, count)
	assert.Len(t, idMap, 100)
}

func TestThreadSafety_NextID(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	var wg sync.WaitGroup
	ids := make(chan uint64, 100)
	errors := make(chan error, 100)

	// Generate IDs from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				id, err := NextID()
				ids <- id
				errors <- err
			}
		}()
	}

	// Wait for all goroutines
	go func() {
		wg.Wait()
		close(ids)
		close(errors)
	}()

	// Collect results
	idMap := make(map[uint64]bool)
	count := 0
	for i := 0; i < 100; i++ {
		id := <-ids
		err := <-errors
		assert.NoError(t, err)
		assert.NotZero(t, id)
		assert.False(t, idMap[id], "Duplicate ID found: %d", id)
		idMap[id] = true
		count++
	}

	assert.Equal(t, 100, count)
	assert.Len(t, idMap, 100)
}

func TestParseID_EdgeCases(t *testing.T) {
	// Test with zero ID
	sid := ParseID(0)
	assert.Equal(t, uint64(0), sid.ID)
	assert.Equal(t, uint64(0), sid.Sequence)
	assert.Equal(t, uint64(0), sid.MachineID)
	assert.Equal(t, uint64(0), sid.Timestamp)

	// Test with max possible ID
	maxID := uint64(1)<<63 - 1 // Max uint64
	sid = ParseID(maxID)
	assert.Equal(t, maxID, sid.ID)
	assert.Equal(t, uint64(MaxSequence), sid.Sequence)
	assert.Equal(t, uint64(MaxMachineID), sid.MachineID)
	// Timestamp will be maxID >> (SequenceLength + MachineIDLength)
}

func TestID_Consistency(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	// Generate ID and parse it
	id := ID()
	sid := ParseID(id)

	// Verify consistency
	assert.Equal(t, id, sid.ID)

	// Extract components manually for verification
	manualTimestamp := id >> uint64(SequenceLength+MachineIDLength)
	manualSequence := id & uint64(MaxSequence)
	manualMachineID := (id & (uint64(MaxMachineID) << SequenceLength)) >> SequenceLength

	assert.Equal(t, manualTimestamp, sid.Timestamp)
	assert.Equal(t, manualSequence, sid.Sequence)
	assert.Equal(t, manualMachineID, sid.MachineID)
}

func TestStartTime(t *testing.T) {
	expected := time.Date(2008, 11, 10, 23, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, startTime)
}

func TestSequenceResolverType(t *testing.T) {
	var resolver SequenceResolver = AtomicResolver
	assert.NotNil(t, resolver)

	// Test function signature
	seq, err := resolver(1000)
	assert.IsType(t, uint16(0), seq)
	assert.IsType(t, error(nil), err)
}

func TestBitOperations(t *testing.T) {
	// Test bit shift operations used in ID construction
	timestamp := uint64(123456)
	machineID := uint64(42)
	sequence := uint64(789)

	id := (timestamp << uint64(timestampMoveLength)) |
		(machineID << uint64(machineIDMoveLength)) |
		sequence

	// Verify we can extract the components back
	extractedTimestamp := id >> uint64(SequenceLength+MachineIDLength)
	extractedSequence := id & uint64(MaxSequence)
	extractedMachineID := (id & (uint64(MaxMachineID) << SequenceLength)) >> SequenceLength

	assert.Equal(t, timestamp, extractedTimestamp)
	assert.Equal(t, sequence, extractedSequence)
	assert.Equal(t, machineID, extractedMachineID)
}

func TestMaxValues(t *testing.T) {
	// Test that max values are correctly calculated
	assert.Equal(t, uint16(1<<12-1), MaxSequence)
	assert.Equal(t, uint64(1<<41-1), MaxTimestamp)
	assert.Equal(t, uint16(1<<10-1), MaxMachineID)
}

func TestID_GenerationOverTime(t *testing.T) {
	// Reset global state
	lastTime = 0
	lastSeq = 0
	machineID = 0

	// Generate IDs with small delay to ensure different timestamps
	id1 := ID()
	time.Sleep(1 * time.Millisecond)
	id2 := ID()

	// IDs should be different
	assert.NotEqual(t, id1, id2)

	// Parse both IDs
	sid1 := ParseID(id1)
	sid2 := ParseID(id2)

	// Timestamps should be different or same (if generated in same ms)
	// but sequences should be different if same timestamp
	if sid1.Timestamp == sid2.Timestamp {
		assert.NotEqual(t, sid1.Sequence, sid2.Sequence)
	}
}
