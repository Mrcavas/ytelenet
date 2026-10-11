// probe measures what the Telemost SFU does to our fake-VP8 data stream:
// throughput, loss and one-way latency (requires NTP-synced clocks).
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"
	"ytelenet/ytnode"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
)

type probeArgs struct {
	Room   string  `arg:"--room,required"`
	Name   string  `arg:"--name,required"`
	To     string  `arg:"--to,required"`
	Rate   float64 `arg:"--rate" default:"0" help:"send rate, Mbit/s (0 = only receive)"`
	Size   int     `arg:"--size" default:"1200" help:"payload size"`
	Dur    int     `arg:"--dur" default:"20" help:"send duration, seconds"`
	Pause  int     `arg:"--pause" default:"0" help:"pause between two send phases, seconds (0 = one phase)"`
	Burst  int     `arg:"--burst" default:"1" help:"payloads sent back-to-back at once"`
	Delay  int     `arg:"--delay" default:"4" help:"seconds to wait before sending"`
	Phases string  `arg:"--phases" help:"rate:seconds list, e.g. 30:40,0:120,30:15 (overrides rate/dur/pause)"`
	Batch  int     `arg:"-b,--batch" default:"64"`
	Debug  bool    `arg:"--debug"`
}

const magic = 0x59544E50 // "YTNP"

func main() {
	var a probeArgs
	arg.MustParse(&a)

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: "15:04:05.000"})
	internal := log.New()
	internal.SetFormatter(log.StandardLogger().Formatter)
	internal.SetLevel(log.InfoLevel)
	if a.Debug {
		internal.SetLevel(log.DebugLevel)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	node, err := ytnode.MakeNew(
		internal, "https://telemost.yandex.ru/j/"+a.Room, a.Name, a.To,
		ytnode.BatchingOptions{MaxBatchSize: a.Batch},
	)
	if err != nil {
		log.Fatalf("make node: %v", err)
	}
	defer node.Stop()

	if st := <-node.Events(); st != ytnode.ConnectedState {
		log.Fatalf("not connected")
	}
	log.Infof("connected")

	if a.Rate > 0 || a.Phases != "" {
		go sender(node, a)
	}

	stats := newStats()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case buf := <-node.Data():
			stats.add(buf)
		case <-tick.C:
			stats.report()
		case st := <-node.Events():
			if st == ytnode.StoppedState {
				log.Warnf("node stopped")
				return
			}
		case <-interrupt:
			stats.total()
			return
		}
	}
}

func sender(node *ytnode.Node, a probeArgs) {
	// give the SFU some time to set up the subscriber side of the target
	time.Sleep(time.Duration(a.Delay) * time.Second)

	type phase struct {
		rate float64
		dur  int
	}
	var phases []phase
	if a.Phases != "" {
		for _, p := range strings.Split(a.Phases, ",") {
			var ph phase
			if _, err := fmt.Sscanf(p, "%g:%d", &ph.rate, &ph.dur); err != nil {
				log.Fatalf("bad phase %q", p)
			}
			phases = append(phases, ph)
		}
	} else {
		phases = append(phases, phase{a.Rate, a.Dur})
		if a.Pause > 0 {
			phases = append(phases, phase{0, a.Pause}, phase{a.Rate, a.Dur})
		}
	}

	var seq uint64
	for _, ph := range phases {
		log.Infof("phase: %.2f Mbit/s for %ds", ph.rate, ph.dur)
		if ph.rate <= 0 {
			time.Sleep(time.Duration(ph.dur) * time.Second)
			continue
		}
		seq = sendPhase(node, a, ph.rate, ph.dur, seq)
	}
	log.Infof("sender done: %d payloads", seq)
}

func sendPhase(node *ytnode.Node, a probeArgs, rate float64, dur int, seq uint64) uint64 {
	pps := rate * 1e6 / 8 / float64(a.Size)
	interval := time.Duration(float64(time.Second) / pps * float64(a.Burst))

	start := time.Now()
	next := start
	for time.Since(start) < time.Duration(dur)*time.Second {
		now := time.Now()
		for !next.After(now) {
			for range a.Burst {
				buf := make([]byte, a.Size)
				binary.BigEndian.PutUint32(buf[0:], magic)
				binary.BigEndian.PutUint64(buf[4:], seq)
				binary.BigEndian.PutUint64(buf[12:], uint64(time.Now().UnixNano()))
				node.Send(buf)
				seq++
			}
			next = next.Add(interval)
		}
		time.Sleep(time.Until(next))
	}
	return seq
}

type stats struct {
	lat         []float64
	bytes       int
	count       int
	maxSeq      int64
	minSeq      int64
	seen        map[uint64]struct{}
	totalBytes  int
	totalCount  int
	totalLat    []float64
	first, last time.Time
}

func newStats() *stats {
	return &stats{maxSeq: -1, minSeq: -1, seen: make(map[uint64]struct{})}
}

func (s *stats) add(buf []byte) {
	if len(buf) < 20 || binary.BigEndian.Uint32(buf) != magic {
		return
	}
	seq := binary.BigEndian.Uint64(buf[4:])
	sent := time.Unix(0, int64(binary.BigEndian.Uint64(buf[12:])))
	now := time.Now()
	if s.first.IsZero() {
		s.first = now
	}
	s.last = now

	l := float64(now.Sub(sent).Microseconds()) / 1000
	s.lat = append(s.lat, l)
	s.totalLat = append(s.totalLat, l)
	s.bytes += len(buf)
	s.count++
	s.totalBytes += len(buf)
	s.totalCount++
	s.seen[seq] = struct{}{}
	if s.minSeq < 0 || int64(seq) < s.minSeq {
		s.minSeq = int64(seq)
	}
	if int64(seq) > s.maxSeq {
		s.maxSeq = int64(seq)
	}
}

func pct(v []float64, p float64) float64 {
	if len(v) == 0 {
		return 0
	}
	return v[int(float64(len(v)-1)*p)]
}

func (s *stats) report() {
	if s.count == 0 {
		return
	}
	slices.Sort(s.lat)
	fmt.Printf(
		"%s rx %6.2f Mbit/s %5d pkts | lat p50 %6.0f p95 %6.0f max %6.0f ms | lost so far %d\n",
		time.Now().Format("15:04:05"), float64(s.bytes)*8/1e6, s.count,
		pct(s.lat, .5), pct(s.lat, .95), pct(s.lat, 1),
		s.maxSeq-s.minSeq+1-int64(len(s.seen)),
	)
	s.lat = s.lat[:0]
	s.bytes, s.count = 0, 0
}

func (s *stats) total() {
	if s.totalCount == 0 {
		fmt.Println("TOTAL: nothing received")
		return
	}
	slices.Sort(s.totalLat)
	dur := s.last.Sub(s.first).Seconds()
	fmt.Printf(
		"TOTAL: %d pkts, %.2f MB in %.1fs = %.2f Mbit/s | loss %.2f%% | lat p50 %.0f p95 %.0f ms\n",
		s.totalCount, float64(s.totalBytes)/1e6, dur, float64(s.totalBytes)*8/1e6/dur,
		100*float64(s.maxSeq-s.minSeq+1-int64(len(s.seen)))/float64(s.maxSeq-s.minSeq+1),
		pct(s.totalLat, .5), pct(s.totalLat, .95),
	)
}
