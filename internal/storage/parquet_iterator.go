package storage

import (
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/parquet"
)

// ParquetIterator Parquet 文件迭代器
type ParquetIterator struct {
	files   []delta.FileInfo
	filters []Filter
	current int
	record  arrow.Record
	err     error
}

// NewParquetIterator 创建 Parquet 迭代器
func NewParquetIterator(files []delta.FileInfo, filters []Filter) (*ParquetIterator, error) {
	return &ParquetIterator{
		files:   files,
		filters: filters,
		current: -1,
	}, nil
}

// Next 移动到下一条记录
func (pi *ParquetIterator) Next() bool {
	// 释放上一条记录
	if pi.record != nil {
		pi.record.Release()
		pi.record = nil
	}

	// 移动到下一个文件
	pi.current++

	if pi.current >= len(pi.files) {
		return false
	}

	// 读取当前文件
	file := pi.files[pi.current]

	// 转换 Filter 类型
	parquetFilters := make([]parquet.Filter, len(pi.filters))
	for i, f := range pi.filters {
		parquetFilters[i] = parquet.Filter{
			Column:   f.Column,
			Operator: f.Operator,
			Value:    f.Value,
			Values:   f.Values, // 支持 IN 操作符的多个值
		}
	}

	record, err := parquet.ReadParquetFile(file.Path, parquetFilters)
	if err != nil {
		pi.err = err
		return false
	}

	pi.record = record
	return true
}

// Record 获取当前记录
func (pi *ParquetIterator) Record() arrow.Record {
	return pi.record
}

// Err 获取错误
func (pi *ParquetIterator) Err() error {
	return pi.err
}

// Close 关闭迭代器
func (pi *ParquetIterator) Close() error {
	if pi.record != nil {
		pi.record.Release()
		pi.record = nil
	}
	return nil
}
