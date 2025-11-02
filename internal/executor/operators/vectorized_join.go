package operators

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
	"go.uber.org/zap"
)

// VectorizedJoin implements vectorized hash join using ArrowHashTable
type VectorizedJoin struct {
	joinType     string               // JOIN type: "INNER", "LEFT", "RIGHT"
	condition    optimizer.Expression // JOIN condition
	left         Operator             // Left child operator
	right        Operator             // Right child operator
	ctx          interface{}
	hashTable    *ArrowHashTable // Hash table for build side
	leftBatches  []*types.Batch  // Cached left (probe) side batches
	rightBatches []*types.Batch  // Cached right (build) side batches
	initialized  bool
	resultSent   bool
}

// NewVectorizedJoin creates a new vectorized join operator
func NewVectorizedJoin(joinType string, condition optimizer.Expression, left, right Operator, ctx interface{}) *VectorizedJoin {
	return &VectorizedJoin{
		joinType:     joinType,
		condition:    condition,
		left:         left,
		right:        right,
		ctx:          ctx,
		hashTable:    NewArrowHashTable(memory.NewGoAllocator()),
		leftBatches:  make([]*types.Batch, 0),
		rightBatches: make([]*types.Batch, 0),
		initialized:  false,
		resultSent:   false,
	}
}

// Init initializes the operator
func (vj *VectorizedJoin) Init(ctx interface{}) error {
	if err := vj.left.Init(ctx); err != nil {
		return err
	}
	return vj.right.Init(ctx)
}

// Next returns the next batch of joined data
func (vj *VectorizedJoin) Next() (*types.Batch, error) {
	if !vj.initialized {
		if err := vj.buildHashTable(); err != nil {
			return nil, fmt.Errorf("failed to build hash table: %w", err)
		}
		vj.initialized = true
	}

	// JOIN returns all results in one batch
	if vj.resultSent {
		return nil, nil
	}
	vj.resultSent = true

	return vj.probeAndBuildResult()
}

// buildHashTable constructs hash table from right (build) side
func (vj *VectorizedJoin) buildHashTable() error {
	logger.WithComponent("vectorized_join").Info("Building hash table",
		zap.String("join_type", vj.joinType))

	// Cache all right side batches
	for {
		batch, err := vj.right.Next()
		if err != nil {
			return err
		}
		if batch == nil {
			break
		}
		vj.rightBatches = append(vj.rightBatches, batch)
	}

	if len(vj.rightBatches) == 0 {
		logger.WithComponent("vectorized_join").Warn("Right side has no data")
		return nil
	}

	// Extract join key columns from condition
	_, rightKey, err := vj.extractJoinKeys()
	if err != nil {
		return err
	}

	// Find right key column index
	rightSchema := vj.rightBatches[0].Record().Schema()
	rightTableName := vj.getRightTableName()
	rightKeyIdx := -1

	logger.WithComponent("vectorized_join").Debug("Searching for right join key",
		zap.String("rightKey", rightKey),
		zap.String("rightTable", rightTableName),
		zap.Int("schema_field_count", len(rightSchema.Fields())))

	for i, field := range rightSchema.Fields() {
		logger.WithComponent("vectorized_join").Debug("Right schema field",
			zap.Int("index", i),
			zap.String("field_name", field.Name))

		if field.Name == rightKey || field.Name == rightTableName+"."+rightKey {
			rightKeyIdx = i
			logger.WithComponent("vectorized_join").Info("Found right join key",
				zap.String("field_name", field.Name),
				zap.Int("index", i))
			break
		}
	}

	if rightKeyIdx == -1 {
		return fmt.Errorf("right join key column '%s' not found in schema (table: %s)", rightKey, rightTableName)
	}

	// Insert all right batches into hash table
	totalRows := int64(0)
	for _, batch := range vj.rightBatches {
		err := vj.hashTable.Insert(batch.Record(), []int{rightKeyIdx})
		if err != nil {
			return fmt.Errorf("failed to insert batch into hash table: %w", err)
		}
		totalRows += batch.Record().NumRows()
	}

	stats := vj.hashTable.GetStatistics()
	logger.WithComponent("vectorized_join").Info("Hash table built successfully",
		zap.Int64("total_rows", totalRows),
		zap.String("stats", stats.String()))

	return nil
}

