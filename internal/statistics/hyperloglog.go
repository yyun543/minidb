package statistics

import (
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"math/bits"
)

// HyperLogLog implements the HyperLogLog cardinality estimation algorithm
// Provides ~2% accuracy with 16KB memory for billions of distinct values
type HyperLogLog struct {
	registers [16384]uint8 // 2^14 = 16384 buckets
	precision int          // 14 bits
	alpha     float64      // Bias correction factor
}

// NewHyperLogLog creates a new HyperLogLog estimator
func NewHyperLogLog() *HyperLogLog {
	// Alpha constant for m = 16384
	alpha := 0.7213 / (1 + 1.079/16384.0)

	return &HyperLogLog{
		registers: [16384]uint8{},
		precision: 14,
		alpha:     alpha,
	}
}

// Add adds a value to the HyperLogLog
func (hll *HyperLogLog) Add(value interface{}) {
	// Compute hash
	hash := hashValue(value)

	// Extract bucket index (first 14 bits)
	bucketIdx := hash >> (64 - hll.precision)

	// Count leading zeros in remaining bits
	w := hash << hll.precision
	leadingZeros := uint8(bits.LeadingZeros64(w)) + 1

	// Update bucket with maximum leading zeros
	if leadingZeros > hll.registers[bucketIdx] {
		hll.registers[bucketIdx] = leadingZeros
	}
}

// Estimate returns the estimated cardinality
func (hll *HyperLogLog) Estimate() int64 {
	// Compute raw estimate using harmonic mean
	sum := 0.0
	zeros := 0
	for _, reg := range hll.registers {
		sum += 1.0 / math.Pow(2, float64(reg))
		if reg == 0 {
			zeros++
		}
	}

	m := float64(len(hll.registers))
	rawEstimate := hll.alpha * m * m / sum

	// Apply bias corrections for different ranges
	var estimate float64

	if rawEstimate <= 2.5*m {
		// Small range correction - use linear counting
		if zeros != 0 {
			estimate = m * math.Log(m/float64(zeros))
		} else {
			estimate = rawEstimate
		}
	} else if rawEstimate <= math.Pow(2, 32)/30.0 {
		// Intermediate range - no correction
		estimate = rawEstimate
	} else {
		// Large range correction
		estimate = -math.Pow(2, 32) * math.Log(1.0-rawEstimate/math.Pow(2, 32))
	}

	return int64(estimate)
}

// Merge combines this HyperLogLog with another
// Useful for distributed statistics collection
func (hll *HyperLogLog) Merge(other *HyperLogLog) error {
	if len(hll.registers) != len(other.registers) {
		return nil // Silently skip incompatible HLL
	}

	// Take maximum of each register
	for i := range hll.registers {
		if other.registers[i] > hll.registers[i] {
			hll.registers[i] = other.registers[i]
		}
	}

	return nil
}

// Reset clears all registers
func (hll *HyperLogLog) Reset() {
	for i := range hll.registers {
		hll.registers[i] = 0
	}
}

// hashValue computes a 64-bit hash of any value
func hashValue(value interface{}) uint64 {
	h := fnv.New64a()

	switch v := value.(type) {
	case int:
		writeInt64(h, int64(v))
	case int8:
		writeInt64(h, int64(v))
	case int16:
		writeInt64(h, int64(v))
	case int32:
		writeInt64(h, int64(v))
	case int64:
		writeInt64(h, v)
	case uint:
		writeUint64(h, uint64(v))
	case uint8:
		writeUint64(h, uint64(v))
	case uint16:
		writeUint64(h, uint64(v))
	case uint32:
		writeUint64(h, uint64(v))
	case uint64:
		writeUint64(h, v)
	case float32:
		writeFloat64(h, float64(v))
	case float64:
		writeFloat64(h, v)
	case string:
		h.Write([]byte(v))
	case []byte:
		h.Write(v)
	case bool:
		if v {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
	case nil:
		h.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	default:
		// Fallback: use fmt.Sprint
		h.Write([]byte(stringValue(v)))
	}

	return h.Sum64()
}

// Helper functions for writing primitive types
func writeInt64(h hash.Hash, v int64) {
	buf := [8]byte{}
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
	buf[2] = byte(v >> 16)
	buf[3] = byte(v >> 24)
	buf[4] = byte(v >> 32)
	buf[5] = byte(v >> 40)
	buf[6] = byte(v >> 48)
	buf[7] = byte(v >> 56)
	h.Write(buf[:])
}

func writeUint64(h hash.Hash, v uint64) {
	buf := [8]byte{}
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
	buf[2] = byte(v >> 16)
	buf[3] = byte(v >> 24)
	buf[4] = byte(v >> 32)
	buf[5] = byte(v >> 40)
	buf[6] = byte(v >> 48)
	buf[7] = byte(v >> 56)
	h.Write(buf[:])
}

func writeFloat64(h hash.Hash, v float64) {
	writeUint64(h, math.Float64bits(v))
}

// stringValue converts any value to string representation
func stringValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", v)
}
