package voice

import (
	"sync"

	domain "verification-platform/internal/domain/voice"
)

// AudioChunk represents a verified audio payload from the WebSocket.
type AudioChunk struct {
	Sequence    uint32
	TimestampMs uint64
	Data        []byte
}

// AudioBuffer provides a thread-safe, bounded queue for incoming audio chunks.
// It applies backpressure by dropping frames if the buffer fills up, and
// ensures sequence numbers are strictly monotonic.
type AudioBuffer struct {
	mu           sync.Mutex
	chunks       []AudioChunk
	maxSize      int
	lastSequence uint32
	isFirst      bool
}

func NewAudioBuffer(maxSize int) *AudioBuffer {
	return &AudioBuffer{
		chunks:  make([]AudioChunk, 0, maxSize),
		maxSize: maxSize,
		isFirst: true,
	}
}

// Push adds a chunk to the buffer.
// Returns ErrSequenceInvalid if the sequence is out of order.
// Returns ErrBufferOverflow if the buffer is full (chunk dropped).
func (b *AudioBuffer) Push(chunk AudioChunk) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Sequence validation
	if !b.isFirst && chunk.Sequence <= b.lastSequence {
		return domain.ErrSequenceInvalid
	}

	// Backpressure validation
	if len(b.chunks) >= b.maxSize {
		return domain.ErrBufferOverflow
	}

	b.chunks = append(b.chunks, chunk)
	b.lastSequence = chunk.Sequence
	b.isFirst = false

	return nil
}

// Pop extracts the oldest chunk from the buffer.
// Returns false if empty.
func (b *AudioBuffer) Pop() (AudioChunk, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.chunks) == 0 {
		return AudioChunk{}, false
	}

	chunk := b.chunks[0]
	// Shift remaining left
	b.chunks = b.chunks[1:]
	return chunk, true
}

// Len returns the current number of chunks in the buffer.
func (b *AudioBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.chunks)
}