// probeAndBuildResult probes hash table with left side and builds result
func (vj *VectorizedJoin) probeAndBuildResult() (*types.Batch, error) {
	// Cache all left side batches
	for {
		batch, err := vj.left.Next()
		if err != nil {
			return nil, err
		}
		if batch == nil {
			break
		}
		vj.leftBatches = append(vj.leftBatches, batch)
	}

	if len(vj.leftBatches) == 0 {
		logger.WithComponent("vectorized_join").Warn("Left side has no data")
		return nil, nil
	}

	// Extract join keys
	leftKey, _, err := vj.extractJoinKeys()
	if err != nil {
		return nil, err
	}

	// Get schemas
	leftSchema := vj.leftBatches[0].Record().Schema()
	var rightSchema *arrow.Schema
	if len(vj.rightBatches) > 0 {
		rightSchema = vj.rightBatches[0].Record().Schema()
	}

	// Find left key column index
	leftTableName := vj.getLeftTableName()
	leftKeyIdx := -1

	logger.WithComponent("vectorized_join").Debug("Searching for left join key",
		zap.String("leftKey", leftKey),
		zap.String("leftTable", leftTableName),
		zap.Int("schema_field_count", len(leftSchema.Fields())))

	for i, field := range leftSchema.Fields() {
		logger.WithComponent("vectorized_join").Debug("Left schema field",
			zap.Int("index", i),
			zap.String("field_name", field.Name))

		if field.Name == leftKey || field.Name == leftTableName+"."+leftKey {
			leftKeyIdx = i
			logger.WithComponent("vectorized_join").Info("Found left join key",
				zap.String("field_name", field.Name),
				zap.Int("index", i))
			break
		}
	}

	if leftKeyIdx == -1 {
		return nil, fmt.Errorf("left join key column '%s' not found in schema (table: %s)", leftKey, leftTableName)
	}

	// Build output schema
	outputSchema := vj.mergeSchemas(leftSchema, rightSchema)

	// Create record builder
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, outputSchema)
	defer builder.Release()

	// Probe each left batch
	for _, leftBatch := range vj.leftBatches {
		leftRec := leftBatch.Record()

		// Probe hash table
		matches, err := vj.hashTable.Probe(leftRec, []int{leftKeyIdx})
		if err != nil {
			return nil, fmt.Errorf("failed to probe hash table: %w", err)
		}

		// Build match map: leftRowIdx -> []buildRowID
		matchMap := make(map[int32][]int32)
		for _, match := range matches {
			matchMap[match.ProbeRow] = append(matchMap[match.ProbeRow], match.BuildRow)
		}

		// Process each left row
		for leftRowIdx := int64(0); leftRowIdx < leftRec.NumRows(); leftRowIdx++ {
			buildRows := matchMap[int32(leftRowIdx)]

			if len(buildRows) > 0 {
				// Has matches - append all combinations
				for _, buildRowID := range buildRows {
					rightBatch, rightRowIdx := vj.hashTable.GetBuildBatch(buildRowID)
					if rightBatch == nil {
						continue
					}
					vj.appendJoinedRow(builder, leftRec, leftRowIdx, rightBatch, int64(rightRowIdx), leftSchema, rightSchema)
				}
			} else if vj.joinType == "LEFT" {
				// LEFT JOIN with no match - append with NULL right columns
				vj.appendLeftJoinRow(builder, leftRec, leftRowIdx, leftSchema, rightSchema)
			}
		}
	}

	// Build final result
	resultRecord := builder.NewRecord()
	if resultRecord.NumRows() == 0 {
		resultRecord.Release()
		return nil, nil
	}

	logger.WithComponent("vectorized_join").Info("Join completed",
		zap.Int64("result_rows", resultRecord.NumRows()))

	return types.NewBatch(resultRecord), nil
}

// extractJoinKeys extracts left and right join key column names from condition
func (vj *VectorizedJoin) extractJoinKeys() (string, string, error) {
	// Assuming condition is a binary expression: left.col = right.col
	if binExpr, ok := vj.condition.(*optimizer.BinaryExpression); ok && binExpr.Operator == "=" {
		leftCol, ok1 := binExpr.Left.(*optimizer.ColumnReference)
		rightCol, ok2 := binExpr.Right.(*optimizer.ColumnReference)

		if ok1 && ok2 {
			return leftCol.Column, rightCol.Column, nil
		}
	}

	return "", "", fmt.Errorf("unsupported join condition: %v", vj.condition)
}

// getLeftTableName extracts left table name from condition
func (vj *VectorizedJoin) getLeftTableName() string {
	if binExpr, ok := vj.condition.(*optimizer.BinaryExpression); ok {
		if leftCol, ok := binExpr.Left.(*optimizer.ColumnReference); ok {
			return leftCol.Table
		}
	}
	return ""
}

