package sarama

import (
	"fmt"
	"strings"
	"vendor"
)

// TestReporter has methods matching go's testing.T to avoid importing
// `testing` in the main part of the library.
type TestReporter interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

// MockResponse is a response builder interface it defines one method that
// allows generating a response based on a request body. MockResponses are used
// to program behavior of MockBroker in tests.
type MockResponse interface {
	For(reqBody vendor.versionedDecoder) (res vendor.encoder)
}

// MockWrapper is a mock response builder that returns a particular concrete
// response regardless of the actual request passed to the `For` method.
type MockWrapper struct {
	res vendor.encoder
}

func (mw *MockWrapper) For(reqBody vendor.versionedDecoder) (res vendor.encoder) {
	return mw.res
}

func NewMockWrapper(res vendor.encoder) *MockWrapper {
	return &MockWrapper{res: res}
}

// MockSequence is a mock response builder that is created from a sequence of
// concrete responses. Every time when a `MockBroker` calls its `For` method
// the next response from the sequence is returned. When the end of the
// sequence is reached the last element from the sequence is returned.
type MockSequence struct {
	responses []MockResponse
}

func NewMockSequence(responses ...interface{}) *MockSequence {
	ms := &MockSequence{}
	ms.responses = make([]MockResponse, len(responses))
	for i, res := range responses {
		switch res := res.(type) {
		case MockResponse:
			ms.responses[i] = res
		case vendor.encoder:
			ms.responses[i] = NewMockWrapper(res)
		default:
			panic(fmt.Sprintf("Unexpected response type: %T", res))
		}
	}
	return ms
}

func (mc *MockSequence) For(reqBody vendor.versionedDecoder) (res vendor.encoder) {
	res = mc.responses[0].For(reqBody)
	if len(mc.responses) > 1 {
		mc.responses = mc.responses[1:]
	}
	return res
}

type MockListGroupsResponse struct {
	groups map[string]string
	t      TestReporter
}

func NewMockListGroupsResponse(t TestReporter) *MockListGroupsResponse {
	return &MockListGroupsResponse{
		groups: make(map[string]string),
		t:      t,
	}
}

func (m *MockListGroupsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	request := reqBody.(*vendor.ListGroupsRequest)
	_ = request
	response := &vendor.ListGroupsResponse{
		Groups: m.groups,
	}
	return response
}

func (m *MockListGroupsResponse) AddGroup(groupID, protocolType string) *MockListGroupsResponse {
	m.groups[groupID] = protocolType
	return m
}

type MockDescribeGroupsResponse struct {
	groups map[string]*vendor.GroupDescription
	t      TestReporter
}

func NewMockDescribeGroupsResponse(t TestReporter) *MockDescribeGroupsResponse {
	return &MockDescribeGroupsResponse{
		t:      t,
		groups: make(map[string]*vendor.GroupDescription),
	}
}

func (m *MockDescribeGroupsResponse) AddGroupDescription(groupID string, description *vendor.GroupDescription) *MockDescribeGroupsResponse {
	m.groups[groupID] = description
	return m
}

func (m *MockDescribeGroupsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	request := reqBody.(*vendor.DescribeGroupsRequest)

	response := &vendor.DescribeGroupsResponse{}
	for _, requestedGroup := range request.Groups {
		if group, ok := m.groups[requestedGroup]; ok {
			response.Groups = append(response.Groups, group)
		} else {
			// Mimic real kafka - if a group doesn't exist, return
			// an entry with state "Dead"
			response.Groups = append(response.Groups, &vendor.GroupDescription{
				GroupId: requestedGroup,
				State:   "Dead",
			})
		}
	}

	return response
}

// MockMetadataResponse is a `MetadataResponse` builder.
type MockMetadataResponse struct {
	controllerID int32
	leaders      map[string]map[int32]int32
	brokers      map[string]int32
	t            TestReporter
}

func NewMockMetadataResponse(t TestReporter) *MockMetadataResponse {
	return &MockMetadataResponse{
		leaders: make(map[string]map[int32]int32),
		brokers: make(map[string]int32),
		t:       t,
	}
}

