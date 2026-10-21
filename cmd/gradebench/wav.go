package main

// Audio for paired samples, prepared the way the browser prepares a learner's
// recording: pauses measured with the recorder's level meter
// (prepyo/src/lib/delivery.ts, AudioRecorder.tsx), then mono 16 kHz, silence
// trimmed at both ends (silence.ts) and uploaded as 16-bit WAV (audio.ts).

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/prepyo/backend/internal/scoring"
)

// readWAV decodes a PCM (8/16/24/32-bit) or float WAV into mono samples in [-1, 1].
func readWAV(path string) ([]float64, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, 0, fmt.Errorf("%s is not a WAV file; save the recording as WAV", path)
	}
	var format, channels, bits int
	var rate int
	var pcm []byte
	for p := 12; p+8 <= len(data); {
		id, size := string(data[p:p+4]), int(binary.LittleEndian.Uint32(data[p+4:p+8]))
		body := data[p+8 : min(len(data), p+8+size)]
		switch id {
		case "fmt ":
			if len(body) < 16 {
				return nil, 0, fmt.Errorf("%s: short fmt chunk", path)
			}
			format = int(binary.LittleEndian.Uint16(body[0:2]))
			channels = int(binary.LittleEndian.Uint16(body[2:4]))
			rate = int(binary.LittleEndian.Uint32(body[4:8]))
			bits = int(binary.LittleEndian.Uint16(body[14:16]))
			if format == 0xFFFE && len(body) >= 26 { // WAVE_FORMAT_EXTENSIBLE: real format in the sub-format GUID
				format = int(binary.LittleEndian.Uint16(body[24:26]))
			}
		case "data":
			pcm = body
		}
		p += 8 + size + size%2
	}
	if channels == 0 || rate == 0 || pcm == nil || (format != 1 && format != 3) {
		return nil, 0, fmt.Errorf("%s: unsupported WAV (format %d); use PCM or float WAV", path, format)
	}
	width := bits / 8
	frames := len(pcm) / (width * channels)
	out := make([]float64, frames)
	for f := 0; f < frames; f++ {
		sum := 0.0
		for c := 0; c < channels; c++ {
			b := pcm[(f*channels+c)*width:]
			var v float64
			switch {
			case format == 3 && bits == 32:
				v = float64(math.Float32frombits(binary.LittleEndian.Uint32(b)))
			case bits == 8:
				v = (float64(b[0]) - 128) / 128
			case bits == 16:
				v = float64(int16(binary.LittleEndian.Uint16(b))) / 32768
			case bits == 24:
				v = float64(int32(uint32(b[0])<<8|uint32(b[1])<<16|uint32(b[2])<<24)>>8) / 8388608
			case bits == 32:
				v = float64(int32(binary.LittleEndian.Uint32(b))) / 2147483648
			default:
				return nil, 0, fmt.Errorf("%s: %d-bit samples not supported", path, bits)
			}
			sum += v
		}
		out[f] = sum / float64(channels)
	}
	return out, rate, nil
}

// measureDelivery runs the recorder's meter over the recording: a level
// (RMS x4 of the analyser's 1024-sample window at 48 kHz) every 1/60 s.
func measureDelivery(samples []float64, rate int) scoring.Delivery {
	const voiceLevel, pauseMs, frameMs = 0.06, 600.0, 1000.0 / 60
	window := max(1, int(math.Round(1024*float64(rate)/48000)))
	hop := float64(rate) / 60
	var measured, voiced, silenceRun, leadIn, longest float64
	started, pauses := false, 0
	closeRun := func() {
		if started && silenceRun >= pauseMs {
			pauses++
			longest = math.Max(longest, silenceRun)
		}
		silenceRun = 0
	}
	for pos := float64(window); int(pos) <= len(samples); pos += hop {
		end := int(pos)
		sum := 0.0
		for _, s := range samples[end-window : end] {
			sum += s * s
		}
		level := math.Min(1, math.Sqrt(sum/float64(window))*4)
		measured += frameMs
		if level >= voiceLevel {
			if !started {
				leadIn = silenceRun
			}
			closeRun()
			started = true
			voiced += frameMs
			continue
		}
		silenceRun += frameMs
	}
	duration := measured
	if started {
		duration = measured - leadIn - silenceRun
	}
	ratio := 0.0
	if duration > 0 {
		ratio = math.Min(1, voiced/duration)
	}
	return scoring.Delivery{
		DurationSeconds:     int(math.Round(duration / 1000)),
		PauseCount:          pauses,
		LongestPauseSeconds: math.Round(longest/100) / 10,
		SpeakingRatio:       ratio,
	}
}

