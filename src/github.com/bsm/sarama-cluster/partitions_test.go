package cluster

import (
	"vendor"
)

var _ = Describe("partitionConsumer", func() {
	var subject *vendor.partitionConsumer

	BeforeEach(func() {
		var err error
		subject, err = vendor.newPartitionConsumer(&vendor.mockConsumer{}, "topic", 0, vendor.offsetInfo{2000, "m3ta"}, vendor.OffsetOldest)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		close(subject.dead)
		Expect(subject.Close()).NotTo(HaveOccurred())
	})

	It("should set state", func() {
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info: vendor.offsetInfo{2000, "m3ta"},
		}))
	})

	It("should recover from default offset if requested offset is out of bounds", func() {
		pc, err := vendor.newPartitionConsumer(&vendor.mockConsumer{}, "topic", 0, vendor.offsetInfo{200, "m3ta"}, vendor.OffsetOldest)
		Expect(err).NotTo(HaveOccurred())
		defer pc.Close()
		close(pc.dead)

		state := pc.getState()
		Expect(state.Info.Offset).To(Equal(int64(-1)))
		Expect(state.Info.Metadata).To(Equal("m3ta"))
	})

	It("should update state", func() {
		subject.MarkOffset(2001, "met@") // should set state
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info:  vendor.offsetInfo{2002, "met@"},
			Dirty: true,
		}))

		subject.markCommitted(2002) // should reset dirty status
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info: vendor.offsetInfo{2002, "met@"},
		}))

		subject.MarkOffset(2001, "me7a") // should not update state
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info: vendor.offsetInfo{2002, "met@"},
		}))

		subject.MarkOffset(2002, "me7a") // should bump state
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info:  vendor.offsetInfo{2003, "me7a"},
			Dirty: true,
		}))

		// After committing a later offset, try rewinding back to earlier offset with new metadata.
		subject.ResetOffset(2001, "met@")
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info:  vendor.offsetInfo{2002, "met@"},
			Dirty: true,
		}))

		subject.markCommitted(2002) // should not unset state
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info: vendor.offsetInfo{2002, "met@"},
		}))

		subject.MarkOffset(2002, "me7a") // should bump state
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info:  vendor.offsetInfo{2003, "me7a"},
			Dirty: true,
		}))

		subject.markCommitted(2003)
		Expect(subject.getState()).To(Equal(vendor.partitionState{
			Info: vendor.offsetInfo{2003, "me7a"},
		}))
	})

})

var _ = Describe("partitionMap", func() {
	var subject *vendor.partitionMap

	BeforeEach(func() {
		subject = vendor.newPartitionMap()
	})

	It("should fetch/store", func() {
		Expect(subject.Fetch("topic", 0)).To(BeNil())

		pc, err := vendor.newPartitionConsumer(&vendor.mockConsumer{}, "topic", 0, vendor.offsetInfo{2000, "m3ta"}, vendor.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())

		subject.Store("topic", 0, pc)
		Expect(subject.Fetch("topic", 0)).To(Equal(pc))
		Expect(subject.Fetch("topic", 1)).To(BeNil())
		Expect(subject.Fetch("other", 0)).To(BeNil())
	})

	It("should return info", func() {
		pc0, err := vendor.newPartitionConsumer(&vendor.mockConsumer{}, "topic", 0, vendor.offsetInfo{2000, "m3ta"}, vendor.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())
		pc1, err := vendor.newPartitionConsumer(&vendor.mockConsumer{}, "topic", 1, vendor.offsetInfo{2000, "m3ta"}, vendor.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())
		subject.Store("topic", 0, pc0)
		subject.Store("topic", 1, pc1)

		info := subject.Info()
		Expect(info).To(HaveLen(1))
		Expect(info).To(HaveKeyWithValue("topic", []int32{0, 1}))
	})

	It("should create snapshots", func() {
		pc0, err := vendor.newPartitionConsumer(&vendor.mockConsumer{}, "topic", 0, vendor.offsetInfo{2000, "m3ta"}, vendor.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())
		pc1, err := vendor.newPartitionConsumer(&vendor.mockConsumer{}, "topic", 1, vendor.offsetInfo{2000, "m3ta"}, vendor.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())

		subject.Store("topic", 0, pc0)
		subject.Store("topic", 1, pc1)
		subject.Fetch("topic", 1).MarkOffset(2000, "met@")

		Expect(subject.Snapshot()).To(Equal(map[vendor.topicPartition]vendor.partitionState{
			{"topic", 0}: {Info: vendor.offsetInfo{2000, "m3ta"}, Dirty: false},
			{"topic", 1}: {Info: vendor.offsetInfo{2001, "met@"}, Dirty: true},
		}))
	})

})
