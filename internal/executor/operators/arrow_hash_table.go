package operators

import (
	"fmt"
	"hash/fnv"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// ArrowHashTable implements a hash table for vectorized JOIN operations
// Uses Apache Arrow columnar format for efficient memory access
type ArrowHashTable struct {
	// Hash index: hash(key) -> []rowID
	// rowID encoding: (batchIndex << 24) | rowIndex
	// Supports up to 256 batches with up to 16M rows each
	index map[uint64][]int32

	// Build-side data batches
	buildBatches []arrow.Record

	// Join key column indices in build side
	buildKeyIndices []int

	// Memory allocator
	allocator memory.Allocator

	// Statistics
	totalRows    int64
	distinctKeys int64
	collisions   int64
}

// NewArrowHashTable creates a new hash table
func NewArrowHashTable(allocator memory.Allocator) *ArrowHashTable {
	return &ArrowHashTable{
		index:        make(map[uint64][]int32),
		buildBatches: make([]arrow.Record, 0),
		allocator:    allocator,
	}
}

// Insert adds a batch to the hash table (Build phase)
func (ht *ArrowHashTable) Insert(batch arrow.Record, keyIndices []int) error {
	if len(keyIndices) == 0 {
		return fmt.Errorf("no join keys specified")
	}

	// Store build key indices (first batch only)
	if ht.buildKeyIndices == nil {
		ht.buildKeyIndices = keyIndices
	}

	// Validate key indices
	for _, idx := range keyIndices {
		if idx < 0 || idx >= int(batch.NumCols()) {
			return fmt.Errorf("invalid key index %d (batch has %d columns)", idx, batch.NumCols())
		}
	}

	// Extract join key columns
	keyColumns := make([]arrow.Array, len(keyIndices))
	for i, idx := range keyIndices {
		keyColumns[i] = batch.Column(idx)
	}

	// Hash each row and build index
	numRows := int(batch.NumRows())
	batchIndex := len(ht.buildBatches)

	if batchIndex >= 256 {
		return fmt.Errorf("too many batches (max 256)")
	}

	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		// Skip NULL join keys (NULL doesn't match anything in standard SQL)
		if ht.hasNullKey(keyColumns, rowIdx) {
			continue
		}

		// Compute composite key hash
		hash := ht.computeHash(keyColumns, rowIdx)

		// Debug: log inserted key value
		keyStr := ""
		if len(keyColumns) > 0 {
			keyStr = keyColumns[0].ValueStr(rowIdx)
		}

		// Store row reference: (batchIndex << 24) | rowIdx
		rowID := int32((batchIndex << 24) | rowIdx)

		ht.index[hash] = append(ht.index[hash], rowID)
		ht.totalRows++

		logger.WithComponent("arrow_hash_table").Debug("Inserted key into hash table",
			zap.Int("batch", batchIndex),
			zap.Int("row", rowIdx),
			zap.Uint64("hash", hash),
			zap.String("key_value", keyStr))
	}

	// Store batch (retain to keep memory alive)
	batch.Retain()
	ht.buildBatches = append(ht.buildBatches, batch)
	ht.distinctKeys = int64(len(ht.index))

	logger.WithComponent("arrow_hash_table").Debug("Inserted batch into hash table",
		zap.Int("batch_index", batchIndex),
		zap.Int("rows", numRows),
		zap.Int64("total_rows", ht.totalRows),
		zap.Int64("distinct_keys", ht.distinctKeys))

	return nil
}

