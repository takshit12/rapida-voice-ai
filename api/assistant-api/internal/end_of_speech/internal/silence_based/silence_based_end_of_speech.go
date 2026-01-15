// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapidsilenceEOS.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapidsilenceEOS.ai for commercial usage.
package internal_silence_based

import (
	"context"
	"sync"
	"time"

	internaltype "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

// SpeechSegment represents accumulated speech with metadata
type SpeechSegment struct {
	ContextID string
	Text      string
	Timestamp time.Time
}

// command defines operations for the worker goroutine
type command struct {
	ctx     context.Context
	timeout time.Duration
	segment SpeechSegment
	fireNow bool
	reset   bool
}

// SilenceBasedEOS detects end-of-speech based on silence duration
type SilenceBasedEOS struct {
	logger         commons.Logger
	callback       internaltype.EndOfSpeechCallback
	silenceTimeout time.Duration

	// worker orchestration
	cmdCh  chan command
	stopCh chan struct{}

	// state
	mu    sync.RWMutex
	state *eosState
}

// eosState holds protected state for end-of-speech detection
type eosState struct {
	segment       SpeechSegment
	callbackFired bool
	generation    uint64
}

// NewSilenceBasedEOS creates a new silence-based end-of-speech detector
func NewSilenceBasedEndOfSpeech(
	logger commons.Logger,
	callback internaltype.EndOfSpeechCallback,
	opts utils.Option,
) (internaltype.EndOfSpeech, error) {
	threshold := 1000 * time.Millisecond
	if v, err := opts.GetFloat64("microphone.eos.timeout"); err == nil {
		threshold = time.Duration(v) * time.Millisecond
	}

	eos := &SilenceBasedEOS{
		logger:         logger,
		callback:       callback,
		silenceTimeout: threshold,
		cmdCh:          make(chan command, 32),
		stopCh:         make(chan struct{}),
		state: &eosState{
			segment: SpeechSegment{},
		},
	}

	go eos.worker()
	return eos, nil
}

// Name returns the component name
func (eos *SilenceBasedEOS) Name() string {
	return "silenceBasedEndOfSpeech"
}

// Analyze processes incoming speech packets
func (eos *SilenceBasedEOS) Analyze(ctx context.Context, pkt internaltype.Packet) error {
	switch p := pkt.(type) {

	case internaltype.UserTextPacket:
		if p.Text == "" {
			return nil
		}
		eos.mu.RLock()
		seg := SpeechSegment{ContextID: p.ContextId(), Text: p.Text, Timestamp: time.Now()}
		eos.state.segment = seg
		eos.mu.RUnlock()
		eos.send(command{
			ctx:     ctx,
			segment: seg,
			fireNow: true,
		})

	case internaltype.InterruptionPacket:
		eos.mu.RLock()
		seg := eos.state.segment
		eos.mu.RUnlock()

		if seg.Text == "" {
			return nil
		}

		eos.send(command{
			ctx:     ctx,
			segment: seg,
			timeout: eos.silenceTimeout,
		})

	case internaltype.SpeechToTextPacket:
		eos.mu.Lock()

		// skip interim-only updates with no prior content
		if p.Interim && eos.state.segment.Text == "" {
			eos.mu.Unlock()
			return nil
		}

		// build new segment
		newSeg := SpeechSegment{
			ContextID: p.ContextId(),
			Timestamp: time.Now(),
		}

		// text accumulation: interim extends, final replaces
		if p.Interim {
			newSeg.Text = eos.state.segment.Text + p.Script
		} else {
			newSeg.Text = p.Script
		}

		eos.state.segment = newSeg
		segCopy := newSeg
		eos.mu.Unlock()

		eos.send(command{
			ctx:     ctx,
			segment: segCopy,
			timeout: eos.silenceTimeout,
		})
	}

	return nil
}

// send dispatches a command to the worker
func (eos *SilenceBasedEOS) send(cmd command) {
	select {
	case eos.cmdCh <- cmd:
	default:
		go func() { eos.cmdCh <- cmd }()
	}
}

// worker manages silence detection and callback invocation
func (eos *SilenceBasedEOS) worker() {
	var (
		timer   *time.Timer
		timerC  <-chan time.Time
		gen     uint64
		ctx     context.Context
		segment SpeechSegment
	)

	cleanup := func() {
		if timer != nil {
			timer.Stop()
			timer = nil
			timerC = nil
		}
	}

	for {
		select {

		case <-eos.stopCh:
			cleanup()
			return

		case cmd := <-eos.cmdCh:
			eos.mu.Lock()

			// handle reset
			if cmd.reset {
				eos.state.callbackFired = false
				eos.state.generation++
				eos.state.segment = SpeechSegment{}
				eos.mu.Unlock()
				continue
			}

			// drop if callback pending
			if eos.state.callbackFired {
				eos.mu.Unlock()
				continue
			}

			// immediate fire
			if cmd.fireNow {
				eos.state.callbackFired = true
				seg := eos.state.segment
				cbCtx := cmd.ctx
				cleanup()
				eos.mu.Unlock()
				eos.fire(cbCtx, seg)
				continue
			}

			// schedule timer
			gen = eos.state.generation + 1
			eos.state.generation = gen
			ctx = cmd.ctx
			segment = cmd.segment
			cleanup()
			timer = time.NewTimer(cmd.timeout)
			timerC = timer.C
			eos.mu.Unlock()

		case <-timerC:
			eos.mu.Lock()

			// stale timer check
			if eos.state.callbackFired || gen != eos.state.generation {
				eos.mu.Unlock()
				continue
			}

			eos.state.callbackFired = true
			seg := segment
			cbCtx := ctx
			cleanup()
			eos.mu.Unlock()

			eos.fire(cbCtx, seg)
		}
	}
}

// fire triggers the callback and enqueues reset
func (eos *SilenceBasedEOS) fire(ctx context.Context, seg SpeechSegment) {
	if ctx.Err() != nil {
		return
	}

	eos.logger.Debugf(
		"EOS: contextID=%s, text=%q, duration=%dms",
		seg.ContextID, seg.Text, time.Since(seg.Timestamp).Milliseconds(),
	)

	_ = eos.callback(ctx, internaltype.EndOfSpeechPacket{
		Speech:    seg.Text,
		ContextID: seg.ContextID,
	})

	eos.send(command{reset: true})
}

// Close shuts down the detector
func (eos *SilenceBasedEOS) Close() error {
	close(eos.stopCh)
	return nil
}
