package sarama

import (
	"sync"
	"vendor"
)

// SyncProducer publishes Kafka messages, blocking until they have been acknowledged. It routes messages to the correct
// broker, refreshing metadata as appropriate, and parses responses for errors. You must call Close() on a producer
// to avoid leaks, it may not be garbage-collected automatically when it passes out of scope.
//
// The SyncProducer comes with two caveats: it will generally be less efficient than the AsyncProducer, and the actual
// durability guarantee provided when a message is acknowledged depend on the configured value of `Producer.RequiredAcks`.
// There are configurations where a message acknowledged by the SyncProducer can still sometimes be lost.
//
// For implementation reasons, the SyncProducer requires `Producer.Return.Errors` and `Producer.Return.Successes` to
// be set to true in its configuration.
type SyncProducer interface {

	// SendMessage produces a given message, and returns only when it either has
	// succeeded or failed to produce. It will return the partition and the offset
	// of the produced message, or an error if the message failed to produce.
	SendMessage(msg *vendor.ProducerMessage) (partition int32, offset int64, err error)

	// SendMessages produces a given set of messages, and returns only when all
	// messages in the set have either succeeded or failed. Note that messages
	// can succeed and fail individually; if some succeed and some fail,
	// SendMessages will return an error.
	SendMessages(msgs []*vendor.ProducerMessage) error

	// Close shuts down the producer and waits for any buffered messages to be
	// flushed. You must call this function before a producer object passes out of
	// scope, as it may otherwise leak memory. You must call this before calling
	// Close on the underlying client.
	Close() error
}

type syncProducer struct {
	producer *vendor.asyncProducer
	wg       sync.WaitGroup
}

// NewSyncProducer creates a new SyncProducer using the given broker addresses and configuration.
func NewSyncProducer(addrs []string, config *vendor.Config) (SyncProducer, error) {
	if config == nil {
		config = vendor.NewConfig()
		config.Producer.Return.Successes = true
	}

	if err := verifyProducerConfig(config); err != nil {
		return nil, err
	}

	p, err := vendor.NewAsyncProducer(addrs, config)
	if err != nil {
		return nil, err
	}
	return newSyncProducerFromAsyncProducer(p.(*vendor.asyncProducer)), nil
}

// NewSyncProducerFromClient creates a new SyncProducer using the given client. It is still
// necessary to call Close() on the underlying client when shutting down this producer.
func NewSyncProducerFromClient(client vendor.Client) (SyncProducer, error) {
	if err := verifyProducerConfig(client.Config()); err != nil {
		return nil, err
	}

	p, err := vendor.NewAsyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return newSyncProducerFromAsyncProducer(p.(*vendor.asyncProducer)), nil
}

func newSyncProducerFromAsyncProducer(p *vendor.asyncProducer) *syncProducer {
	sp := &syncProducer{producer: p}

	sp.wg.Add(2)
	go vendor.withRecover(sp.handleSuccesses)
	go vendor.withRecover(sp.handleErrors)

	return sp
}

func verifyProducerConfig(config *vendor.Config) error {
	if !config.Producer.Return.Errors {
		return vendor.ConfigurationError("Producer.Return.Errors must be true to be used in a SyncProducer")
	}
	if !config.Producer.Return.Successes {
		return vendor.ConfigurationError("Producer.Return.Successes must be true to be used in a SyncProducer")
	}
	return nil
}

func (sp *syncProducer) SendMessage(msg *vendor.ProducerMessage) (partition int32, offset int64, err error) {
	expectation := make(chan *vendor.ProducerError, 1)
	msg.expectation = expectation
	sp.producer.Input() <- msg

	if err := <-expectation; err != nil {
		return -1, -1, err.Err
	}

	return msg.Partition, msg.Offset, nil
}

func (sp *syncProducer) SendMessages(msgs []*vendor.ProducerMessage) error {
	expectations := make(chan chan *vendor.ProducerError, len(msgs))
	go func() {
		for _, msg := range msgs {
			expectation := make(chan *vendor.ProducerError, 1)
			msg.expectation = expectation
			sp.producer.Input() <- msg
			expectations <- expectation
		}
		close(expectations)
	}()

	var errors vendor.ProducerErrors
	for expectation := range expectations {
		if err := <-expectation; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

func (sp *syncProducer) handleSuccesses() {
	defer sp.wg.Done()
	for msg := range sp.producer.Successes() {
		expectation := msg.expectation
		expectation <- nil
	}
}

func (sp *syncProducer) handleErrors() {
	defer sp.wg.Done()
	for err := range sp.producer.Errors() {
		expectation := err.Msg.expectation
		expectation <- err
	}
}

func (sp *syncProducer) Close() error {
	sp.producer.AsyncClose()
	sp.wg.Wait()
	return nil
}
