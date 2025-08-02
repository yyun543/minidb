package types

import (
	"fmt"
	"hash/crc32"
)

// PartitionInfo 分区信息，为分布式扩展做准备
type PartitionInfo struct {
	TableName   string            // 表名
	PartitionID uint32            // 分区ID
	ShardCount  uint32            // 总分片数
	KeyColumns  []string          // 分区键列名
	Metadata    map[string]string // 扩展元数据
}

// PartitionType 分区类型
type PartitionType uint8

const (
	HashPartition  PartitionType = iota // 哈希分区（适合均匀分布）
	RangePartition                      // 范围分区（适合有序数据）
	ListPartition                       // 列表分区（适合枚举值）
)

// PartitionStrategy 分区策略
type PartitionStrategy struct {
	Type       PartitionType    // 分区类型
	Columns    []string         // 分区列
	ShardCount uint32           // 分片数量
	Ranges     []PartitionRange // 范围分区的范围定义
	Lists      []PartitionList  // 列表分区的值定义
}

// PartitionRange 范围分区定义
type PartitionRange struct {
	Start interface{} // 开始值
	End   interface{} // 结束值
	Shard uint32      // 分片ID
}

// PartitionList 列表分区定义
type PartitionList struct {
	Values []interface{} // 值列表
	Shard  uint32        // 分片ID
}

// PartitionManager 分区管理器
type PartitionManager struct {
	strategy *PartitionStrategy
}

// NewPartitionManager 创建分区管理器
func NewPartitionManager(strategy *PartitionStrategy) *PartitionManager {
	return &PartitionManager{strategy: strategy}
}

// GetPartitionID 根据分区键值计算分区ID
func (pm *PartitionManager) GetPartitionID(keyValues []interface{}) (uint32, error) {
	if len(keyValues) != len(pm.strategy.Columns) {
		return 0, fmt.Errorf("key values count mismatch: expected %d, got %d",
			len(pm.strategy.Columns), len(keyValues))
	}

	switch pm.strategy.Type {
	case HashPartition:
		return pm.hashPartition(keyValues), nil
	case RangePartition:
		return pm.rangePartition(keyValues)
	case ListPartition:
		return pm.listPartition(keyValues)
	default:
		return 0, fmt.Errorf("unsupported partition type: %v", pm.strategy.Type)
	}
}

// hashPartition 哈希分区计算
func (pm *PartitionManager) hashPartition(keyValues []interface{}) uint32 {
	// 使用CRC32作为哈希函数，确保一致性
	hasher := crc32.NewIEEE()

	for _, value := range keyValues {
		var data []byte
		switch v := value.(type) {
		case string:
			data = []byte(v)
		case int64:
			data = []byte(fmt.Sprintf("%d", v))
		case float64:
			data = []byte(fmt.Sprintf("%f", v))
		case bool:
			if v {
				data = []byte("true")
			} else {
				data = []byte("false")
			}
		default:
			data = []byte(fmt.Sprintf("%v", v))
		}
		hasher.Write(data)
	}

	return hasher.Sum32() % pm.strategy.ShardCount
}

// rangePartition 范围分区计算
func (pm *PartitionManager) rangePartition(keyValues []interface{}) (uint32, error) {
	// 简化实现：只支持单列范围分区
	if len(keyValues) != 1 {
		return 0, fmt.Errorf("range partition only supports single column")
	}

	value := keyValues[0]
	for _, r := range pm.strategy.Ranges {
		if pm.compareValue(value, r.Start) >= 0 && pm.compareValue(value, r.End) < 0 {
			return r.Shard, nil
		}
	}

	return 0, fmt.Errorf("value %v does not fall in any range", value)
}

// listPartition 列表分区计算
func (pm *PartitionManager) listPartition(keyValues []interface{}) (uint32, error) {
	// 简化实现：只支持单列列表分区
	if len(keyValues) != 1 {
		return 0, fmt.Errorf("list partition only supports single column")
	}

	value := keyValues[0]
	for _, l := range pm.strategy.Lists {
		for _, listValue := range l.Values {
			if pm.compareValue(value, listValue) == 0 {
				return l.Shard, nil
			}
		}
	}

	return 0, fmt.Errorf("value %v not found in any list", value)
}

// compareValue 比较两个值（简化实现）
func (pm *PartitionManager) compareValue(a, b interface{}) int {
	// 简化的值比较，实际应用中需要更严格的类型处理
	switch va := a.(type) {
	case int64:
		if vb, ok := b.(int64); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			} else {
				return 0
			}
		}
	case string:
		if vb, ok := b.(string); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			} else {
				return 0
			}
		}
	}
	return 0
}

// ShardInfo 分片信息（为分布式做准备）
type ShardInfo struct {
	ShardID  uint32            // 分片ID
	NodeID   string            // 节点ID（分布式环境中的节点标识）
	Replicas []string          // 副本节点列表
	Status   ShardStatus       // 分片状态
	Metadata map[string]string // 扩展元数据
}

// ShardStatus 分片状态
type ShardStatus uint8

const (
	ShardActive      ShardStatus = iota // 活跃状态
	ShardRebalancing                    // 重新平衡中
	ShardOffline                        // 离线状态
	ShardReadOnly                       // 只读状态
)

// String 返回分片状态的字符串表示
func (s ShardStatus) String() string {
	switch s {
	case ShardActive:
		return "ACTIVE"
	case ShardRebalancing:
		return "REBALANCING"
	case ShardOffline:
		return "OFFLINE"
	case ShardReadOnly:
		return "READONLY"
	default:
		return "UNKNOWN"
	}
}

// ReplicationInfo 副本信息
type ReplicationInfo struct {
	Factor    int      // 副本因子
	Strategy  string   // 副本策略（如 "SimpleStrategy", "NetworkTopologyStrategy"）
	Placement []string // 副本放置策略
}