// Probe searches for matches in the hash table (Probe phase)
func (ht *ArrowHashTable) Probe(batch arrow.Record, keyIndices []int) ([]JoinMatch, error) {
	if len(keyIndices) == 0 {
		return nil, fmt.Errorf("no join keys specified")
	}

	// Validate key indices
	for _, idx := range keyIndices {
		if idx < 0 || idx >= int(batch.NumCols()) {
			return nil, fmt.Errorf("invalid key index %d (batch has %d columns)", idx, batch.NumCols())
		}
	}

	keyColumns := make([]arrow.Array, len(keyIndices))
	for i, idx := range keyIndices {
		keyColumns[i] = batch.Column(idx)
	}

	matches := make([]JoinMatch, 0)
	numRows := int(batch.NumRows())

	logger.WithComponent("arrow_hash_table").Debug("Starting probe phase",
		zap.Int("probe_rows", numRows),
		zap.Int("index_size", len(ht.index)))

	for probeRowIdx := 0; probeRowIdx < numRows; probeRowIdx++ {
		// Skip NULL join keys
		if ht.hasNullKey(keyColumns, probeRowIdx) {
			logger.WithComponent("arrow_hash_table").Debug("Skipping NULL key", zap.Int("row", probeRowIdx))
			continue
		}

		hash := ht.computeHash(keyColumns, probeRowIdx)

		// Debug: log probe key value
		probeKeyStr := ""
		if len(keyColumns) > 0 {
			probeKeyStr = keyColumns[0].ValueStr(probeRowIdx)
		}

		logger.WithComponent("arrow_hash_table").Debug("Probing row",
			zap.Int("row", probeRowIdx),
			zap.Uint64("hash", hash),
			zap.String("key_value", probeKeyStr))

		// Look up in hash table
		if buildRowIDs, found := ht.index[hash]; found {
			logger.WithComponent("arrow_hash_table").Debug("Hash found in index",
				zap.Int("probe_row", probeRowIdx),
				zap.Int("candidate_count", len(buildRowIDs)))

			// For each potential match, verify full key equality (handle hash collisions)
			for _, buildRowID := range buildRowIDs {
				batchIdx := int(buildRowID >> 24)
				rowIdx := int(buildRowID & 0xFFFFFF)

				if ht.keysEqual(keyColumns, probeRowIdx, batchIdx, rowIdx, keyIndices) {
					matches = append(matches, JoinMatch{
						ProbeRow: int32(probeRowIdx),
						BuildRow: buildRowID,
					})
					logger.WithComponent("arrow_hash_table").Debug("Match found!",
						zap.Int("probe_row", probeRowIdx),
						zap.Int("build_batch", batchIdx),
						zap.Int("build_row", rowIdx))
				} else {
					ht.collisions++
					logger.WithComponent("arrow_hash_table").Debug("Hash collision - keys not equal",
						zap.Int("probe_row", probeRowIdx))
				}
			}
		} else {
			logger.WithComponent("arrow_hash_table").Debug("Hash NOT found in index",
				zap.Int("probe_row", probeRowIdx),
				zap.Uint64("hash", hash))
		}
	}

	logger.WithComponent("arrow_hash_table").Info("Probe phase completed",
		zap.Int("matches_found", len(matches)))

	return matches, nil
}

// hasNullKey checks if any join key is NULL
func (ht *ArrowHashTable) hasNullKey(keyColumns []arrow.Array, rowIdx int) bool {
	for _, col := range keyColumns {
		if col.IsNull(rowIdx) {
			return true
		}
	}
	return false
}

// computeHash computes a hash value for a row's join keys
func (ht *ArrowHashTable) computeHash(keyColumns []arrow.Array, rowIdx int) uint64 {
	h := fnv.New64a()

	for _, col := range keyColumns {
		if col.IsNull(rowIdx) {
			// NULL values: use special marker
			h.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
			continue
		}

		switch arr := col.(type) {
		case *array.Int64:
			val := arr.Value(rowIdx)
			h.Write([]byte{
				byte(val), byte(val >> 8), byte(val >> 16), byte(val >> 24),
				byte(val >> 32), byte(val >> 40), byte(val >> 48), byte(val >> 56),
			})

		case *array.Int32:
			val := arr.Value(rowIdx)
			h.Write([]byte{
				byte(val), byte(val >> 8), byte(val >> 16), byte(val >> 24),
			})

		case *array.String:
			val := arr.Value(rowIdx)
			h.Write([]byte(val))

		case *array.Binary:
			val := arr.Value(rowIdx)
			h.Write(val)

		case *array.Float64:
			val := arr.Value(rowIdx)
			bits := uint64(val)
			h.Write([]byte{
				byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
				byte(bits >> 32), byte(bits >> 40), byte(bits >> 48), byte(bits >> 56),
			})

		case *array.Float32:
			val := arr.Value(rowIdx)
			bits := uint32(val)
			h.Write([]byte{
				byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
			})

		default:
			// Fallback: use string representation
			h.Write([]byte(col.ValueStr(rowIdx)))
		}
	}

	return h.Sum64()
}

