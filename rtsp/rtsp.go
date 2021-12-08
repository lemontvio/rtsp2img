package rtsp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aler9/gortsplib"
	"github.com/aler9/gortsplib/pkg/base"
	"github.com/aler9/gortsplib/pkg/rtph264"
	"github.com/pion/rtp"
)

type RTSP struct {
	sn       string
	url      string
	filename string
	duration int
	rtsp     *gortsplib.Client
	pts      time.Duration
	sps      []byte
	pps      []byte
	log      *log.Logger
	error    chan error
	done     chan bool
}

func New(sn string, url string, filename string, duration int, logger *log.Logger) *RTSP {
	return &RTSP{
		sn:       sn,
		url:      url,
		filename: filename,
		duration: duration,
		log:      logger,
		error:    make(chan error, 1),
		done:     make(chan bool, 1),
	}
}

func (r *RTSP) Close() {
	r.done <- true
}

func (r *RTSP) Do() {
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()
	go r.do(ctx)

	select {
	case <-r.done:
		return
	}
}

func (r *RTSP) Error() chan error {
	return r.error
}

func (r *RTSP) do(ctx context.Context) {
	os.MkdirAll(filepath.Dir(r.filename), 0644)

	r.rtsp = &gortsplib.Client{}
	defer r.rtsp.Close()

	go r.request()

	select {
	case <-ctx.Done():
		return
	}
}

func (r *RTSP) request() {
	raw, err := base.ParseURL(r.url)
	if err != nil {
		r.error <- err
		return
	}

	if err := r.rtsp.Start(raw.Scheme, raw.Host); err != nil {
		r.error <- err
		return
	}

	if _, err := r.rtsp.Options(raw); err != nil {
		r.error <- err
		return
	}

	tracks, raw2, _, err := r.rtsp.Describe(raw)
	if err != nil {
		r.error <- err
		return
	}

	var ok bool
	var tid int
	for index, track := range tracks {
		if track.IsH264() {
			conf, err := track.ExtractConfigH264()
			if err != nil {
				r.error <- err
				return
			}

			ok = true
			tid = index
			r.sps = conf.SPS
			r.pps = conf.PPS
			break
		}
	}

	if !ok {
		r.error <- errors.New("h264 track not found")
		return
	}

	dec := rtph264.NewDecoder()
	r.rtsp.OnPacketRTP = func(trackID int, payload []byte) {
		if trackID != tid {
			return
		}

		// parse RTP packet
		var pkt rtp.Packet
		err := pkt.Unmarshal(payload)
		if err != nil {
			return
		}

		// decode H264 NALUs from RTP packets
		nalus, pts, err := dec.DecodeUntilMarker(&pkt)
		if err != nil {
			return
		}

		// print NALUs
		for _, nalu := range nalus {
			switch nalu[0] & 0x1f {
			case 5:
				if pts-r.pts >= time.Millisecond*1000*time.Duration(r.duration) || r.pts < time.Millisecond*1000*time.Duration(r.duration) {
					r.write(nalu)
					r.pts = pts
				}
			}
		}
	}

	if err := r.rtsp.SetupAndPlay(tracks, raw2); err != nil {
		r.error <- err
		return
	}

	r.error <- r.rtsp.Wait()
}

func (r *RTSP) write(nalu []byte) error {
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x1})
	buf.Write(r.sps)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x1})
	buf.Write(r.pps)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x1})
	buf.Write(nalu)

	file := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d_0X%x.h264", r.sn, time.Now().Unix(), &r))
	if err := ioutil.WriteFile(file, buf.Bytes(), 0644); err != nil {
		return err
	}
	defer os.RemoveAll(file)

	outfile := filepath.Join(filepath.Dir(r.filename), strings.ReplaceAll(filepath.Base(r.filename), "sn", r.sn))
	cmd := exec.Command("ffmpeg", "-y", "-i", file, "-vframes", "1", outfile)
	ebuf := bytes.NewBuffer(nil)
	cmd.Stderr = ebuf
	if err := cmd.Run(); err != nil {
		return errors.New(ebuf.String())
	}

	r.log.Printf("write file %v", outfile)

	return nil
}