func (mmr *MockMetadataResponse) SetLeader(topic string, partition, brokerID int32) *MockMetadataResponse {
	partitions := mmr.leaders[topic]
	if partitions == nil {
		partitions = make(map[int32]int32)
		mmr.leaders[topic] = partitions
	}
	partitions[partition] = brokerID
	return mmr
}

func (mmr *MockMetadataResponse) SetBroker(addr string, brokerID int32) *MockMetadataResponse {
	mmr.brokers[addr] = brokerID
	return mmr
}

func (mmr *MockMetadataResponse) SetController(brokerID int32) *MockMetadataResponse {
	mmr.controllerID = brokerID
	return mmr
}

func (mmr *MockMetadataResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	metadataRequest := reqBody.(*vendor.MetadataRequest)
	metadataResponse := &vendor.MetadataResponse{
		Version:      metadataRequest.version(),
		ControllerID: mmr.controllerID,
	}
	for addr, brokerID := range mmr.brokers {
		metadataResponse.AddBroker(addr, brokerID)
	}

	// Generate set of replicas
	replicas := []int32{}
	offlineReplicas := []int32{}
	for _, brokerID := range mmr.brokers {
		replicas = append(replicas, brokerID)
	}

	if len(metadataRequest.Topics) == 0 {
		for topic, partitions := range mmr.leaders {
			for partition, brokerID := range partitions {
				metadataResponse.AddTopicPartition(topic, partition, brokerID, replicas, replicas, offlineReplicas, vendor.ErrNoError)
			}
		}
		return metadataResponse
	}
	for _, topic := range metadataRequest.Topics {
		for partition, brokerID := range mmr.leaders[topic] {
			metadataResponse.AddTopicPartition(topic, partition, brokerID, replicas, replicas, offlineReplicas, vendor.ErrNoError)
		}
	}
	return metadataResponse
}

// MockOffsetResponse is an `OffsetResponse` builder.
type MockOffsetResponse struct {
	offsets map[string]map[int32]map[int64]int64
	t       TestReporter
	version int16
}

func NewMockOffsetResponse(t TestReporter) *MockOffsetResponse {
	return &MockOffsetResponse{
		offsets: make(map[string]map[int32]map[int64]int64),
		t:       t,
	}
}

func (mor *MockOffsetResponse) SetVersion(version int16) *MockOffsetResponse {
	mor.version = version
	return mor
}

func (mor *MockOffsetResponse) SetOffset(topic string, partition int32, time, offset int64) *MockOffsetResponse {
	partitions := mor.offsets[topic]
	if partitions == nil {
		partitions = make(map[int32]map[int64]int64)
		mor.offsets[topic] = partitions
	}
	times := partitions[partition]
	if times == nil {
		times = make(map[int64]int64)
		partitions[partition] = times
	}
	times[time] = offset
	return mor
}

func (mor *MockOffsetResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	offsetRequest := reqBody.(*vendor.OffsetRequest)
	offsetResponse := &vendor.OffsetResponse{Version: mor.version}
	for topic, partitions := range offsetRequest.blocks {
		for partition, block := range partitions {
			offset := mor.getOffset(topic, partition, block.time)
			offsetResponse.AddTopicPartition(topic, partition, offset)
		}
	}
	return offsetResponse
}

func (mor *MockOffsetResponse) getOffset(topic string, partition int32, time int64) int64 {
	partitions := mor.offsets[topic]
	if partitions == nil {
		mor.t.Errorf("missing topic: %s", topic)
	}
	times := partitions[partition]
	if times == nil {
		mor.t.Errorf("missing partition: %d", partition)
	}
	offset, ok := times[time]
	if !ok {
		mor.t.Errorf("missing time: %d", time)
	}
	return offset
}

// MockFetchResponse is a `FetchResponse` builder.
type MockFetchResponse struct {
	messages       map[string]map[int32]map[int64]vendor.Encoder
	highWaterMarks map[string]map[int32]int64
	t              TestReporter
	batchSize      int
	version        int16
}

