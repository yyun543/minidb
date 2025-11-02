package statistics

import (
	"fmt"
	"math"
	"sort"
)

// EquiHeightHistogram implements an equi-height (equi-depth) histogram
// Each bucket contains approximately the same number of rows
// Better for skewed data distributions than equi-width histograms
type EquiHeightHistogram struct {
	Buckets    []EquiHeightBucket
	BucketSize int // Target rows per bucket
	TotalRows  int
}

// EquiHeightBucket represents one bucket in the equi-height histogram
type EquiHeightBucket struct {
	LowerBound     interface{}
	UpperBound     interface{}
	RowCount       int // Actual row count (≈ BucketSize)
	DistinctValues int // Number of distinct values in bucket
}

// BuildEquiHeightHistogram constructs an equi-height histogram from sorted values
func BuildEquiHeightHistogram(values []interface{}, bucketCount int) *EquiHeightHistogram {
	if len(values) < bucketCount {
		// Not enough data for requested buckets
		return nil
	}

	// Sort values
	sorted := make([]interface{}, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return compareValues(sorted[i], sorted[j]) < 0
	})

	// Calculate bucket size (rows per bucket)
	bucketSize := len(sorted) / bucketCount

	// Build buckets
	histogram := &EquiHeightHistogram{
		Buckets:    make([]EquiHeightBucket, 0, bucketCount),
		BucketSize: bucketSize,
		TotalRows:  len(sorted),
	}

	for i := 0; i < bucketCount; i++ {
		start := i * bucketSize
		end := start + bucketSize
		if i == bucketCount-1 {
			// Last bucket gets remaining values
			end = len(sorted)
		}

		bucketValues := sorted[start:end]
		if len(bucketValues) == 0 {
			continue
		}

		histogram.Buckets = append(histogram.Buckets, EquiHeightBucket{
			LowerBound:     bucketValues[0],
			UpperBound:     bucketValues[len(bucketValues)-1],
			RowCount:       len(bucketValues),
			DistinctValues: countDistinctValues(bucketValues),
		})
	}

	return histogram
}

// EstimateSelectivity estimates the fraction of rows matching a predicate
// Returns value between 0.0 and 1.0
func (hist *EquiHeightHistogram) EstimateSelectivity(op string, value interface{}) float64 {
	if len(hist.Buckets) == 0 {
		return 0.5 // Default 50% when no histogram
	}

	switch op {
	case "=", "==":
		return hist.estimateEquality(value)
	case ">":
		return hist.estimateGreaterThan(value, false)
	case ">=":
		return hist.estimateGreaterThan(value, true)
	case "<":
		return hist.estimateLessThan(value, false)
	case "<=":
		return hist.estimateLessThan(value, true)
	case "!=", "<>":
		return 1.0 - hist.estimateEquality(value)
	default:
		return 0.5 // Unknown operator, default 50%
	}
}

// estimateEquality estimates selectivity for equality predicate (col = value)
func (hist *EquiHeightHistogram) estimateEquality(value interface{}) float64 {
	// Find bucket containing the value
	for _, bucket := range hist.Buckets {
		if compareValues(value, bucket.LowerBound) >= 0 &&
			compareValues(value, bucket.UpperBound) <= 0 {
			// Value is in this bucket
			// Assume uniform distribution within bucket
			// Selectivity = (1 / distinct_values_in_bucket) * (bucket_rows / total_rows)
			if bucket.DistinctValues > 0 {
				bucketSelectivity := 1.0 / float64(bucket.DistinctValues)
				return bucketSelectivity * (float64(bucket.RowCount) / float64(hist.TotalRows))
			}
		}
	}

	// Value not in any bucket - very rare
	return 0.001
}

