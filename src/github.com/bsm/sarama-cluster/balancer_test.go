package cluster

import (
	"vendor"
)

var _ = Describe("Notification", func() {

	It("should init and convert", func() {
		n := vendor.newNotification(map[string][]int32{
			"a": {1, 2, 3},
			"b": {4, 5},
			"c": {1, 2},
		})
		Expect(n).To(Equal(&vendor.Notification{
			Type:    vendor.RebalanceStart,
			Current: map[string][]int32{"a": {1, 2, 3}, "b": {4, 5}, "c": {1, 2}},
		}))

		o := n.success(map[string][]int32{
			"a": {3, 4},
			"b": {1, 2, 3, 4},
			"d": {3, 4},
		})
		Expect(o).To(Equal(&vendor.Notification{
			Type:     vendor.RebalanceOK,
			Claimed:  map[string][]int32{"a": {4}, "b": {1, 2, 3}, "d": {3, 4}},
			Released: map[string][]int32{"a": {1, 2}, "b": {5}, "c": {1, 2}},
			Current:  map[string][]int32{"a": {3, 4}, "b": {1, 2, 3, 4}, "d": {3, 4}},
		}))
	})

	It("should copy on error", func() {
		n := vendor.newNotification(map[string][]int32{
			"a": {1, 2, 3},
			"b": {4, 5},
			"c": {1, 2},
		})
		o := n.error()

		Expect(n).To(Equal(&vendor.Notification{
			Type:    vendor.RebalanceStart,
			Current: map[string][]int32{"a": {1, 2, 3}, "b": {4, 5}, "c": {1, 2}},
		}))

		Expect(o).To(Equal(&vendor.Notification{
			Type:     vendor.RebalanceError,
			Claimed:  map[string][]int32{},
			Released: map[string][]int32{},
			Current:  map[string][]int32{"a": {1, 2, 3}, "b": {4, 5}, "c": {1, 2}},
		}))
	})

})

var _ = Describe("balancer", func() {
	var subject *vendor.balancer

	BeforeEach(func() {
		client := &vendor.mockClient{
			topics: map[string][]int32{
				"one":   {0, 1, 2, 3},
				"two":   {0, 1, 2},
				"three": {0, 1},
			},
		}

		var err error
		subject, err = vendor.newBalancerFromMeta(client, vendor.StrategyRange, map[string]vendor.ConsumerGroupMemberMetadata{
			"b": {Topics: []string{"three", "one"}},
			"a": {Topics: []string{"one", "two"}},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should parse from meta data", func() {
		Expect(subject.topics).To(HaveLen(3))
	})

	It("should perform", func() {
		Expect(subject.Perform()).To(Equal(map[string]map[string][]int32{
			"a": {"one": {0, 1}, "two": {0, 1, 2}},
			"b": {"one": {2, 3}, "three": {0, 1}},
		}))

		subject.strategy = vendor.StrategyRoundRobin
		Expect(subject.Perform()).To(Equal(map[string]map[string][]int32{
			"a": {"one": {0, 2}, "two": {0, 1, 2}},
			"b": {"one": {1, 3}, "three": {0, 1}},
		}))
	})

})

var _ = Describe("topicInfo", func() {

	DescribeTable("Ranges",
		func(memberIDs []string, partitions []int32, expected map[string][]int32) {
			info := vendor.topicInfo{MemberIDs: memberIDs, Partitions: partitions}
			Expect(info.Ranges()).To(Equal(expected))
		},

		Entry("three members, three partitions", []string{"M1", "M2", "M3"}, []int32{0, 1, 2}, map[string][]int32{
			"M1": {0}, "M2": {1}, "M3": {2},
		}),
		Entry("member ID order", []string{"M3", "M1", "M2"}, []int32{0, 1, 2}, map[string][]int32{
			"M1": {0}, "M2": {1}, "M3": {2},
		}),
		Entry("more members than partitions", []string{"M1", "M2", "M3"}, []int32{0, 1}, map[string][]int32{
			"M1": {0}, "M3": {1},
		}),
		Entry("far more members than partitions", []string{"M1", "M2", "M3"}, []int32{0}, map[string][]int32{
			"M2": {0},
		}),
		Entry("fewer members than partitions", []string{"M1", "M2", "M3"}, []int32{0, 1, 2, 3}, map[string][]int32{
			"M1": {0}, "M2": {1, 2}, "M3": {3},
		}),
		Entry("uneven members/partitions ratio", []string{"M1", "M2", "M3"}, []int32{0, 2, 4, 6, 8}, map[string][]int32{
			"M1": {0, 2}, "M2": {4}, "M3": {6, 8},
		}),
	)

	DescribeTable("RoundRobin",
		func(memberIDs []string, partitions []int32, expected map[string][]int32) {
			info := vendor.topicInfo{MemberIDs: memberIDs, Partitions: partitions}
			Expect(info.RoundRobin()).To(Equal(expected))
		},

		Entry("three members, three partitions", []string{"M1", "M2", "M3"}, []int32{0, 1, 2}, map[string][]int32{
			"M1": {0}, "M2": {1}, "M3": {2},
		}),
		Entry("member ID order", []string{"M3", "M1", "M2"}, []int32{0, 1, 2}, map[string][]int32{
			"M1": {0}, "M2": {1}, "M3": {2},
		}),
		Entry("more members than partitions", []string{"M1", "M2", "M3"}, []int32{0, 1}, map[string][]int32{
			"M1": {0}, "M2": {1},
		}),
		Entry("far more members than partitions", []string{"M1", "M2", "M3"}, []int32{0}, map[string][]int32{
			"M1": {0},
		}),
		Entry("fewer members than partitions", []string{"M1", "M2", "M3"}, []int32{0, 1, 2, 3}, map[string][]int32{
			"M1": {0, 3}, "M2": {1}, "M3": {2},
		}),
		Entry("uneven members/partitions ratio", []string{"M1", "M2", "M3"}, []int32{0, 2, 4, 6, 8}, map[string][]int32{
			"M1": {0, 6}, "M2": {2, 8}, "M3": {4},
		}),
	)

})
