package mocks

import (
	"errors"
	"strings"
	"testing"
	"vendor"
)

func TestMockSyncProducerImplementsSyncProducerInterface(t *testing.T) {
	var mp interface{} = &vendor.SyncProducer{}
	if _, ok := mp.(vendor.SyncProducer); !ok {
		t.Error("The mock async producer should implement the sarama.SyncProducer interface.")
	}
}

func TestSyncProducerReturnsExpectationsToSendMessage(t *testing.T) {
	sp := vendor.NewSyncProducer(t, nil)
	defer func() {
		if err := sp.Close(); err != nil {
			t.Error(err)
		}
	}()

	sp.ExpectSendMessageAndSucceed()
	sp.ExpectSendMessageAndSucceed()
	sp.ExpectSendMessageAndFail(vendor.ErrOutOfBrokers)

	msg := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}

	_, offset, err := sp.SendMessage(msg)
	if err != nil {
		t.Errorf("The first message should have been produced successfully, but got %s", err)
	}
	if offset != 1 || offset != msg.Offset {
		t.Errorf("The first message should have been assigned offset 1, but got %d", msg.Offset)
	}

	_, offset, err = sp.SendMessage(msg)
	if err != nil {
		t.Errorf("The second message should have been produced successfully, but got %s", err)
	}
	if offset != 2 || offset != msg.Offset {
		t.Errorf("The second message should have been assigned offset 2, but got %d", offset)
	}

	_, _, err = sp.SendMessage(msg)
	if err != vendor.ErrOutOfBrokers {
		t.Errorf("The third message should not have been produced successfully")
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}
}

func TestSyncProducerWithTooManyExpectations(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageAndSucceed()
	sp.ExpectSendMessageAndFail(vendor.ErrOutOfBrokers)

	msg := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	if _, _, err := sp.SendMessage(msg); err != nil {
		t.Error("No error expected on first SendMessage call", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 1 {
		t.Error("Expected to report an error")
	}
}

func TestSyncProducerWithTooFewExpectations(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageAndSucceed()

	msg := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	if _, _, err := sp.SendMessage(msg); err != nil {
		t.Error("No error expected on first SendMessage call", err)
	}
	if _, _, err := sp.SendMessage(msg); err != vendor.errOutOfExpectations {
		t.Error("errOutOfExpectations expected on second SendMessage call, found:", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 1 {
		t.Error("Expected to report an error")
	}
}

func TestSyncProducerWithCheckerFunction(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes"))
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes$"))

	msg := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	if _, _, err := sp.SendMessage(msg); err != nil {
		t.Error("No error expected on first SendMessage call, found: ", err)
	}
	msg = &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	if _, _, err := sp.SendMessage(msg); err == nil || !strings.HasPrefix(err.Error(), "No match") {
		t.Error("Error during value check expected on second SendMessage call, found:", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 1 {
		t.Error("Expected to report an error")
	}
}

func TestSyncProducerWithCheckerFunctionForSendMessagesWithError(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes"))
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes$"))

	msg1 := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	msg2 := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	msgs := []*vendor.ProducerMessage{msg1, msg2}

	if err := sp.SendMessages(msgs); err == nil || !strings.HasPrefix(err.Error(), "No match") {
		t.Error("Error during value check expected on second message, found: ", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 1 {
		t.Error("Expected to report an error")
	}
}

func TestSyncProducerWithCheckerFunctionForSendMessagesWithoutError(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes"))

	msg1 := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	msgs := []*vendor.ProducerMessage{msg1}

	if err := sp.SendMessages(msgs); err != nil {
		t.Error("No error expected on SendMessages call, found: ", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 0 {
		t.Errorf("Expected to not report any errors, found: %v", trm.errors)
	}
}

func TestSyncProducerSendMessagesExpectationsMismatchTooFew(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes"))

	msg1 := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	msg2 := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}

	msgs := []*vendor.ProducerMessage{msg1, msg2}

	if err := sp.SendMessages(msgs); err == nil {
		t.Error("Error during value check expected on second message, found: ", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 2 {
		t.Error("Expected to report 2 errors")
	}
}

func TestSyncProducerSendMessagesExpectationsMismatchTooMany(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes"))
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes"))

	msg1 := &vendor.ProducerMessage{Topic: "test", Value: vendor.StringEncoder("test")}
	msgs := []*vendor.ProducerMessage{msg1}

	if err := sp.SendMessages(msgs); err != nil {
		t.Error("No error expected on SendMessages call, found: ", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 1 {
		t.Error("Expected to report 1 errors")
	}
}

func TestSyncProducerSendMessagesFaultyEncoder(t *testing.T) {
	trm := vendor.newTestReporterMock()

	sp := vendor.NewSyncProducer(trm, nil)
	sp.ExpectSendMessageWithCheckerFunctionAndSucceed(vendor.generateRegexpChecker("^tes"))

	msg1 := &vendor.ProducerMessage{Topic: "test", Value: faultyEncoder("123")}
	msgs := []*vendor.ProducerMessage{msg1}

	if err := sp.SendMessages(msgs); err == nil || !strings.HasPrefix(err.Error(), "encode error") {
		t.Error("Encoding error expected, found: ", err)
	}

	if err := sp.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 1 {
		t.Error("Expected to report 1 errors")
	}
}

type faultyEncoder []byte

func (f faultyEncoder) Encode() ([]byte, error) {
	return nil, errors.New("encode error")
}

func (f faultyEncoder) Length() int {
	return len(f)
}
