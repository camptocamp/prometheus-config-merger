package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	"github.com/jessevdk/go-flags"
)

var version = "undefined"

type Config struct {
	Version bool   `short:"V" long:"version" description:"Display version."`
	Sleep   string `short:"s" long:"sleep" description:"Sleep time between queries." env:"PROMETHEUS_CONFIG_MERGER_SLEEP" default:"5s"`
	Manpage bool   `short:"m" long:"manpage" description:"Output manpage."`
}

/*
NOTE: adapted from https://play.golang.org/p/8jlJUbEJKf to support merging slices
*/

// merge merges the two JSON-marshalable values x1 and x2,
// preferring x1 over x2 except where x1 and x2 are
// JSON objects, in which case the keys from both objects
// are included and their values merged recursively.
//
// It returns an error if x1 or x2 cannot be JSON-marshaled.
func merge(x1, x2 interface{}) (interface{}, error) {
	data1, err := json.Marshal(x1)
	if err != nil {
		return nil, err
	}
	data2, err := json.Marshal(x2)
	if err != nil {
		return nil, err
	}
	var j1 interface{}
	err = json.Unmarshal(data1, &j1)
	if err != nil {
		return nil, err
	}
	var j2 interface{}
	err = json.Unmarshal(data2, &j2)
	if err != nil {
		return nil, err
	}
	return merge1(j1, j2), nil
}

func merge1(x1, x2 interface{}) interface{} {
	switch x1 := x1.(type) {
	case map[string]interface{}:
		x2, ok := x2.(map[string]interface{})
		if !ok {
			return x1
		}
		for k, v2 := range x2 {
			if v1, ok := x1[k]; ok {
				x1[k] = merge1(v1, v2)
			} else {
				x1[k] = v2
			}
		}
	case []interface{}:
		x2, ok := x2.([]interface{})
		if !ok {
			return x1
		}
		for i := range x2 {
			x1 = append(x1, x2[i])
		}
		return x1
	case nil:
		// merge(nil, map[string]interface{...}) -> map[string]interface{...}
		x2, ok := x2.(map[string]interface{})
		if ok {
			return x2
		}
	}
	return x1
}

func main() {
	cfg, err := loadConfig(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sleep, err := time.ParseDuration(cfg.Sleep)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		var config interface{}

		files, _ := filepath.Glob("/etc/prometheus/conf.d/*.yml")
		for i := range files {
			// Read
			raw, err := ioutil.ReadFile(files[i])
			if err != nil {
				fmt.Println(err.Error())
			}

			// Unmarshal
			var c map[string]interface{}
			err = yaml.Unmarshal(raw, &c)
			if err != nil {
				fmt.Println(err.Error())
			}

			// Merge
			config, err = merge(c, config)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		data, err := yaml.Marshal(config)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("%s\n", data)
		fmt.Printf("Sleeping for %v\n", sleep)
		time.Sleep(sleep)
	}
}

func loadConfig(version string) (c Config, err error) {
	parser := flags.NewParser(&c, flags.Default)
	_, err = parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if c.Version {
		fmt.Printf("Prometheus-config-merger v%v\n", version)
		os.Exit(0)
	}

	if c.Manpage {
		var buf bytes.Buffer
		parser.ShortDescription = "Prometheus configuration merger"
		parser.WriteManPage(&buf)
		fmt.Printf(buf.String())
		os.Exit(0)
	}
	return
}
