package core

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/glutechnologies/glubng/pkg/kea"
)

type Config struct {
	SrcKeaSocket string
	SrcVppSocket string
}

func LoadConfig(config *Config) {
	// Load configuration
	configSrc := flag.String("config", "/etc/glubng.toml", "Config source path")
	flag.Parse()

	body, err := ioutil.ReadFile(*configSrc)

	if err != nil {
		log.Fatalf("Error loading configuration file")
	}

	_, err = toml.Decode(string(body), &config)

	if err != nil {
		log.Fatalf("Error decoding configuration file")
	}
}

func Init() {
	// Load initial configuration
	var config Config
	LoadConfig(&config)

	k := &kea.KeaSocket{}

	// Init kea listener
	k.Init(config.SrcKeaSocket)

	// Create a channel to process signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to exit when everything has stopped gracefully
	done := make(chan bool, 1)

	fmt.Println("Running GluBNGd...")

	go func() {
		<-c
		k.Close()
		done <- true
	}()
	<-done
	fmt.Println("Exiting GluBNGd...")
}