func NewMockFetchResponse(t TestReporter, batchSize int) *MockFetchResponse {
	return &MockFetchResponse{
		messages:       make(map[string]map[int32]map[int64]vendor.Encoder),
		highWaterMarks: make(map[string]map[int32]int64),
		t:              t,
		batchSize:      batchSize,
	}
}

func (mfr *MockFetchResponse) SetVersion(version int16) *MockFetchResponse {
	mfr.version = version
	return mfr
}

func (mfr *MockFetchResponse) SetMessage(topic string, partition int32, offset int64, msg vendor.Encoder) *MockFetchResponse {
	partitions := mfr.messages[topic]
	if partitions == nil {
		partitions = make(map[int32]map[int64]vendor.Encoder)
		mfr.messages[topic] = partitions
	}
	messages := partitions[partition]
	if messages == nil {
		messages = make(map[int64]vendor.Encoder)
		partitions[partition] = messages
	}
	messages[offset] = msg
	return mfr
}

func (mfr *MockFetchResponse) SetHighWaterMark(topic string, partition int32, offset int64) *MockFetchResponse {
	partitions := mfr.highWaterMarks[topic]
	if partitions == nil {
		partitions = make(map[int32]int64)
		mfr.highWaterMarks[topic] = partitions
	}
	partitions[partition] = offset
	return mfr
}

func (mfr *MockFetchResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	fetchRequest := reqBody.(*vendor.FetchRequest)
	res := &vendor.FetchResponse{
		Version: mfr.version,
	}
	for topic, partitions := range fetchRequest.blocks {
		for partition, block := range partitions {
			initialOffset := block.fetchOffset
			offset := initialOffset
			maxOffset := initialOffset + int64(mfr.getMessageCount(topic, partition))
			for i := 0; i < mfr.batchSize && offset < maxOffset; {
				msg := mfr.getMessage(topic, partition, offset)
				if msg != nil {
					res.AddMessage(topic, partition, nil, msg, offset)
					i++
				}
				offset++
			}
			fb := res.GetBlock(topic, partition)
			if fb == nil {
				res.AddError(topic, partition, vendor.ErrNoError)
				fb = res.GetBlock(topic, partition)
			}
			fb.HighWaterMarkOffset = mfr.getHighWaterMark(topic, partition)
		}
	}
	return res
}

func (mfr *MockFetchResponse) getMessage(topic string, partition int32, offset int64) vendor.Encoder {
	partitions := mfr.messages[topic]
	if partitions == nil {
		return nil
	}
	messages := partitions[partition]
	if messages == nil {
		return nil
	}
	return messages[offset]
}

func (mfr *MockFetchResponse) getMessageCount(topic string, partition int32) int {
	partitions := mfr.messages[topic]
	if partitions == nil {
		return 0
	}
	messages := partitions[partition]
	if messages == nil {
		return 0
	}
	return len(messages)
}

func (mfr *MockFetchResponse) getHighWaterMark(topic string, partition int32) int64 {
	partitions := mfr.highWaterMarks[topic]
	if partitions == nil {
		return 0
	}
	return partitions[partition]
}

// MockConsumerMetadataResponse is a `ConsumerMetadataResponse` builder.
type MockConsumerMetadataResponse struct {
	coordinators map[string]interface{}
	t            TestReporter
}

func NewMockConsumerMetadataResponse(t TestReporter) *MockConsumerMetadataResponse {
	return &MockConsumerMetadataResponse{
		coordinators: make(map[string]interface{}),
		t:            t,
	}
}

func (mr *MockConsumerMetadataResponse) SetCoordinator(group string, broker *vendor.MockBroker) *MockConsumerMetadataResponse {
	mr.coordinators[group] = broker
	return mr
}

func (mr *MockConsumerMetadataResponse) SetError(group string, kerror vendor.KError) *MockConsumerMetadataResponse {
	mr.coordinators[group] = kerror
	return mr
}

