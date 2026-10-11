package ytnode

import (
	"bytes"
	"math/rand/v2"
	"testing"

	"github.com/pion/rtp"
)

func makeIPPacket(seq, size int) []byte {
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = byte(seq + i)
	}
	buf[0], buf[1] = byte(seq>>8), byte(seq)
	return buf
}

// packFrames packs packets into frames of framePackets IP packets each and
// returns the resulting RTP packets
func packFrames(t *testing.T, packets [][]byte, framePackets int) []*rtp.Packet {
	t.Helper()

	c := NewRTPConstructor()
	var out []*rtp.Packet
	write := func(p *rtp.Packet) error {
		if len(p.Payload) > RTPMaxPacketSize {
			t.Fatalf("RTP payload too big: %d", len(p.Payload))
		}
		clone := p.Clone()
		out = append(out, clone)
		return nil
	}

	for i := 0; i < len(packets); i += framePackets {
		c.BeginFrame()
		for _, p := range packets[i:min(i+framePackets, len(packets))] {
			c.Add(p)
		}
		if err := c.EndFrame(write); err != nil {
			t.Fatal(err)
		}
	}

	return out
}

func extractAll(t *testing.T, rtpPackets []*rtp.Packet) [][]byte {
	t.Helper()

	e := NewRTPExtractor()
	var got [][]byte
	for _, p := range rtpPackets {
		if !e.Extract(p, func(b []byte) { got = append(got, bytes.Clone(b)) }) {
			t.Fatalf("malformed packet %d", p.SequenceNumber)
		}
	}
	return got
}

func TestRoundTrip(t *testing.T) {
	if err := InitDummyFrames(); err != nil {
		t.Fatal(err)
	}

	for _, framePackets := range []int{1, 3, 50} {
		var packets [][]byte
		for i := range 500 {
			packets = append(packets, makeIPPacket(i, 20+rand.IntN(3000)))
		}

		got := extractAll(t, packFrames(t, packets, framePackets))
		if len(got) != len(packets) {
			t.Fatalf("frame %d: got %d packets, want %d", framePackets, len(got), len(packets))
		}
		for i := range packets {
			if !bytes.Equal(got[i], packets[i]) {
				t.Fatalf("frame %d: packet %d differs", framePackets, i)
			}
		}
	}
}

func TestLossOnlyAffectsContainedPackets(t *testing.T) {
	if err := InitDummyFrames(); err != nil {
		t.Fatal(err)
	}

	var packets [][]byte
	for i := range 2000 {
		packets = append(packets, makeIPPacket(i, 20+rand.IntN(1500)))
	}
	rtpPackets := packFrames(t, packets, 20)

	var kept []*rtp.Packet
	for _, p := range rtpPackets {
		if rand.IntN(10) != 0 {
			kept = append(kept, p)
		}
	}

	got := extractAll(t, kept)
	want := make(map[string]bool, len(packets))
	for _, p := range packets {
		want[string(p)] = true
	}
	for _, g := range got {
		if !want[string(g)] {
			t.Fatalf("got a corrupted packet of length %d", len(g))
		}
	}

	// ~10% of RTP packets lost, so most of the IP packets must survive
	if len(got) < len(packets)*7/10 {
		t.Fatalf("too many packets lost: got %d of %d", len(got), len(packets))
	}
	t.Logf("got %d of %d IP packets after losing %d of %d RTP packets",
		len(got), len(packets), len(rtpPackets)-len(kept), len(rtpPackets))
}

func TestSmallPacketsFitOneRTPPacket(t *testing.T) {
	packets := [][]byte{makeIPPacket(0, MaxTunnelMTU)}
	if n := len(packFrames(t, packets, 1)); n != 1 {
		t.Fatalf("MTU-sized packet took %d RTP packets", n)
	}
}

func TestDefaultMTUFillsTwoRTPPackets(t *testing.T) {
	packets := [][]byte{makeIPPacket(0, DefaultTunnelMTU)}
	rtpPackets := packFrames(t, packets, 1)
	if len(rtpPackets) != 2 {
		t.Fatalf("MTU-sized packet took %d RTP packets", len(rtpPackets))
	}
	for _, p := range rtpPackets {
		if len(p.Payload) != RTPMaxPacketSize {
			t.Fatalf("RTP packet isn't full: %d bytes", len(p.Payload))
		}
	}

	got := extractAll(t, rtpPackets)
	if len(got) != 1 || !bytes.Equal(got[0], packets[0]) {
		t.Fatal("packet didn't survive the round trip")
	}
}
