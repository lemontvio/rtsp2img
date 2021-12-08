package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lemontvio/rtsp2img/config"
	"github.com/lemontvio/rtsp2img/flag"
	"github.com/lemontvio/rtsp2img/rtsp"
)

var (
	cfg    *config.Config
	logger = log.New(os.Stdout, "", log.LstdFlags)
)

func init() {
	var err error
	if cfg, err = config.New(flag.Parse()); err != nil {
		fmt.Printf("%s %v\n", flag.Parse(), err)
		os.Exit(1)
	}
}

func main() {
	for _, camera := range cfg.Cameras {
		go func(camera config.Camera) {
			logger.Printf("run sn@%v, url@%v\n", camera.Sn, camera.URL)
			for {
				srv := rtsp.New(camera.Sn, camera.URL, cfg.SavePath, cfg.CameraSize, logger)
				go srv.Do()
				select {
				case err := <-srv.Error():
					logger.Printf("sn@%v, err@%v\n", camera.Sn, err)
				}
				<-time.After(time.Second)
			}
		}(camera)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-sig:
		return
	}
}
