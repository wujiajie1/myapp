package cluster

import (
	"sync"
	"vendor"
)

// OffsetStash allows to accumulate offsets and
// mark them as processed in a bulk
type OffsetStash struct {
	offsets map[vendor.topicPartition]vendor.offsetInfo
	mu      sync.Mutex
}

// NewOffsetStash inits a blank stash
func NewOffsetStash() *OffsetStash {
	return &OffsetStash{offsets: make(map[vendor.topicPartition]vendor.offsetInfo)}
}

// MarkOffset stashes the provided message offset
func (s *OffsetStash) MarkOffset(msg *vendor.ConsumerMessage, metadata string) {
	s.MarkPartitionOffset(msg.Topic, msg.Partition, msg.Offset, metadata)
}

// MarkPartitionOffset stashes the offset for the provided topic/partition combination
func (s *OffsetStash) MarkPartitionOffset(topic string, partition int32, offset int64, metadata string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := vendor.topicPartition{Topic: topic, Partition: partition}
	if info := s.offsets[key]; offset >= info.Offset {
		info.Offset = offset
		info.Metadata = metadata
		s.offsets[key] = info
	}
}

// ResetPartitionOffset stashes the offset for the provided topic/partition combination.
// Difference between ResetPartitionOffset and MarkPartitionOffset is that, ResetPartitionOffset supports earlier offsets
func (s *OffsetStash) ResetPartitionOffset(topic string, partition int32, offset int64, metadata string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := vendor.topicPartition{Topic: topic, Partition: partition}
	if info := s.offsets[key]; offset <= info.Offset {
		info.Offset = offset
		info.Metadata = metadata
		s.offsets[key] = info
	}
}

// ResetOffset stashes the provided message offset
// See ResetPartitionOffset for explanation
func (s *OffsetStash) ResetOffset(msg *vendor.ConsumerMessage, metadata string) {
	s.ResetPartitionOffset(msg.Topic, msg.Partition, msg.Offset, metadata)
}

// Offsets returns the latest stashed offsets by topic-partition
func (s *OffsetStash) Offsets() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make(map[string]int64, len(s.offsets))
	for tp, info := range s.offsets {
		res[tp.String()] = info.Offset
	}
	return res
}
