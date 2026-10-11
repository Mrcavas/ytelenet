package ytnode

import (
	"bytes"
	_ "embed"
	"errors"
	"io"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4/pkg/media/ivfreader"
)

// RTPMaxPacketSize is the max RTP payload size (same as browsers use, safe for
// mobile networks with small path MTU)
const RTPMaxPacketSize = 1200

const (
	vp8DescriptorLen = 4
	vp8HeaderLen     = 10

	// every IP packet (or a fragment of it) inside an RTP payload is prefixed
	// with a 2-byte chunk header: 2 flag bits + 14 bits of length
	chunkHeaderLen  = 2
	chunkMore       = 0x8000 // the packet continues in the next RTP packet
	chunkContinued  = 0x4000 // this chunk continues a packet from the previous RTP packet
	chunkLengthMask = 0x3FFF

	// don't start a chunk in an RTP packet with less space than that, fragmenting
	// a packet into a tiny piece isn't worth it
	minChunkSpace = 64
)

// MaxTunnelMTU is the largest IP packet that always fits into a single RTP
// packet. Bigger packets are fragmented, and losing any fragment loses the
// packet.
const MaxTunnelMTU = RTPMaxPacketSize - vp8DescriptorLen - vp8HeaderLen - chunkHeaderLen

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

// RTPConstructor packs IP packets into fake VP8 keyframes.
//
// Unlike a real VP8 frame, every RTP packet of the frame can be parsed on its
// own: it contains whole IP packets (and at most one fragment at each end).
// The SFU tail-drops packets that exceed its bandwidth allocation, so this way
// a dropped RTP packet costs only the IP packets inside it, not the whole
// frame.
type RTPConstructor struct {
	startTime     time.Time
	picCount      uint32
	seqNumber     uint16
	lastTimestamp uint32
	p             rtp.Packet

	// payloads of the frame being built, the last one is being filled
	payloads [][]byte
	picId    uint32
	inFrame  bool
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
	timestamp := uint32(90_000 * time.Since(c.startTime).Seconds())
	if timestamp <= c.lastTimestamp {
		timestamp = c.lastTimestamp + 1
	}

	c.lastTimestamp = timestamp
	return timestamp
}

// FramePackets returns the amount of RTP packets in the current frame
// (including the one being filled)
func (c *RTPConstructor) FramePackets() int {
	return len(c.payloads)
}

func (c *RTPConstructor) BeginFrame() {
	c.picCount++
	c.picId = c.picCount & 0x7FFF
	c.inFrame = true
	c.payloads = c.payloads[:0]
	c.startPacket()
}

func (c *RTPConstructor) startPacket() {
	first := len(c.payloads) == 0

	// reuse the buffers of the previous frames
	var buf []byte
	if len(c.payloads) < cap(c.payloads) {
		buf = c.payloads[:len(c.payloads)+1][len(c.payloads)][:0]
	}
	if buf == nil {
		buf = make([]byte, 0, RTPMaxPacketSize)
	}

	buf = buf[:vp8DescriptorLen]
	c.FillPayloadDescriptor(buf, c.picId, first)
	if first {
		buf = buf[:vp8DescriptorLen+vp8HeaderLen]
		c.FillPayloadHeader(buf)
	}
	c.payloads = append(c.payloads, buf)
}

func (c *RTPConstructor) current() *[]byte {
	return &c.payloads[len(c.payloads)-1]
}

func (c *RTPConstructor) appendChunk(header uint16, chunk []byte) {
	cur := c.current()
	*cur = append(*cur, byte(header>>8), byte(header))
	*cur = append(*cur, chunk...)
}

// Add appends an IP packet to the current frame
func (c *RTPConstructor) Add(buf []byte) {
	continued := false

	for {
		space := RTPMaxPacketSize - len(*c.current()) - chunkHeaderLen
		if space < len(buf) && space < minChunkSpace {
			c.startPacket()
			continue
		}

		chunk := buf[:min(space, len(buf))]
		buf = buf[len(chunk):]

		header := uint16(len(chunk))
		if continued {
			header |= chunkContinued
		}
		if len(buf) > 0 {
			header |= chunkMore
		}

		c.appendChunk(header, chunk)

		if len(buf) == 0 {
			return
		}

		// the rest goes to the next RTP packet
		c.startPacket()
		continued = true
	}
}

// EndFrame writes out the RTP packets of the frame
func (c *RTPConstructor) EndFrame(write func(*rtp.Packet) error) error {
	if !c.inFrame {
		return nil
	}
	c.inFrame = false

	timestamp := c.NewTimestamp()
	for i, payload := range c.payloads {
		c.seqNumber++
		c.p.Payload = payload
		c.p.Marker = i == len(c.payloads)-1
		c.p.SequenceNumber = c.seqNumber
		c.p.Timestamp = timestamp

		if err := write(&c.p); err != nil {
			return err
		}
	}

	return nil
}