// estimateGreaterThan estimates selectivity for > or >= predicate
func (hist *EquiHeightHistogram) estimateGreaterThan(value interface{}, inclusive bool) float64 {
	matchingRows := 0

	for _, bucket := range hist.Buckets {
		cmpLower := compareValues(value, bucket.LowerBound)
		cmpUpper := compareValues(value, bucket.UpperBound)

		if cmpUpper < 0 || (cmpUpper == 0 && inclusive) {
			// Value <= bucket upper bound (with optional equality)
			if cmpLower < 0 || (cmpLower == 0 && !inclusive) {
				// Value < bucket lower bound, entire bucket matches
				matchingRows += bucket.RowCount
			} else if cmpLower > 0 || (cmpLower == 0 && inclusive) {
				// Value is within bucket, estimate fraction
				fraction := estimateFractionInBucket(value, bucket.LowerBound, bucket.UpperBound)
				matchingRows += int(float64(bucket.RowCount) * (1.0 - fraction))
			}
		}
		// else: value > bucket upper bound, bucket doesn't match
	}

	return float64(matchingRows) / float64(hist.TotalRows)
}

// estimateLessThan estimates selectivity for < or <= predicate
func (hist *EquiHeightHistogram) estimateLessThan(value interface{}, inclusive bool) float64 {
	matchingRows := 0

	for _, bucket := range hist.Buckets {
		cmpLower := compareValues(value, bucket.LowerBound)
		cmpUpper := compareValues(value, bucket.UpperBound)

		if cmpLower > 0 || (cmpLower == 0 && inclusive) {
			// Value >= bucket lower bound (with optional equality)
			if cmpUpper > 0 || (cmpUpper == 0 && !inclusive) {
				// Value > bucket upper bound, entire bucket matches
				matchingRows += bucket.RowCount
			} else if cmpUpper < 0 || (cmpUpper == 0 && inclusive) {
				// Value is within bucket, estimate fraction
				fraction := estimateFractionInBucket(value, bucket.LowerBound, bucket.UpperBound)
				matchingRows += int(float64(bucket.RowCount) * fraction)
			}
		}
		// else: value < bucket lower bound, bucket doesn't match
	}

	return float64(matchingRows) / float64(hist.TotalRows)
}

// EstimateRangeSelectivity estimates selectivity for range predicate (col BETWEEN low AND high)
func (hist *EquiHeightHistogram) EstimateRangeSelectivity(low, high interface{}) float64 {
	if compareValues(low, high) > 0 {
		return 0.0 // Invalid range
	}

	// Use: P(low <= col <= high) = P(col <= high) - P(col < low)
	upperSel := hist.estimateLessThan(high, true)
	lowerSel := hist.estimateLessThan(low, false)

	return math.Max(0.0, upperSel-lowerSel)
}

// String returns a human-readable representation
func (hist *EquiHeightHistogram) String() string {
	if len(hist.Buckets) == 0 {
		return "EquiHeightHistogram{empty}"
	}

	return fmt.Sprintf("EquiHeightHistogram{buckets=%d, total_rows=%d, bucket_size=%d}",
		len(hist.Buckets), hist.TotalRows, hist.BucketSize)
}

// estimateFractionInBucket estimates where a value falls within a bucket
// Returns 0.0 (at lower bound) to 1.0 (at upper bound)
func estimateFractionInBucket(value, lowerBound, upperBound interface{}) float64 {
	// For numeric types, use linear interpolation
	vFloat, vOk := toFloat64(value)
	lFloat, lOk := toFloat64(lowerBound)
	uFloat, uOk := toFloat64(upperBound)

	if vOk && lOk && uOk && uFloat > lFloat {
		fraction := (vFloat - lFloat) / (uFloat - lFloat)
		return math.Max(0.0, math.Min(1.0, fraction))
	}

	// For non-numeric types, assume middle of bucket
	return 0.5
}

// countDistinctValues counts unique values in a slice
func countDistinctValues(values []interface{}) int {
	seen := make(map[interface{}]struct{})
	for _, v := range values {
		// Use string representation as map key for complex types
		key := stringValue(v)
		seen[key] = struct{}{}
	}
	return len(seen)
}

// toFloat64 converts various numeric types to float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
