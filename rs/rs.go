package rs

import (
	"fmt"
	"io"

	"github.com/klauspost/reedsolomon"
	stream "github.com/zhangmingkai4315/go-storage/stream"
)

const (
	DATA_SHARDS     = 4
	PARITY_SHARDS   = 2
	ALL_SHARDS      = DATA_SHARDS + PARITY_SHARDS
	BLOCK_PER_SHARD = 8000
	BLOCK_SIZE      = BLOCK_PER_SHARD * DATA_SHARDS
)

type encoder struct {
	writers []io.Writer
	enc     reedsolomon.Encoder
	cache   []byte
}

type RSPutStream struct {
	*encoder
}

func NewRSPutStream(dataServers []string, hash string, size int64) (*RSPutStream, error) {
	if len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("Data servers number mismatch")
	}
	perShard := (size + DATA_SHARDS - 1) / DATA_SHARDS
	writers := make([]io.Writer, ALL_SHARDS)
	var err error
	for i := range writers {
		writers[i], err = stream.NewTempPutStream(dataServers[i], fmt.Sprintf("%s.%d", hash, i), perShard)
		if err != nil {
			return nil, err
		}
	}
	enc := NewEncoder(writers)
	return &RSPutStream{enc}, nil
}

func NewEncoder(writers []io.Writer) *encoder {
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &encoder{writers, enc, nil}
}

func (e *encoder) Flush() {
	if len(e.cache) == 0 {
		return
	}
	shards, _ := e.enc.Split(e.cache)
	e.enc.Encode(shards)
	for i := range shards {
		e.writers[i].Write(shards[i])
	}
	e.cache = []byte{}
}
func (e *encoder) Writer(p []byte) (n int, err error) {
	length := len(p)
	current := 0
	for length != 0 {
		next := BLOCK_SIZE - len(e.cache)
		if next > length {
			next = length
		}
		e.cache = append(e.cache, p[current:current+next]...)
		if len(e.cache) == BLOCK_SIZE {
			e.Flush()
		}
		current += next
		length -= next
	}
	return len(p), nil
}

func (s *RSPutStream) Commit(success bool) {
	s.Flush()
	for i := range s.writers {
		s.writers[i].(*stream.TempPutStream).Commit(success)
	}
}

type decoder struct {
	readers   []io.Reader
	writers   []io.Writer
	enc       reedsolomon.Encoder
	size      int64
	cache     []byte
	cacheSize int
	total     int64
}

func NewDecoder(readers []io.Reader, writers []io.Writer, size int64) *decoder {
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &decoder{readers, writers, enc, size, nil, 0, 0}
}

type RSGetStream struct {
	*decoder
}

func NewRSGetStream(locateInfo map[int]string, dataServers []string, hash string, size int64) (*RSGetStream, error) {
	if len(locateInfo)+len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("Dataservers number mismatch")
	}
	readers := make([]io.Reader, ALL_SHARDS)
	for i := 0; i < ALL_SHARDS; i++ {
		server := locateInfo[i]
		if server == "" {
			locateInfo[i] = dataServers[0]
			dataServers = dataServers[1:]
			continue
		}
		reader, err := stream.NewGetStream(server, fmt.Sprintf("%s.%d", hash, i))
		if err == nil {
			readers[i] = reader
		}
	}
	writers := make([]io.Writer, ALL_SHARDS)
	perShard := (size + DATA_SHARDS - 1) / DATA_SHARDS
	var err error
	for i := range readers {
		if readers[i] == nil {
			writers[i], err = stream.NewTempPutStream(locateInfo[i], fmt.Sprintf("%s.%d", hash, i), perShard)
			if err != nil {
				return nil, err
			}
		}
	}
	dec := NewDecoder(readers, writers, size)
	return &RSGetStream{dec}, nil
}

func (d *decoder) Read(p []byte) (n int, err error) {
	if d.cacheSize == 0 {
		err := d.getData()
		if err != nil {
			return 0, err
		}
	}

	length := len(p)
	if d.cacheSize < length {
		length = d.cacheSize
	}
	d.cacheSize -= length
	copy(p, d.cache[:length])
	d.cache = d.cache[length:]
	return length, nil
}

func (d *decoder) getData() error {
	if d.total == d.size {
		return io.EOF
	}
	shards := make([][]byte, ALL_SHARDS)
	repairIDs := make([]int, 0)
	for i := range shards {
		if d.readers[i] == nil {
			repairIDs = append(repairIDs, i)
		} else {
			shards[i] = make([]byte, BLOCK_PER_SHARD)
			n, err := io.ReadFull(d.readers[i], shards[i])
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				shards[i] = nil
			} else if n != BLOCK_PER_SHARD {
				shards[i] = shards[i][:n]
			}
		}
	}
	err := d.enc.Reconstruct(shards)
	if err != nil {
		return err
	}
	for i := range repairIDs {
		id := repairIDs[i]
		d.writers[id].Write(shards[id])
	}

	for i := 0; i < DATA_SHARDS; i++ {
		shardSize := int64(len(shards[i]))
		if d.total+shardSize > d.size {
			shardSize -= d.total + shardSize - d.size
		}
		d.cache = append(d.cache, shards[i][:shardSize]...)
		d.cacheSize += int(shardSize)
		d.total += shardSize
	}
	return nil
}

func (s *RSGetStream) Close() {
	for i := range s.writers {
		if s.writers[i] != nil {
			s.writers[i].(*stream.TempPutStream).Commit(true)
		}
	}
}