// keysEqual checks if two rows have equal join keys
func (ht *ArrowHashTable) keysEqual(
	probeKeys []arrow.Array,
	probeRowIdx int,
	buildBatchIdx int,
	buildRowIdx int,
	keyIndices []int, // Not used - kept for API compatibility
) bool {
	if buildBatchIdx >= len(ht.buildBatches) {
		return false
	}

	buildBatch := ht.buildBatches[buildBatchIdx]

	// Use stored buildKeyIndices instead of the parameter
	for i, keyIdx := range ht.buildKeyIndices {
		if keyIdx >= int(buildBatch.NumCols()) {
			return false
		}

		probeCol := probeKeys[i]
		buildCol := buildBatch.Column(keyIdx)

		// Handle NULLs
		probeNull := probeCol.IsNull(probeRowIdx)
		buildNull := buildCol.IsNull(buildRowIdx)

		if probeNull != buildNull {
			return false // One is NULL, other isn't
		}
		if probeNull && buildNull {
			continue // Both NULL, considered equal for JOIN (though we skip NULLs earlier)
		}

		// Compare values
		if !ht.valuesEqual(probeCol, probeRowIdx, buildCol, buildRowIdx) {
			return false
		}
	}

	return true
}

// valuesEqual compares two non-NULL values
func (ht *ArrowHashTable) valuesEqual(
	col1 arrow.Array, idx1 int,
	col2 arrow.Array, idx2 int,
) bool {
	// Type-specific comparison
	switch arr1 := col1.(type) {
	case *array.Int64:
		arr2, ok := col2.(*array.Int64)
		if !ok {
			return false
		}
		return arr1.Value(idx1) == arr2.Value(idx2)

	case *array.Int32:
		arr2, ok := col2.(*array.Int32)
		if !ok {
			return false
		}
		return arr1.Value(idx1) == arr2.Value(idx2)

	case *array.String:
		arr2, ok := col2.(*array.String)
		if !ok {
			return false
		}
		return arr1.Value(idx1) == arr2.Value(idx2)

	case *array.Binary:
		arr2, ok := col2.(*array.Binary)
		if !ok {
			return false
		}
		val1 := arr1.Value(idx1)
		val2 := arr2.Value(idx2)
		if len(val1) != len(val2) {
			return false
		}
		for i := range val1 {
			if val1[i] != val2[i] {
				return false
			}
		}
		return true

	case *array.Float64:
		arr2, ok := col2.(*array.Float64)
		if !ok {
			return false
		}
		return arr1.Value(idx1) == arr2.Value(idx2)

	case *array.Float32:
		arr2, ok := col2.(*array.Float32)
		if !ok {
			return false
		}
		return arr1.Value(idx1) == arr2.Value(idx2)

	default:
		// Fallback: string comparison
		return col1.ValueStr(idx1) == col2.ValueStr(idx2)
	}
}

// GetBuildBatch returns the build batch at the given index
func (ht *ArrowHashTable) GetBuildBatch(rowID int32) (arrow.Record, int) {
	batchIdx := int(rowID >> 24)
	rowIdx := int(rowID & 0xFFFFFF)

	if batchIdx >= len(ht.buildBatches) {
		return nil, -1
	}

	return ht.buildBatches[batchIdx], rowIdx
}

// Release frees resources
func (ht *ArrowHashTable) Release() {
	for _, batch := range ht.buildBatches {
		batch.Release()
	}
	ht.buildBatches = nil
	ht.index = nil

	logger.WithComponent("arrow_hash_table").Debug("Released hash table resources",
		zap.Int64("total_rows_processed", ht.totalRows),
		zap.Int64("collisions", ht.collisions))
}

// JoinMatch represents a match between probe and build rows
type JoinMatch struct {
	ProbeRow int32 // Row index in current probe batch
	BuildRow int32 // Encoded: (batchIndex << 24) | rowIndex
}

// GetStatistics returns hash table statistics
func (ht *ArrowHashTable) GetStatistics() *HashTableStats {
	loadFactor := 0.0
	if len(ht.index) > 0 {
		loadFactor = float64(ht.totalRows) / float64(len(ht.index))
	}

	return &HashTableStats{
		TotalRows:    ht.totalRows,
		DistinctKeys: ht.distinctKeys,
		Collisions:   ht.collisions,
		LoadFactor:   loadFactor,
		BucketCount:  int64(len(ht.index)),
	}
}

// HashTableStats contains statistics about the hash table
type HashTableStats struct {
	TotalRows    int64   // Total rows inserted
	DistinctKeys int64   // Number of distinct keys
	Collisions   int64   // Number of hash collisions detected
	LoadFactor   float64 // Average rows per bucket
	BucketCount  int64   // Number of hash buckets
}

// String returns a string representation of the statistics
func (s *HashTableStats) String() string {
	return fmt.Sprintf("HashTable{rows=%d, keys=%d, buckets=%d, load=%.2f, collisions=%d}",
		s.TotalRows, s.DistinctKeys, s.BucketCount, s.LoadFactor, s.Collisions)
}