func (mr *MockConsumerMetadataResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.ConsumerMetadataRequest)
	group := req.ConsumerGroup
	res := &vendor.ConsumerMetadataResponse{}
	v := mr.coordinators[group]
	switch v := v.(type) {
	case *vendor.MockBroker:
		res.Coordinator = &vendor.Broker{id: v.BrokerID(), addr: v.Addr()}
	case vendor.KError:
		res.Err = v
	}
	return res
}

// MockFindCoordinatorResponse is a `FindCoordinatorResponse` builder.
type MockFindCoordinatorResponse struct {
	groupCoordinators map[string]interface{}
	transCoordinators map[string]interface{}
	t                 TestReporter
}

func NewMockFindCoordinatorResponse(t TestReporter) *MockFindCoordinatorResponse {
	return &MockFindCoordinatorResponse{
		groupCoordinators: make(map[string]interface{}),
		transCoordinators: make(map[string]interface{}),
		t:                 t,
	}
}

func (mr *MockFindCoordinatorResponse) SetCoordinator(coordinatorType vendor.CoordinatorType, group string, broker *vendor.MockBroker) *MockFindCoordinatorResponse {
	switch coordinatorType {
	case vendor.CoordinatorGroup:
		mr.groupCoordinators[group] = broker
	case vendor.CoordinatorTransaction:
		mr.transCoordinators[group] = broker
	}
	return mr
}

func (mr *MockFindCoordinatorResponse) SetError(coordinatorType vendor.CoordinatorType, group string, kerror vendor.KError) *MockFindCoordinatorResponse {
	switch coordinatorType {
	case vendor.CoordinatorGroup:
		mr.groupCoordinators[group] = kerror
	case vendor.CoordinatorTransaction:
		mr.transCoordinators[group] = kerror
	}
	return mr
}

func (mr *MockFindCoordinatorResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.FindCoordinatorRequest)
	res := &vendor.FindCoordinatorResponse{}
	var v interface{}
	switch req.CoordinatorType {
	case vendor.CoordinatorGroup:
		v = mr.groupCoordinators[req.CoordinatorKey]
	case vendor.CoordinatorTransaction:
		v = mr.transCoordinators[req.CoordinatorKey]
	}
	switch v := v.(type) {
	case *vendor.MockBroker:
		res.Coordinator = &vendor.Broker{id: v.BrokerID(), addr: v.Addr()}
	case vendor.KError:
		res.Err = v
	}
	return res
}

// MockOffsetCommitResponse is a `OffsetCommitResponse` builder.
type MockOffsetCommitResponse struct {
	errors map[string]map[string]map[int32]vendor.KError
	t      TestReporter
}

func NewMockOffsetCommitResponse(t TestReporter) *MockOffsetCommitResponse {
	return &MockOffsetCommitResponse{t: t}
}

func (mr *MockOffsetCommitResponse) SetError(group, topic string, partition int32, kerror vendor.KError) *MockOffsetCommitResponse {
	if mr.errors == nil {
		mr.errors = make(map[string]map[string]map[int32]vendor.KError)
	}
	topics := mr.errors[group]
	if topics == nil {
		topics = make(map[string]map[int32]vendor.KError)
		mr.errors[group] = topics
	}
	partitions := topics[topic]
	if partitions == nil {
		partitions = make(map[int32]vendor.KError)
		topics[topic] = partitions
	}
	partitions[partition] = kerror
	return mr
}

func (mr *MockOffsetCommitResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.OffsetCommitRequest)
	group := req.ConsumerGroup
	res := &vendor.OffsetCommitResponse{}
	for topic, partitions := range req.blocks {
		for partition := range partitions {
			res.AddError(topic, partition, mr.getError(group, topic, partition))
		}
	}
	return res
}

func (mr *MockOffsetCommitResponse) getError(group, topic string, partition int32) vendor.KError {
	topics := mr.errors[group]
	if topics == nil {
		return vendor.ErrNoError
	}
	partitions := topics[topic]
	if partitions == nil {
		return vendor.ErrNoError
	}
	kerror, ok := partitions[partition]
	if !ok {
		return vendor.ErrNoError
	}
	return kerror
}