// getRightTableName extracts right table name from condition
func (vj *VectorizedJoin) getRightTableName() string {
	if binExpr, ok := vj.condition.(*optimizer.BinaryExpression); ok {
		if rightCol, ok := binExpr.Right.(*optimizer.ColumnReference); ok {
			return rightCol.Table
		}
	}
	return ""
}

// mergeSchemas merges left and right schemas
func (vj *VectorizedJoin) mergeSchemas(leftSchema, rightSchema *arrow.Schema) *arrow.Schema {
	var fields []arrow.Field

	// Add left fields
	for _, field := range leftSchema.Fields() {
		fields = append(fields, field)
	}

	// Add right fields
	if rightSchema != nil {
		for _, field := range rightSchema.Fields() {
			fields = append(fields, field)
		}
	}

	return arrow.NewSchema(fields, nil)
}

// appendJoinedRow appends a joined row to the builder
func (vj *VectorizedJoin) appendJoinedRow(
	builder *array.RecordBuilder,
	leftRec arrow.Record,
	leftRowIdx int64,
	rightRec arrow.Record,
	rightRowIdx int64,
	leftSchema, rightSchema *arrow.Schema,
) {
	fieldIdx := 0

	// Append left columns
	for colIdx := 0; colIdx < len(leftSchema.Fields()); colIdx++ {
		if colIdx < int(leftRec.NumCols()) {
			vj.appendValue(builder.Field(fieldIdx), leftRec.Column(colIdx), int(leftRowIdx))
		} else {
			builder.Field(fieldIdx).AppendNull()
		}
		fieldIdx++
	}

	// Append right columns
	if rightSchema != nil {
		for colIdx := 0; colIdx < len(rightSchema.Fields()); colIdx++ {
			if colIdx < int(rightRec.NumCols()) {
				vj.appendValue(builder.Field(fieldIdx), rightRec.Column(colIdx), int(rightRowIdx))
			} else {
				builder.Field(fieldIdx).AppendNull()
			}
			fieldIdx++
		}
	}
}

// appendLeftJoinRow appends left row with NULL right columns
func (vj *VectorizedJoin) appendLeftJoinRow(
	builder *array.RecordBuilder,
	leftRec arrow.Record,
	leftRowIdx int64,
	leftSchema, rightSchema *arrow.Schema,
) {
	fieldIdx := 0

	// Append left columns
	for colIdx := 0; colIdx < len(leftSchema.Fields()); colIdx++ {
		if colIdx < int(leftRec.NumCols()) {
			vj.appendValue(builder.Field(fieldIdx), leftRec.Column(colIdx), int(leftRowIdx))
		} else {
			builder.Field(fieldIdx).AppendNull()
		}
		fieldIdx++
	}

	// Append NULL for right columns
	if rightSchema != nil {
		for colIdx := 0; colIdx < len(rightSchema.Fields()); colIdx++ {
			builder.Field(fieldIdx).AppendNull()
			fieldIdx++
		}
	}
}

// appendValue appends a value from an array to a builder
func (vj *VectorizedJoin) appendValue(field array.Builder, arr arrow.Array, rowIdx int) {
	if arr.IsNull(rowIdx) {
		field.AppendNull()
		return
	}

	switch b := field.(type) {
	case *array.Int64Builder:
		if col, ok := arr.(*array.Int64); ok {
			b.Append(col.Value(rowIdx))
		} else {
			b.AppendNull()
		}
	case *array.Int32Builder:
		if col, ok := arr.(*array.Int32); ok {
			b.Append(col.Value(rowIdx))
		} else {
			b.AppendNull()
		}
	case *array.Float64Builder:
		if col, ok := arr.(*array.Float64); ok {
			b.Append(col.Value(rowIdx))
		} else {
			b.AppendNull()
		}
	case *array.Float32Builder:
		if col, ok := arr.(*array.Float32); ok {
			b.Append(col.Value(rowIdx))
		} else {
			b.AppendNull()
		}
	case *array.StringBuilder:
		if col, ok := arr.(*array.String); ok {
			b.Append(col.Value(rowIdx))
		} else {
			b.AppendNull()
		}
	case *array.BinaryBuilder:
		if col, ok := arr.(*array.Binary); ok {
			b.Append(col.Value(rowIdx))
		} else {
			b.AppendNull()
		}
	case *array.BooleanBuilder:
		if col, ok := arr.(*array.Boolean); ok {
			b.Append(col.Value(rowIdx))
		} else {
			b.AppendNull()
		}
	default:
		field.AppendNull()
	}
}

// Close closes the operator
func (vj *VectorizedJoin) Close() error {
	vj.hashTable.Release()

	if err := vj.left.Close(); err != nil {
		return err
	}
	return vj.right.Close()
}
