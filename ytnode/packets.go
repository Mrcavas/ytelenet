package ytnode

import (
	"bytes"
	_ "embed"
	"errors"
	"io"
	"iter"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4/pkg/media/ivfreader"
)

const RTPMaxPacketSize = 1200

//go:embed dummy.ivf
var dummyIVFBytes []byte

var DummyFrames [][]byte = nil

func InitDummyFrames() error {
	if DummyFrames != nil {
		return nil
	}

	reader := bytes.NewReader(dummyIVFBytes)
	ivf, _, err := ivfreader.NewWith(reader)
	if err != nil {
		return err
	}

	for {
		frameBytes, _, err := ivf.ParseNextFrame()
		if errors.Is(err, io.EOF) {
			break // End of animation
		}
		if err != nil {
			return err
		}

		frameCopy := make([]byte, len(frameBytes))
		copy(frameCopy, frameBytes)

		DummyFrames = append(DummyFrames, frameCopy)
	}

	if len(DummyFrames) == 0 {
		return errors.New("no frames found in embedded dummy.ivf")
	}

	return nil
}

type RTPConstructor struct {
	startTime     time.Time
	picCount      uint32
	seqNumber     uint16
	payloadBuf    [RTPMaxPacketSize]byte
	lastTimestamp uint32
	p             rtp.Packet
}

func NewRTPConstructor() *RTPConstructor {
	return &RTPConstructor{
		startTime: time.Now(),
		p: rtp.Packet{
			Header: rtp.Header{
				Version:     2,
				PayloadType: 96,
			},
		},
	}
}

func (c *RTPConstructor) NewTimestamp() uint32 {
	// timestamp := c.lastTimestamp + 1
	timestamp := uint32(90_000 * time.Since(c.startTime).Seconds())
	if c.lastTimestamp == timestamp {
		timestamp++
	}

	c.lastTimestamp = timestamp
	return timestamp
}

func (c *RTPConstructor) NewPackets(buf []byte) iter.Seq[*rtp.Packet] {
	return func(yield func(*rtp.Packet) bool) {
		timestamp := c.NewTimestamp()
		c.picCount++
		picId := c.picCount & 0x7FFF

		remaining := buf
		isFirst := true

		for len(remaining) > 0 {
			chunkSize := len(remaining)
			headerLen := 4

			if isFirst {
				headerLen += 10
			}

			if chunkSize+headerLen > RTPMaxPacketSize {
				chunkSize = RTPMaxPacketSize - headerLen
			}

			chunk := remaining[:chunkSize]
			remaining = remaining[chunkSize:]

			isLast := len(remaining) == 0

			c.p.Payload = c.payloadBuf[:headerLen+len(chunk)]
			c.FillPayloadDescriptor(picId, isFirst)
			if isFirst {
				c.FillPayloadHeader()
			}

			copy(c.p.Payload[headerLen:], chunk)

			c.seqNumber++
			c.p.Marker = isLast
			c.p.SequenceNumber = c.seqNumber
			c.p.Timestamp = timestamp

			if !yield(&c.p) {
				return
			}

			isFirst = false
		}
	}
}

func (c *RTPConstructor) FillPayloadDescriptor(picId uint32, start bool) {
	// --- VP8 RTP Payload Descriptor ---
	c.p.Payload[0] = 0b10000000 // X=1 (расширения)
	if start {
		c.p.Payload[0] = 0b10010000 // X=1, S=1 (начало партиции)
	}
	c.p.Payload[1] = 0b10000000 // I=1 (присутствует PictureID)
	c.p.Payload[2] = 0x80 | byte((picId>>8)&0x7F)
	c.p.Payload[3] = byte(picId & 0xFF)
}

func (c *RTPConstructor) FillPayloadHeader() {
	// --- VP8 Payload Header ---
	c.p.Payload[4] = 0b00010000 // Бит 0 = 0 (Это Keyframe), ShowFrame = 1
	c.p.Payload[5] = 0x00       // Size
	c.p.Payload[6] = 0x00       // Size

	// --- VP8 Sync Code ---
	c.p.Payload[7] = 0x9D
	c.p.Payload[8] = 0x01
	c.p.Payload[9] = 0x2A

	// --- Resolution (LE) ---
	c.p.Payload[10] = 0x80
	c.p.Payload[11] = 0x02 // Width = 640 (0x0280)
	c.p.Payload[12] = 0xE0
	c.p.Payload[13] = 0x01 // Height = 480 (0x01E0)
}