// MockProduceResponse is a `ProduceResponse` builder.
type MockProduceResponse struct {
	version int16
	errors  map[string]map[int32]vendor.KError
	t       TestReporter
}

func NewMockProduceResponse(t TestReporter) *MockProduceResponse {
	return &MockProduceResponse{t: t}
}

func (mr *MockProduceResponse) SetVersion(version int16) *MockProduceResponse {
	mr.version = version
	return mr
}

func (mr *MockProduceResponse) SetError(topic string, partition int32, kerror vendor.KError) *MockProduceResponse {
	if mr.errors == nil {
		mr.errors = make(map[string]map[int32]vendor.KError)
	}
	partitions := mr.errors[topic]
	if partitions == nil {
		partitions = make(map[int32]vendor.KError)
		mr.errors[topic] = partitions
	}
	partitions[partition] = kerror
	return mr
}

func (mr *MockProduceResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.ProduceRequest)
	res := &vendor.ProduceResponse{
		Version: mr.version,
	}
	for topic, partitions := range req.records {
		for partition := range partitions {
			res.AddTopicPartition(topic, partition, mr.getError(topic, partition))
		}
	}
	return res
}

func (mr *MockProduceResponse) getError(topic string, partition int32) vendor.KError {
	partitions := mr.errors[topic]
	if partitions == nil {
		return vendor.ErrNoError
	}
	kerror, ok := partitions[partition]
	if !ok {
		return vendor.ErrNoError
	}
	return kerror
}

// MockOffsetFetchResponse is a `OffsetFetchResponse` builder.
type MockOffsetFetchResponse struct {
	offsets map[string]map[string]map[int32]*vendor.OffsetFetchResponseBlock
	t       TestReporter
}

func NewMockOffsetFetchResponse(t TestReporter) *MockOffsetFetchResponse {
	return &MockOffsetFetchResponse{t: t}
}

func (mr *MockOffsetFetchResponse) SetOffset(group, topic string, partition int32, offset int64, metadata string, kerror vendor.KError) *MockOffsetFetchResponse {
	if mr.offsets == nil {
		mr.offsets = make(map[string]map[string]map[int32]*vendor.OffsetFetchResponseBlock)
	}
	topics := mr.offsets[group]
	if topics == nil {
		topics = make(map[string]map[int32]*vendor.OffsetFetchResponseBlock)
		mr.offsets[group] = topics
	}
	partitions := topics[topic]
	if partitions == nil {
		partitions = make(map[int32]*vendor.OffsetFetchResponseBlock)
		topics[topic] = partitions
	}
	partitions[partition] = &vendor.OffsetFetchResponseBlock{offset, 0, metadata, kerror}
	return mr
}

func (mr *MockOffsetFetchResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.OffsetFetchRequest)
	group := req.ConsumerGroup
	res := &vendor.OffsetFetchResponse{}
	for topic, partitions := range mr.offsets[group] {
		for partition, block := range partitions {
			res.AddBlock(topic, partition, block)
		}
	}
	return res
}

type MockCreateTopicsResponse struct {
	t TestReporter
}

func NewMockCreateTopicsResponse(t TestReporter) *MockCreateTopicsResponse {
	return &MockCreateTopicsResponse{t: t}
}

func (mr *MockCreateTopicsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.CreateTopicsRequest)
	res := &vendor.CreateTopicsResponse{
		Version: req.Version,
	}
	res.TopicErrors = make(map[string]*vendor.TopicError)

	for topic := range req.TopicDetails {
		if res.Version >= 1 && strings.HasPrefix(topic, "_") {
			msg := "insufficient permissions to create topic with reserved prefix"
			res.TopicErrors[topic] = &vendor.TopicError{
				Err:    vendor.ErrTopicAuthorizationFailed,
				ErrMsg: &msg,
			}
			continue
		}
		res.TopicErrors[topic] = &vendor.TopicError{Err: vendor.ErrNoError}
	}
	return res
}

