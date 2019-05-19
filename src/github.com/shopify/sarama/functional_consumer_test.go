package sarama

import (
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"testing"
	"time"
	"vendor"

	"github.com/stretchr/testify/require"
)

func TestFuncConsumerOffsetOutOfRange(t *testing.T) {
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	consumer, err := vendor.NewConsumer(vendor.kafkaBrokers, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := consumer.ConsumePartition("test.1", 0, -10); err != vendor.ErrOffsetOutOfRange {
		t.Error("Expected ErrOffsetOutOfRange, got:", err)
	}

	if _, err := consumer.ConsumePartition("test.1", 0, math.MaxInt64); err != vendor.ErrOffsetOutOfRange {
		t.Error("Expected ErrOffsetOutOfRange, got:", err)
	}

	vendor.safeClose(t, consumer)
}

func TestConsumerHighWaterMarkOffset(t *testing.T) {
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	p, err := vendor.NewSyncProducer(vendor.kafkaBrokers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer vendor.safeClose(t, p)

	_, offset, err := p.SendMessage(&vendor.ProducerMessage{Topic: "test.1", Value: vendor.StringEncoder("Test")})
	if err != nil {
		t.Fatal(err)
	}

	c, err := vendor.NewConsumer(vendor.kafkaBrokers, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer vendor.safeClose(t, c)

	pc, err := c.ConsumePartition("test.1", 0, offset)
	if err != nil {
		t.Fatal(err)
	}

	<-pc.Messages()

	if hwmo := pc.HighWaterMarkOffset(); hwmo != offset+1 {
		t.Logf("Last produced offset %d; high water mark should be one higher but found %d.", offset, hwmo)
	}

	vendor.safeClose(t, pc)
}

// Makes sure that messages produced by all supported client versions/
// compression codecs (except LZ4) combinations can be consumed by all
// supported consumer versions. It relies on the KAFKA_VERSION environment
// variable to provide the version of the test Kafka cluster.
//
// Note that LZ4 codec was introduced in v0.10.0.0 and therefore is excluded
// from this test case. It has a similar version matrix test case below that
// only checks versions from v0.10.0.0 until KAFKA_VERSION.
func TestVersionMatrix(t *testing.T) {
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	// Produce lot's of message with all possible combinations of supported
	// protocol versions and compressions for the except of LZ4.
	testVersions := versionRange(vendor.V0_8_2_0)
	allCodecsButLZ4 := []vendor.CompressionCodec{vendor.CompressionNone, vendor.CompressionGZIP, vendor.CompressionSnappy}
	producedMessages := produceMsgs(t, testVersions, allCodecsButLZ4, 17, 100, false)

	// When/Then
	consumeMsgs(t, testVersions, producedMessages)
}

// Support for LZ4 codec was introduced in v0.10.0.0 so a version matrix to
// test LZ4 should start with v0.10.0.0.
func TestVersionMatrixLZ4(t *testing.T) {
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	// Produce lot's of message with all possible combinations of supported
	// protocol versions starting with v0.10 (first where LZ4 was supported)
	// and all possible compressions.
	testVersions := versionRange(vendor.V0_10_0_0)
	allCodecs := []vendor.CompressionCodec{vendor.CompressionNone, vendor.CompressionGZIP, vendor.CompressionSnappy, vendor.CompressionLZ4}
	producedMessages := produceMsgs(t, testVersions, allCodecs, 17, 100, false)

	// When/Then
	consumeMsgs(t, testVersions, producedMessages)
}

func TestVersionMatrixIdempotent(t *testing.T) {
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	// Produce lot's of message with all possible combinations of supported
	// protocol versions starting with v0.11 (first where idempotent was supported)
	testVersions := versionRange(vendor.V0_11_0_0)
	producedMessages := produceMsgs(t, testVersions, []vendor.CompressionCodec{vendor.CompressionNone}, 17, 100, true)

	// When/Then
	consumeMsgs(t, testVersions, producedMessages)
}

func TestReadOnlyAndAllCommittedMessages(t *testing.T) {
	vendor.checkKafkaVersion(t, "0.11.0")
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	config := vendor.NewConfig()
	config.Consumer.IsolationLevel = vendor.ReadCommitted
	config.Version = vendor.V0_11_0_0

	consumer, err := vendor.NewConsumer(vendor.kafkaBrokers, config)
	if err != nil {
		t.Fatal(err)
	}

	pc, err := consumer.ConsumePartition("uncommitted-topic-test-4", 0, vendor.OffsetOldest)
	require.NoError(t, err)

	msgChannel := pc.Messages()
	for i := 1; i <= 6; i++ {
		msg := <-msgChannel
		require.Equal(t, fmt.Sprintf("Committed %v", i), string(msg.Value))
	}
}

func prodMsg2Str(prodMsg *vendor.ProducerMessage) string {
	return fmt.Sprintf("{offset: %d, value: %s}", prodMsg.Offset, string(prodMsg.Value.(vendor.StringEncoder)))
}

func consMsg2Str(consMsg *vendor.ConsumerMessage) string {
	return fmt.Sprintf("{offset: %d, value: %s}", consMsg.Offset, string(consMsg.Value))
}

func versionRange(lower vendor.KafkaVersion) []vendor.KafkaVersion {
	// Get the test cluster version from the environment. If there is nothing
	// there then assume the highest.
	upper, err := vendor.ParseKafkaVersion(os.Getenv("KAFKA_VERSION"))
	if err != nil {
		upper = vendor.MaxVersion
	}

	versions := make([]vendor.KafkaVersion, 0, len(vendor.SupportedVersions))
	for _, v := range vendor.SupportedVersions {
		if !v.IsAtLeast(lower) {
			continue
		}
		if !upper.IsAtLeast(v) {
			return versions
		}
		versions = append(versions, v)
	}
	return versions
}

func produceMsgs(t *testing.T, clientVersions []vendor.KafkaVersion, codecs []vendor.CompressionCodec, flush int, countPerVerCodec int, idempotent bool) []*vendor.ProducerMessage {
	var wg sync.WaitGroup
	var producedMessagesMu sync.Mutex
	var producedMessages []*vendor.ProducerMessage
	for _, prodVer := range clientVersions {
		for _, codec := range codecs {
			prodCfg := vendor.NewConfig()
			prodCfg.Version = prodVer
			prodCfg.Producer.Return.Successes = true
			prodCfg.Producer.Return.Errors = true
			prodCfg.Producer.Flush.MaxMessages = flush
			prodCfg.Producer.Compression = codec
			prodCfg.Producer.Idempotent = idempotent
			if idempotent {
				prodCfg.Producer.RequiredAcks = vendor.WaitForAll
				prodCfg.Net.MaxOpenRequests = 1
			}

			p, err := vendor.NewSyncProducer(vendor.kafkaBrokers, prodCfg)
			if err != nil {
				t.Errorf("Failed to create producer: version=%s, compression=%s, err=%v", prodVer, codec, err)
				continue
			}
			defer vendor.safeClose(t, p)
			for i := 0; i < countPerVerCodec; i++ {
				msg := &vendor.ProducerMessage{
					Topic: "test.1",
					Value: vendor.StringEncoder(fmt.Sprintf("msg:%s:%s:%d", prodVer, codec, i)),
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _, err := p.SendMessage(msg)
					if err != nil {
						t.Errorf("Failed to produce message: %s, err=%v", msg.Value, err)
					}
					producedMessagesMu.Lock()
					producedMessages = append(producedMessages, msg)
					producedMessagesMu.Unlock()
				}()
			}
		}
	}
	wg.Wait()

	// Sort produced message in ascending offset order.
	sort.Slice(producedMessages, func(i, j int) bool {
		return producedMessages[i].Offset < producedMessages[j].Offset
	})
	t.Logf("*** Total produced %d, firstOffset=%d, lastOffset=%d\n",
		len(producedMessages), producedMessages[0].Offset, producedMessages[len(producedMessages)-1].Offset)
	return producedMessages
}

func consumeMsgs(t *testing.T, clientVersions []vendor.KafkaVersion, producedMessages []*vendor.ProducerMessage) {
	// Consume all produced messages with all client versions supported by the
	// cluster.
consumerVersionLoop:
	for _, consVer := range clientVersions {
		t.Logf("*** Consuming with client version %s\n", consVer)
		// Create a partition consumer that should start from the first produced
		// message.
		consCfg := vendor.NewConfig()
		consCfg.Version = consVer
		c, err := vendor.NewConsumer(vendor.kafkaBrokers, consCfg)
		if err != nil {
			t.Fatal(err)
		}
		defer vendor.safeClose(t, c)
		pc, err := c.ConsumePartition("test.1", 0, producedMessages[0].Offset)
		if err != nil {
			t.Fatal(err)
		}
		defer vendor.safeClose(t, pc)

		// Consume as many messages as there have been produced and make sure that
		// order is preserved.
		for i, prodMsg := range producedMessages {
			select {
			case consMsg := <-pc.Messages():
				if consMsg.Offset != prodMsg.Offset {
					t.Errorf("Consumed unexpected offset: version=%s, index=%d, want=%s, got=%s",
						consVer, i, prodMsg2Str(prodMsg), consMsg2Str(consMsg))
					continue consumerVersionLoop
				}
				if string(consMsg.Value) != string(prodMsg.Value.(vendor.StringEncoder)) {
					t.Errorf("Consumed unexpected msg: version=%s, index=%d, want=%s, got=%s",
						consVer, i, prodMsg2Str(prodMsg), consMsg2Str(consMsg))
					continue consumerVersionLoop
				}
			case <-time.After(3 * time.Second):
				t.Fatalf("Timeout waiting for: index=%d, offset=%d, msg=%s", i, prodMsg.Offset, prodMsg.Value)
			}
		}
	}
}
