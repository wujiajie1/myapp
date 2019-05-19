package cluster

import (
	"vendor"
)

var _ = Describe("Client", func() {
	var subject *vendor.Client

	BeforeEach(func() {
		var err error
		subject, err = vendor.NewClient(vendor.testKafkaAddrs, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not allow to share clients across multiple consumers", func() {
		c1, err := vendor.NewConsumerFromClient(subject, vendor.testGroup, vendor.testTopics)
		Expect(err).NotTo(HaveOccurred())
		defer c1.Close()

		_, err = vendor.NewConsumerFromClient(subject, vendor.testGroup, vendor.testTopics)
		Expect(err).To(MatchError("cluster: client is already used by another consumer"))

		Expect(c1.Close()).To(Succeed())
		c2, err := vendor.NewConsumerFromClient(subject, vendor.testGroup, vendor.testTopics)
		Expect(err).NotTo(HaveOccurred())
		Expect(c2.Close()).To(Succeed())
	})

})