type MockDeleteTopicsResponse struct {
	t TestReporter
}

func NewMockDeleteTopicsResponse(t TestReporter) *MockDeleteTopicsResponse {
	return &MockDeleteTopicsResponse{t: t}
}

func (mr *MockDeleteTopicsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.DeleteTopicsRequest)
	res := &vendor.DeleteTopicsResponse{}
	res.TopicErrorCodes = make(map[string]vendor.KError)

	for _, topic := range req.Topics {
		res.TopicErrorCodes[topic] = vendor.ErrNoError
	}
	return res
}

type MockCreatePartitionsResponse struct {
	t TestReporter
}

func NewMockCreatePartitionsResponse(t TestReporter) *MockCreatePartitionsResponse {
	return &MockCreatePartitionsResponse{t: t}
}

func (mr *MockCreatePartitionsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.CreatePartitionsRequest)
	res := &vendor.CreatePartitionsResponse{}
	res.TopicPartitionErrors = make(map[string]*vendor.TopicPartitionError)

	for topic := range req.TopicPartitions {
		if strings.HasPrefix(topic, "_") {
			msg := "insufficient permissions to create partition on topic with reserved prefix"
			res.TopicPartitionErrors[topic] = &vendor.TopicPartitionError{
				Err:    vendor.ErrTopicAuthorizationFailed,
				ErrMsg: &msg,
			}
			continue
		}
		res.TopicPartitionErrors[topic] = &vendor.TopicPartitionError{Err: vendor.ErrNoError}
	}
	return res
}

type MockDeleteRecordsResponse struct {
	t TestReporter
}

func NewMockDeleteRecordsResponse(t TestReporter) *MockDeleteRecordsResponse {
	return &MockDeleteRecordsResponse{t: t}
}

func (mr *MockDeleteRecordsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.DeleteRecordsRequest)
	res := &vendor.DeleteRecordsResponse{}
	res.Topics = make(map[string]*vendor.DeleteRecordsResponseTopic)

	for topic, deleteRecordRequestTopic := range req.Topics {
		partitions := make(map[int32]*vendor.DeleteRecordsResponsePartition)
		for partition := range deleteRecordRequestTopic.PartitionOffsets {
			partitions[partition] = &vendor.DeleteRecordsResponsePartition{Err: vendor.ErrNoError}
		}
		res.Topics[topic] = &vendor.DeleteRecordsResponseTopic{Partitions: partitions}
	}
	return res
}

type MockDescribeConfigsResponse struct {
	t TestReporter
}

func NewMockDescribeConfigsResponse(t TestReporter) *MockDescribeConfigsResponse {
	return &MockDescribeConfigsResponse{t: t}
}

func (mr *MockDescribeConfigsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.DescribeConfigsRequest)
	res := &vendor.DescribeConfigsResponse{}

	for _, r := range req.Resources {
		var configEntries []*vendor.ConfigEntry
		switch r.Type {
		case vendor.TopicResource:
			configEntries = append(configEntries,
				&vendor.ConfigEntry{Name: "max.message.bytes",
					Value:     "1000000",
					ReadOnly:  false,
					Default:   true,
					Sensitive: false,
				}, &vendor.ConfigEntry{Name: "retention.ms",
					Value:     "5000",
					ReadOnly:  false,
					Default:   false,
					Sensitive: false,
				}, &vendor.ConfigEntry{Name: "password",
					Value:     "12345",
					ReadOnly:  false,
					Default:   false,
					Sensitive: true,
				})
			res.Resources = append(res.Resources, &vendor.ResourceResponse{
				Name:    r.Name,
				Configs: configEntries,
			})
		}
	}
	return res
}

type MockAlterConfigsResponse struct {
	t TestReporter
}

func NewMockAlterConfigsResponse(t TestReporter) *MockAlterConfigsResponse {
	return &MockAlterConfigsResponse{t: t}
}

