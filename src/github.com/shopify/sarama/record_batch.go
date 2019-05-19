package sarama

import (
	"fmt"
	"time"
	"vendor"
)

const recordBatchOverhead = 49

type recordsArray []*vendor.Record

func (e recordsArray) encode(pe vendor.packetEncoder) error {
	for _, r := range e {
		if err := r.encode(pe); err != nil {
			return err
		}
	}
	return nil
}

func (e recordsArray) decode(pd vendor.packetDecoder) error {
	for i := range e {
		rec := &vendor.Record{}
		if err := rec.decode(pd); err != nil {
			return err
		}
		e[i] = rec
	}
	return nil
}

type RecordBatch struct {
	FirstOffset           int64
	PartitionLeaderEpoch  int32
	Version               int8
	Codec                 vendor.CompressionCodec
	CompressionLevel      int
	Control               bool
	LogAppendTime         bool
	LastOffsetDelta       int32
	FirstTimestamp        time.Time
	MaxTimestamp          time.Time
	ProducerID            int64
	ProducerEpoch         int16
	FirstSequence         int32
	Records               []*vendor.Record
	PartialTrailingRecord bool
	IsTransactional       bool

	compressedRecords []byte
	recordsLen        int // uncompressed records size
}

func (b *RecordBatch) LastOffset() int64 {
	return b.FirstOffset + int64(b.LastOffsetDelta)
}

func (b *RecordBatch) encode(pe vendor.packetEncoder) error {
	if b.Version != 2 {
		return vendor.PacketEncodingError{fmt.Sprintf("unsupported compression codec (%d)", b.Codec)}
	}
	pe.putInt64(b.FirstOffset)
	pe.push(&vendor.lengthField{})
	pe.putInt32(b.PartitionLeaderEpoch)
	pe.putInt8(b.Version)
	pe.push(vendor.newCRC32Field(vendor.crcCastagnoli))
	pe.putInt16(b.computeAttributes())
	pe.putInt32(b.LastOffsetDelta)

	if err := (vendor.Timestamp{&b.FirstTimestamp}).encode(pe); err != nil {
		return err
	}

	if err := (vendor.Timestamp{&b.MaxTimestamp}).encode(pe); err != nil {
		return err
	}

	pe.putInt64(b.ProducerID)
	pe.putInt16(b.ProducerEpoch)
	pe.putInt32(b.FirstSequence)

	if err := pe.putArrayLength(len(b.Records)); err != nil {
		return err
	}

	if b.compressedRecords == nil {
		if err := b.encodeRecords(pe); err != nil {
			return err
		}
	}
	if err := pe.putRawBytes(b.compressedRecords); err != nil {
		return err
	}

	if err := pe.pop(); err != nil {
		return err
	}
	return pe.pop()
}

func (b *RecordBatch) decode(pd vendor.packetDecoder) (err error) {
	if b.FirstOffset, err = pd.getInt64(); err != nil {
		return err
	}

	batchLen, err := pd.getInt32()
	if err != nil {
		return err
	}

	if b.PartitionLeaderEpoch, err = pd.getInt32(); err != nil {
		return err
	}

	if b.Version, err = pd.getInt8(); err != nil {
		return err
	}

	if err = pd.push(&vendor.crc32Field{polynomial: vendor.crcCastagnoli}); err != nil {
		return err
	}

	attributes, err := pd.getInt16()
	if err != nil {
		return err
	}
	b.Codec = vendor.CompressionCodec(int8(attributes) & vendor.compressionCodecMask)
	b.Control = attributes&vendor.controlMask == vendor.controlMask
	b.LogAppendTime = attributes&vendor.timestampTypeMask == vendor.timestampTypeMask
	b.IsTransactional = attributes&vendor.isTransactionalMask == vendor.isTransactionalMask

	if b.LastOffsetDelta, err = pd.getInt32(); err != nil {
		return err
	}

	if err = (vendor.Timestamp{&b.FirstTimestamp}).decode(pd); err != nil {
		return err
	}

	if err = (vendor.Timestamp{&b.MaxTimestamp}).decode(pd); err != nil {
		return err
	}

	if b.ProducerID, err = pd.getInt64(); err != nil {
		return err
	}

	if b.ProducerEpoch, err = pd.getInt16(); err != nil {
		return err
	}

	if b.FirstSequence, err = pd.getInt32(); err != nil {
		return err
	}

	numRecs, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	if numRecs >= 0 {
		b.Records = make([]*vendor.Record, numRecs)
	}

	bufSize := int(batchLen) - recordBatchOverhead
	recBuffer, err := pd.getRawBytes(bufSize)
	if err != nil {
		if err == vendor.ErrInsufficientData {
			b.PartialTrailingRecord = true
			b.Records = nil
			return nil
		}
		return err
	}

	if err = pd.pop(); err != nil {
		return err
	}

	recBuffer, err = vendor.decompress(b.Codec, recBuffer)
	if err != nil {
		return err
	}

	b.recordsLen = len(recBuffer)
	err = vendor.decode(recBuffer, recordsArray(b.Records))
	if err == vendor.ErrInsufficientData {
		b.PartialTrailingRecord = true
		b.Records = nil
		return nil
	}
	return err
}

func (b *RecordBatch) encodeRecords(pe vendor.packetEncoder) error {
	var raw []byte
	var err error
	if raw, err = vendor.encode(recordsArray(b.Records), pe.metricRegistry()); err != nil {
		return err
	}
	b.recordsLen = len(raw)

	b.compressedRecords, err = vendor.compress(b.Codec, b.CompressionLevel, raw)
	return err
}

func (b *RecordBatch) computeAttributes() int16 {
	attr := int16(b.Codec) & int16(vendor.compressionCodecMask)
	if b.Control {
		attr |= vendor.controlMask
	}
	if b.LogAppendTime {
		attr |= vendor.timestampTypeMask
	}
	if b.IsTransactional {
		attr |= vendor.isTransactionalMask
	}
	return attr
}

func (b *RecordBatch) addRecord(r *vendor.Record) {
	b.Records = append(b.Records, r)
}