func (c *RTPConstructor) FillPayloadDescriptor(buf []byte, picId uint32, start bool) {
	// --- VP8 RTP Payload Descriptor ---
	buf[0] = 0b10000000 // X=1 (расширения)
	if start {
		buf[0] = 0b10010000 // X=1, S=1 (начало партиции)
	}
	buf[1] = 0b10000000 // I=1 (присутствует PictureID)
	buf[2] = 0x80 | byte((picId>>8)&0x7F)
	buf[3] = byte(picId & 0xFF)
}

func (c *RTPConstructor) FillPayloadHeader(buf []byte) {
	// --- VP8 Payload Header ---
	buf[4] = 0b00010000 // Бит 0 = 0 (Это Keyframe), ShowFrame = 1
	buf[5] = 0x00       // Size
	buf[6] = 0x00       // Size

	// --- VP8 Sync Code ---
	buf[7] = 0x9D
	buf[8] = 0x01
	buf[9] = 0x2A

	// --- Resolution (LE) ---
	buf[10] = 0x80
	buf[11] = 0x07 // Width = 1920 (0x0780)
	buf[12] = 0x38
	buf[13] = 0x04 // Height = 1080 (0x0438)
}

func (c *RTPConstructor) NewDummyPacket() *rtp.Packet {
	c.picCount++
	picId := c.picCount & 0x7FFF

	frameBytes := DummyFrames[c.picCount%uint32(len(DummyFrames))]

	c.p.Payload = make([]byte, 4+len(frameBytes)) // 4 (Descriptor) + colored frame

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

// RTPExtractor parses IP packets out of RTP packets made by RTPConstructor
type RTPExtractor struct {
	// IP packet fragment from the previous RTP packet
	pending    []byte
	pendingSeq uint16
	hasPending bool
}

func NewRTPExtractor() *RTPExtractor {
	return &RTPExtractor{}
}

// Extract calls emit for every complete IP packet in the RTP packet. The buffer
// passed to emit is only valid during the call. Returns false if the packet is
// malformed.
func (e *RTPExtractor) Extract(packet *rtp.Packet, emit func([]byte)) bool {
	payload := packet.Payload
	if len(payload) < 1 {
		return false
	}

	// Разбор VP8 Payload Descriptor
	offset := 1

	if (payload[0] & 0b10000000) != 0 { // Бит X (Extension present)
		if offset >= len(payload) {
			return false
		}
		ext := payload[offset]
		offset++

		if (ext & 0b10000000) != 0 { // Бит I (PictureID present)
			if offset >= len(payload) {
				return false
			}
			if (payload[offset] & 0b10000000) != 0 {
				offset += 2 // 16-bit PictureID
			} else {
				offset += 1 // 8-bit PictureID
			}
		}
		if (ext & 0b01000000) != 0 { // Бит L
			offset++
		}
		if (ext&0b00100000) != 0 || (ext&0b00010000) != 0 { // Биты T или K
			offset++
		}
	}

	if offset >= len(payload) {
		return false
	}

	isStart := (payload[0] & 0b00010000) != 0

	if isStart {
		if packet.Marker { // это начало и конец (т.е. один пакет)
			vp8Payload := payload[offset:]

			for _, dummyFrame := range DummyFrames {
				if bytes.Equal(vp8Payload, dummyFrame) {
					// это пустышка, дропаем.
					return true
				}
			}

			// VP8 Payload Header. LSB первого байта - флаг типа кадра (0 = Keyframe)
			if (payload[offset] & 0b00000001) != 0 {
				return true
			}
		}

		offset += vp8HeaderLen // в стартовом пакете в любом случае ожидаем Header
		if offset > len(payload) {
			return false
		}
	}

	pendingOk := e.hasPending && packet.SequenceNumber == e.pendingSeq+1
	e.hasPending = false

	data := payload[offset:]
	for len(data) > 0 {
		if len(data) < chunkHeaderLen {
			return false
		}

		header := uint16(data[0])<<8 | uint16(data[1])
		length := int(header & chunkLengthMask)
		data = data[chunkHeaderLen:]

		if len(data) < length {
			return false
		}
		chunk := data[:length]
		data = data[length:]

		if header&chunkContinued != 0 {
			if !pendingOk {
				// the beginning of the packet was lost
				continue
			}
			pendingOk = false
			chunk = append(e.pending, chunk...)
		}

		if header&chunkMore != 0 {
			e.pending = append(e.pending[:0], chunk...)
			e.pendingSeq = packet.SequenceNumber
			e.hasPending = true
			continue
		}

		emit(chunk)
	}

	return true
}