func (mr *MockAlterConfigsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.AlterConfigsRequest)
	res := &vendor.AlterConfigsResponse{}

	for _, r := range req.Resources {
		res.Resources = append(res.Resources, &vendor.AlterConfigsResourceResponse{Name: r.Name,
			Type:     vendor.TopicResource,
			ErrorMsg: "",
		})
	}
	return res
}

type MockCreateAclsResponse struct {
	t TestReporter
}

func NewMockCreateAclsResponse(t TestReporter) *MockCreateAclsResponse {
	return &MockCreateAclsResponse{t: t}
}

func (mr *MockCreateAclsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.CreateAclsRequest)
	res := &vendor.CreateAclsResponse{}

	for range req.AclCreations {
		res.AclCreationResponses = append(res.AclCreationResponses, &vendor.AclCreationResponse{Err: vendor.ErrNoError})
	}
	return res
}

type MockListAclsResponse struct {
	t TestReporter
}

func NewMockListAclsResponse(t TestReporter) *MockListAclsResponse {
	return &MockListAclsResponse{t: t}
}

func (mr *MockListAclsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.DescribeAclsRequest)
	res := &vendor.DescribeAclsResponse{}

	res.Err = vendor.ErrNoError
	acl := &vendor.ResourceAcls{}
	acl.Resource.ResourceName = *req.ResourceName
	acl.Resource.ResourceType = req.ResourceType
	acl.Acls = append(acl.Acls, &vendor.Acl{})
	res.ResourceAcls = append(res.ResourceAcls, acl)

	return res
}

type MockSaslAuthenticateResponse struct {
	t             TestReporter
	kerror        vendor.KError
	saslAuthBytes []byte
}

func NewMockSaslAuthenticateResponse(t TestReporter) *MockSaslAuthenticateResponse {
	return &MockSaslAuthenticateResponse{t: t}
}

func (msar *MockSaslAuthenticateResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	res := &vendor.SaslAuthenticateResponse{}
	res.Err = msar.kerror
	res.SaslAuthBytes = msar.saslAuthBytes
	return res
}

func (msar *MockSaslAuthenticateResponse) SetError(kerror vendor.KError) *MockSaslAuthenticateResponse {
	msar.kerror = kerror
	return msar
}

func (msar *MockSaslAuthenticateResponse) SetAuthBytes(saslAuthBytes []byte) *MockSaslAuthenticateResponse {
	msar.saslAuthBytes = saslAuthBytes
	return msar
}

type MockDeleteAclsResponse struct {
	t TestReporter
}

type MockSaslHandshakeResponse struct {
	enabledMechanisms []string
	kerror            vendor.KError
	t                 TestReporter
}

func NewMockSaslHandshakeResponse(t TestReporter) *MockSaslHandshakeResponse {
	return &MockSaslHandshakeResponse{t: t}
}

func (mshr *MockSaslHandshakeResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	res := &vendor.SaslHandshakeResponse{}
	res.Err = mshr.kerror
	res.EnabledMechanisms = mshr.enabledMechanisms
	return res
}

func (mshr *MockSaslHandshakeResponse) SetError(kerror vendor.KError) *MockSaslHandshakeResponse {
	mshr.kerror = kerror
	return mshr
}

func (mshr *MockSaslHandshakeResponse) SetEnabledMechanisms(enabledMechanisms []string) *MockSaslHandshakeResponse {
	mshr.enabledMechanisms = enabledMechanisms
	return mshr
}

func NewMockDeleteAclsResponse(t TestReporter) *MockDeleteAclsResponse {
	return &MockDeleteAclsResponse{t: t}
}

func (mr *MockDeleteAclsResponse) For(reqBody vendor.versionedDecoder) vendor.encoder {
	req := reqBody.(*vendor.DeleteAclsRequest)
	res := &vendor.DeleteAclsResponse{}

	for range req.Filters {
		response := &vendor.FilterResponse{Err: vendor.ErrNoError}
		response.MatchingAcls = append(response.MatchingAcls, &vendor.MatchingAcl{Err: vendor.ErrNoError})
		res.FilterResponses = append(res.FilterResponses, response)
	}
	return res
}
