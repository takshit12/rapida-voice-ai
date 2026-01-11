// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_silence_based_end_of_speech

import (
	"context"
	"strings"
	"sync"
	"time"
	"unicode"

	internal_end_of_speech "github.com/rapidaai/api/assistant-api/internal/end_of_speech"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

type silenceBasedEndOfSpeech struct {
	logger            commons.Logger
	onCallback        internal_end_of_speech.EndOfSpeechCallback
	thresholdDuration time.Duration

	// worker coordination
	inputCh chan workerEvent
	stopCh  chan struct{}

	// protected state
	mutex         sync.Mutex
	callbackFired bool
	generation    uint64
	inputSpeech   string
}

type workerEvent struct {
	ctx     context.Context
	timeout time.Duration
	speech  string
	fireNow bool
	reset   bool
}

func NewSilenceBasedEndOfSpeech(
	logger commons.Logger,
	onCallback internal_end_of_speech.EndOfSpeechCallback,
	opts utils.Option,
) (internal_end_of_speech.EndOfSpeech, error) {

	duration := 1000 * time.Millisecond
	if v, err := opts.GetFloat64("microphone.eos.timeout"); err == nil {
		duration = time.Duration(v) * time.Millisecond
	}

	a := &silenceBasedEndOfSpeech{
		logger:            logger,
		onCallback:        onCallback,
		thresholdDuration: duration,
		inputCh:           make(chan workerEvent, 16),
		stopCh:            make(chan struct{}),
	}

	go a.worker()
	return a, nil
}

func (a *silenceBasedEndOfSpeech) Name() string {
	return "silenceBasedEndOfSpeech"
}

func (a *silenceBasedEndOfSpeech) Analyze(
	ctx context.Context,
	msg internal_end_of_speech.EndOfSpeechInput,
) error {

	switch input := msg.(type) {

	case *internal_end_of_speech.UserEndOfSpeechInput:
		a.enqueue(workerEvent{
			ctx:     ctx,
			speech:  input.GetMessage(),
			fireNow: true,
		})

	case *internal_end_of_speech.SystemEndOfSpeechInput:
		a.enqueue(workerEvent{
			ctx:     ctx,
			speech:  input.GetMessage(),
			timeout: a.thresholdDuration,
		})

	case *internal_end_of_speech.STTEndOfSpeechInput:
		a.handleSTT(ctx, input)
	}

	return nil
}

func (a *silenceBasedEndOfSpeech) handleSTT(
	ctx context.Context,
	input *internal_end_of_speech.STTEndOfSpeechInput,
) {
	a.mutex.Lock()

	timeout := a.thresholdDuration
	text := input.GetMessage()

	if input.IsComplete && a.inputSpeech != "" {
		if normalizeSTTText(a.inputSpeech) == normalizeSTTText(text) {
			timeout = a.thresholdDuration / 2
		}
	}

	a.inputSpeech = text
	a.mutex.Unlock()

	a.enqueue(workerEvent{
		ctx:     ctx,
		speech:  text,
		timeout: timeout,
	})
}

func (a *silenceBasedEndOfSpeech) enqueue(evt workerEvent) {
	select {
	case a.inputCh <- evt:
	default:
		// avoid deadlock under load
		go func() { a.inputCh <- evt }()
	}
}

func (a *silenceBasedEndOfSpeech) worker() {
	var (
		timer      *time.Timer
		timerC     <-chan time.Time
		generation uint64
		ctx        context.Context
		speech     string
	)

	stopTimer := func() {
		if timer != nil {
			timer.Stop()
			timer = nil
			timerC = nil
		}
	}

	for {
		select {
		case <-a.stopCh:
			stopTimer()
			return

		case evt := <-a.inputCh:

			// --- RESET EVENT (after callback) ---
			if evt.reset {
				a.mutex.Lock()
				a.callbackFired = false
				a.generation++
				a.inputSpeech = ""
				a.mutex.Unlock()
				continue
			}

			a.mutex.Lock()
			if a.callbackFired {
				a.mutex.Unlock()
				continue
			}

			a.generation++
			generation = a.generation

			if evt.fireNow {
				a.callbackFired = true
				stopTimer()
				a.mutex.Unlock()
				a.invokeCallback(evt.ctx, evt.speech)
				// Reset is enqueued by invokeCallback
				continue
			}

			ctx = evt.ctx
			speech = evt.speech

			stopTimer()
			timer = time.NewTimer(evt.timeout)
			timerC = timer.C

			a.mutex.Unlock()

		case <-timerC:
			a.mutex.Lock()
			if a.callbackFired || generation != a.generation {
				a.mutex.Unlock()
				continue
			}

			a.callbackFired = true
			text := speech
			cbCtx := ctx
			stopTimer()
			a.mutex.Unlock()

			a.invokeCallback(cbCtx, text)
			// Reset is enqueued by invokeCallback
		}
	}
}

func (a *silenceBasedEndOfSpeech) invokeCallback(
	ctx context.Context,
	speech string,
) {
	if speech == "" || ctx.Err() != nil {
		return
	}

	now := time.Now()
	seg := &internal_end_of_speech.EndOfSpeechResult{
		StartAt: float64(now.UnixNano()) / 1e9,
		EndAt:   float64(now.UnixNano()) / 1e9,
		Speech:  speech,
	}

	a.logger.Debugf("End of speech detected: '%s'", speech)
	_ = a.onCallback(ctx, seg)
	a.enqueue(workerEvent{reset: true})
}

func normalizeSTTText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, s)
}

func (a *silenceBasedEndOfSpeech) Close() error {
	close(a.stopCh)
	return nil
}