func (c *RTPConstructor) NewDummyPacket() *rtp.Packet {
	c.picCount++
	picId := c.picCount & 0x7FFF

	frameBytes := DummyFrames[c.picCount%uint32(len(DummyFrames))]

	totalLen := 4 + len(frameBytes) // 4 (Descriptor) + colored frame
	c.p.Payload = c.payloadBuf[:totalLen]

	c.p.Payload[0] = 0b10010000
	c.p.Payload[1] = 0b10000000
	c.p.Payload[2] = 0x80 | byte((picId>>8)&0x7F)
	c.p.Payload[3] = byte(picId & 0xFF)

	copy(c.p.Payload[4:], frameBytes)

	c.seqNumber++
	c.p.Marker = true
	c.p.SequenceNumber = c.seqNumber
	c.p.Timestamp = c.NewTimestamp()

	return &c.p
}

type RTPExtractor struct {
	incomplete map[uint16]*IncompletePackets
}

type IncompletePackets struct {
	startSeqNumber *uint16
	endSeqNumber   *uint16
	bufs           map[uint16][]byte
}

func NewRTPExtractor() *RTPExtractor {
	return &RTPExtractor{
		incomplete: make(map[uint16]*IncompletePackets),
	}
}

func (e *RTPExtractor) Extract(packet *rtp.Packet) ([]byte, bool) {
	payload := packet.Payload
	if len(payload) < 1 {
		return nil, false
	}

	// Разбор VP8 Payload Descriptor
	var picId uint16
	offset := 1

	if (payload[0] & 0x80) != 0 { // Бит X (Extension present)
		if offset >= len(payload) {
			return nil, false
		}
		ext := payload[offset]
		offset++

		if (ext & 0x80) != 0 { // Бит I (PictureID present)
			if offset >= len(payload) {
				return nil, false
			}
			if (payload[offset] & 0x80) != 0 {
				picId = uint16(payload[offset+1]) | (uint16(payload[offset]&0x7F) << 8)
				offset += 2 // 16-bit PictureID
			} else {
				picId = uint16(payload[offset] & 0x7F)
				offset += 1 // 8-bit PictureID
			}
		}
		if (ext & 0x40) != 0 { // Бит L
			offset++
		}
		if (ext&0x20) != 0 || (ext&0x10) != 0 { // Биты T или K
			offset++
		}
	}
	isStart := (payload[0] & 0x10) != 0

	if offset >= len(payload) {
		return nil, false
	}

	if packet.Marker && isStart { // это начало и конец (т.е. один пакет)
		vp8Payload := payload[offset:]

		for _, dummyFrame := range DummyFrames {
			if bytes.Equal(vp8Payload, dummyFrame) {
				// это пустышка, дропаем.
				return []byte{}, true
			}
		}

		// VP8 Payload Header. LSB первого байта - флаг типа кадра (0 = Keyframe)
		isKeyframe := (payload[offset] & 0x01) == 0
		if !isKeyframe {
			return []byte{}, true
		}

		offset += 10
		if offset > len(payload) {
			return nil, false
		}

		return payload[offset:], true
	}

	inc, ok := e.incomplete[picId]
	if !ok {
		inc = &IncompletePackets{
			bufs: make(map[uint16][]byte, 2),
		}
		e.incomplete[picId] = inc
	}

	if isStart {
		offset += 10 // в стартовом пакете в любом случае ожидаем Header
		inc.startSeqNumber = &packet.SequenceNumber
	}

	if packet.Marker {
		inc.endSeqNumber = &packet.SequenceNumber
	}

	inc.bufs[packet.SequenceNumber] = payload[offset:]

	// проверяем, дошли ли все пакеты
	if inc.endSeqNumber == nil ||
		inc.startSeqNumber == nil ||
		len(inc.bufs) != int(1+*inc.endSeqNumber-*inc.startSeqNumber) {
		return []byte{}, true
	}

	// если всё дошло, объединяем
	defer delete(e.incomplete, picId)

	completeLen := 0
	for _, v := range inc.bufs {
		completeLen += len(v)
	}

	completeBuf := make([]byte, completeLen)
	currentStart := 0
	for i := *inc.startSeqNumber; i <= *inc.endSeqNumber; i++ {
		buf, ok := inc.bufs[i]
		if !ok {
			return nil, false
		}
		currentStart += copy(completeBuf[currentStart:], buf)
	}

	return completeBuf, true
}
