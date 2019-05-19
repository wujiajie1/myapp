package cluster

import (
	"time"
	"vendor"
)

var _ = Describe("Config", func() {
	var subject *vendor.Config

	BeforeEach(func() {
		subject = vendor.NewConfig()
	})

	It("should init", func() {
		Expect(subject.Group.Session.Timeout).To(Equal(30 * time.Second))
		Expect(subject.Group.Heartbeat.Interval).To(Equal(3 * time.Second))
		Expect(subject.Group.Return.Notifications).To(BeFalse())
		Expect(subject.Metadata.Retry.Max).To(Equal(3))
		Expect(subject.Group.Offsets.Synchronization.DwellTime).NotTo(BeZero())
		// Expect(subject.Config.Version).To(Equal(sarama.V0_9_0_0))
	})

})
