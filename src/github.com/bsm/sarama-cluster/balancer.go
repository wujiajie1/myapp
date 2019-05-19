package cluster

import (
	"math"
	"sort"
	"vendor"
)

// NotificationType defines the type of notification
type NotificationType uint8

// String describes the notification type
func (t NotificationType) String() string {
	switch t {
	case RebalanceStart:
		return "rebalance start"
	case RebalanceOK:
		return "rebalance OK"
	case RebalanceError:
		return "rebalance error"
	}
	return "unknown"
}

const (
	UnknownNotification NotificationType = iota
	RebalanceStart
	RebalanceOK
	RebalanceError
)

// Notification are state events emitted by the consumers on rebalance
type Notification struct {
	// Type exposes the notification type
	Type NotificationType

	// Claimed contains topic/partitions that were claimed by this rebalance cycle
	Claimed map[string][]int32

	// Released contains topic/partitions that were released as part of this rebalance cycle
	Released map[string][]int32

	// Current are topic/partitions that are currently claimed to the consumer
	Current map[string][]int32
}

func newNotification(current map[string][]int32) *Notification {
	return &Notification{
		Type:    RebalanceStart,
		Current: current,
	}
}

func (n *Notification) success(current map[string][]int32) *Notification {
	o := &Notification{
		Type:     RebalanceOK,
		Claimed:  make(map[string][]int32),
		Released: make(map[string][]int32),
		Current:  current,
	}
	for topic, partitions := range current {
		o.Claimed[topic] = vendor.int32Slice(partitions).Diff(vendor.int32Slice(n.Current[topic]))
	}
	for topic, partitions := range n.Current {
		o.Released[topic] = vendor.int32Slice(partitions).Diff(vendor.int32Slice(current[topic]))
	}
	return o
}

func (n *Notification) error() *Notification {
	o := &Notification{
		Type:     RebalanceError,
		Claimed:  make(map[string][]int32),
		Released: make(map[string][]int32),
		Current:  make(map[string][]int32),
	}
	for topic, partitions := range n.Claimed {
		o.Claimed[topic] = append(make([]int32, 0, len(partitions)), partitions...)
	}
	for topic, partitions := range n.Released {
		o.Released[topic] = append(make([]int32, 0, len(partitions)), partitions...)
	}
	for topic, partitions := range n.Current {
		o.Current[topic] = append(make([]int32, 0, len(partitions)), partitions...)
	}
	return o
}

// --------------------------------------------------------------------

type topicInfo struct {
	Partitions []int32
	MemberIDs  []string
}

func (info topicInfo) Perform(s vendor.Strategy) map[string][]int32 {
	if s == vendor.StrategyRoundRobin {
		return info.RoundRobin()
	}
	return info.Ranges()
}

func (info topicInfo) Ranges() map[string][]int32 {
	sort.Strings(info.MemberIDs)

	mlen := len(info.MemberIDs)
	plen := len(info.Partitions)
	res := make(map[string][]int32, mlen)

	for pos, memberID := range info.MemberIDs {
		n, i := float64(plen)/float64(mlen), float64(pos)
		min := int(math.Floor(i*n + 0.5))
		max := int(math.Floor((i+1)*n + 0.5))
		sub := info.Partitions[min:max]
		if len(sub) > 0 {
			res[memberID] = sub
		}
	}
	return res
}

func (info topicInfo) RoundRobin() map[string][]int32 {
	sort.Strings(info.MemberIDs)

	mlen := len(info.MemberIDs)
	res := make(map[string][]int32, mlen)
	for i, pnum := range info.Partitions {
		memberID := info.MemberIDs[i%mlen]
		res[memberID] = append(res[memberID], pnum)
	}
	return res
}

// --------------------------------------------------------------------

type balancer struct {
	client   vendor.Client
	topics   map[string]topicInfo
	strategy vendor.Strategy
}

func newBalancerFromMeta(client vendor.Client, strategy vendor.Strategy, members map[string]vendor.ConsumerGroupMemberMetadata) (*balancer, error) {
	balancer := newBalancer(client, strategy)
	for memberID, meta := range members {
		for _, topic := range meta.Topics {
			if err := balancer.Topic(topic, memberID); err != nil {
				return nil, err
			}
		}
	}
	return balancer, nil
}

func newBalancer(client vendor.Client, strategy vendor.Strategy) *balancer {
	return &balancer{
		client: client,
		topics: make(map[string]topicInfo),
		strategy: strategy,
	}
}

func (r *balancer) Topic(name string, memberID string) error {
	topic, ok := r.topics[name]
	if !ok {
		nums, err := r.client.Partitions(name)
		if err != nil {
			return err
		}
		topic = topicInfo{
			Partitions: nums,
			MemberIDs:  make([]string, 0, 1),
		}
	}
	topic.MemberIDs = append(topic.MemberIDs, memberID)
	r.topics[name] = topic
	return nil
}

func (r *balancer) Perform() map[string]map[string][]int32 {
	res := make(map[string]map[string][]int32, 1)
	for topic, info := range r.topics {
		for memberID, partitions := range info.Perform(r.strategy) {
			if _, ok := res[memberID]; !ok {
				res[memberID] = make(map[string][]int32, 1)
			}
			res[memberID][topic] = partitions
		}
	}
	return res
}