// toMono16k resamples to 16 kHz, averaging the source samples behind each output sample.
func toMono16k(samples []float64, rate int) []float64 {
	if rate == 16000 {
		return samples
	}
	step := float64(rate) / 16000
	out := make([]float64, int(float64(len(samples))/step))
	for i := range out {
		lo, hi := int(float64(i)*step), int(float64(i+1)*step)
		hi = max(lo+1, min(hi, len(samples)))
		sum := 0.0
		for _, s := range samples[lo:hi] {
			sum += s
		}
		out[i] = sum / float64(hi-lo)
	}
	return out
}

// trimSilence cuts silence before the first word and after the last, as silence.ts does.
func trimSilence(samples []float64, rate int) []float64 {
	frame := max(1, int(math.Round(float64(rate)*0.02)))
	frames := len(samples) / frame
	if frames < 2 {
		return samples
	}
	levels := make([]float64, frames)
	for f := range levels {
		sum := 0.0
		for _, s := range samples[f*frame : (f+1)*frame] {
			sum += s * s
		}
		levels[f] = math.Sqrt(sum / float64(frame))
	}
	sorted := append([]float64(nil), levels...)
	sort.Float64s(sorted)
	threshold := math.Min(math.Max(0.003, sorted[int(float64(frames)*0.1)]*4), sorted[frames-1]/2)
	first, last := -1, -1
	for f, l := range levels {
		if l >= threshold {
			if first < 0 {
				first = f
			}
			last = f
		}
	}
	if first < 0 {
		return samples
	}
	pad := int(math.Round(float64(rate) * 0.5))
	start, end := max(0, first*frame-pad), min(len(samples), (last+1)*frame+pad)
	if float64(len(samples)-(end-start)) < float64(rate)*0.5 {
		return samples
	}
	return samples[start:end]
}

// encodeWAV16 writes mono 16-bit PCM.
func encodeWAV16(samples []float64, rate int) []byte {
	var b bytes.Buffer
	n := len(samples) * 2
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(36+n))
	b.WriteString("WAVEfmt ")
	for _, v := range []any{uint32(16), uint16(1), uint16(1), uint32(rate), uint32(rate * 2), uint16(2), uint16(16)} {
		binary.Write(&b, binary.LittleEndian, v)
	}
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, uint32(n))
	for _, s := range samples {
		c := math.Max(-1, math.Min(1, s))
		v := int16(c * 32767)
		if c < 0 {
			v = int16(c * 32768)
		}
		binary.Write(&b, binary.LittleEndian, v)
	}
	return b.Bytes()
}

// prepareRecording returns the upload the app would send and what the browser would have measured.
func prepareRecording(path string) ([]byte, scoring.Delivery, int, error) {
	if !strings.EqualFold(strings.TrimPrefix(fileExt(path), "."), "wav") {
		return nil, scoring.Delivery{}, 0, fmt.Errorf("%s: recordings must be WAV (Sound Recorder: Settings > Recording format > WAV)", path)
	}
	samples, rate, err := readWAV(path)
	if err != nil {
		return nil, scoring.Delivery{}, 0, err
	}
	delivery := measureDelivery(samples, rate)
	seconds := int(math.Round(float64(len(samples)) / float64(rate)))
	mono := trimSilence(toMono16k(samples, rate), 16000)
	return encodeWAV16(mono, 16000), delivery, seconds, nil
}

func fileExt(path string) string {
	if i := strings.LastIndex(path, "."); i >= 0 {
		return path[i:]
	}
	return ""
}
