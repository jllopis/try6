// Package getconf load the variables to be used in a program from different sources:
//   1. Environment variables
//   2. command line
//   3. etcd server
// This is also the precedence order. In case that etcd watching is required, you can watch
// the variables for changes. When a change event is notified, the config struct will be
// updated and the program notified via channel.
// The package know about the options by way of a struct definition that must be passed.
// Example:
//
//    type Config struct {
//        key1 int,
//        key2 string
//    }
//    config := getconf.New(&Config{}, "", "")
//    fmt.Printf("Key1 = %d\nKey2 = %s\n", config.Get(key1), config.GetString(key2))
//
// The default names and behaviour can be modified by the use of defined tags in the variable
// declaration. This way you can state if a var have to be watched for changes in etcd or
// if must be ignored for example. Also it is possible to define the names to look for.
// As the package parse the command line options, it should be called at the program start,
// it must be the first action to call.
package getconf

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/coreos/go-etcd/etcd"
)

const (
	VERSION = "0.3.1"
)

func Version() string {
	return VERSION
}

var (
	etcdConn      *etcd.Client
	responseChan  chan *etcd.Response
	watchStopChan chan bool
	wg            sync.WaitGroup
)

type IncompatibleTypes struct {
	msg   string
	baset reflect.Kind
	valt  reflect.Kind
}

func (e IncompatibleTypes) Error() string {
	return fmt.Sprintf("%s. Got: %v  Expected: %v", e.msg, e.valt, e.baset)
}

// Option holds the data related to a variable as its type, value and the names
// it will take in the different enviroments: command line flags, environment and etcd
type Option struct {
	name         string
	envName      string
	etcdName     string
	flagName     string
	value        interface{}
	oType        reflect.Kind
	defaultValue interface{}
	noetcd       bool
}

// GetConf is the struct providing access to the Option slice and holds global variables for GetConf
type GetConf struct {
	// Option struct slice
	allOptions map[string]*Option
	// Prefix for the etcd and environment variables
	envPrefix string
	// Watch etcd server for keys
	watches []map[string]string
	// Command line has been parsed
	clParsed bool
	// Number of options managed
	nKeys int
	// tell us if we should connect to an etcd server
	connectEtcd bool
	// URL for the etcd server. Defaults to localhost
	etcdURL string
	// Channel to get updates from watched vars in etcd
	ConfChanged chan map[string]interface{}
}

// New create an instance of GetConf struct that will
// contain the keys and values gathered from the different
// configuration options
func New(c interface{}, prefix string, useEtcd bool, etcdURL string) *GetConf {
	defer dontPanic()
	elem := reflect.ValueOf(c).Elem()
	if reflect.Struct != elem.Kind() {
		log.Printf("getconf.new: ERROR param must be a struct holding configuration keys")
		return nil
	}

	config, err := parse(elem)
	if err != nil {
		log.Printf("getconf.new: error parsing options -> %s", err)
	}

	err = config.getEnv(prefix)
	if err != nil {
		log.Printf("getconf.new: error parsing ENV options -> %s", err)
	}

	err = config.getFlags()
	if err != nil {
		log.Printf("getconf.new: error parsing FLAGS options -> %s", err)
	}

	if useEtcd {
		config.connectEtcd = true
		if etcdURL != "" {
			config.etcdURL = etcdURL
		} else {
			config.etcdURL = "http://localhost:4001"
		}
		err = config.getEtcd()
		if err != nil {
			log.Printf("[ERR] getconf.new: error getting etcd options -> %s", err)
		}
	}

	return config
}

// parse will grab the options defined and creates the default GetConf. It fills the metadata for the Options
func parse(e reflect.Value) (*GetConf, error) {
	var w []map[string]string
	gc := &GetConf{allOptions: make(map[string]*Option), watches: w, clParsed: false, nKeys: e.NumField(), ConfChanged: make(chan map[string]interface{})}
	for i := 0; i < e.NumField(); i++ {
		f := e.Type().Field(i)
		o := &Option{name: f.Name,
			oType: e.Field(i).Kind(),
		}
		tag := f.Tag
		if t := tag.Get("getconf"); t != "" {
			err := parseTags(o, t)
			if err != nil && err.Error() == "untrack" {
				log.Printf("getconf.parse: option %s not tracked!", o.name)
				continue
			}
		}
		gc.allOptions[f.Name] = o
	}
	return gc, nil
}

// parseTags read the tags and set the corresponding variables in the Option struct
func parseTags(o *Option, t string) error {
	for _, k := range strings.Split(t, ",") {
		if strings.TrimSpace(k) == "-" {
			return errors.New("untrack")
		}
		switch strings.Fields(k)[0] {
		case "etcd":
			// All vars to be watched on etcd must have a name specified as a tag. Otherwise they will be ignored
			// They also get trimmed the initial slash if starts with it
			o.etcdName = strings.TrimLeft(strings.Fields(k)[1], "/")
		case "env":
			// if there is a tag specifying the var name in the environment, use it. If not, take prefix + '_' + ToUpper(varname)
			o.envName = strings.Fields(k)[1]
		case "flag":
			o.flagName = strings.Fields(k)[1]
		case "noetcd":
			o.noetcd = true
		}
	}
	return nil
}

// getEnv will get the variables found in the environment and assign the value to the Option
func (c *GetConf) getEnv(prefix string) error {
	// get env vars. If not specified in a tag, the values must be uppercase to prevent mixed case errors. They also must have the prefix applied and separated
	// from the key by an underscore
	for _, v := range c.allOptions {
		if v.envName == "" {
			v.envName = strings.ToUpper(prefix + "_" + v.name)
		}
		if val := os.Getenv(v.envName); val != "" {
			c.Set(v, val)
		}
	}
	return nil
}

// getFlags will define the variables that can be found in the command line (via the flags package)
func (c *GetConf) getFlags() error {
	// parse command line
	if flag.Parsed() {
		return errors.New("getconf.getFlags: flags already parsed")
	} else {
		//fl := flag.NewFlagSet("flag", flag.ContinueOnError)
		for _, v := range c.allOptions {
			if v.flagName == "" {
				v.flagName = strings.ToLower(v.name)
			}
			switch v.oType {
			case reflect.Int:
				flag.Int(v.flagName, 0, "")
			case reflect.Int64:
				flag.Int64(v.flagName, 0, "")
			case reflect.Float64:
				flag.Float64(v.flagName, 0, "")
			case reflect.Bool:
				flag.Bool(v.flagName, false, "")
			case reflect.String:
				flag.String(v.flagName, "", "")
			}
		}
	}
	return nil
}

// Parse call flag.Parse() and get the command line flags into the appropiate variables.
// It is respectful with other packages by not calling flag.Parse() by default to prevent collisions.
// If this function is called, it must be INSTEAD of flag.Parse()
func (c *GetConf) Parse() {
	flag.Parse()
	flag.Visit(func(item *flag.Flag) {
		for _, v := range c.allOptions {
			if v.flagName == item.Name {
				c.Set(v, item.Value.String())
			}
		}
	})
	c.clParsed = true

}

// Set function sets the key to the value of the appropiate type
// If the value passed is not of the expected type, does nothing
// and return an IncompatibleTypes error
func (c *GetConf) Set(o *Option, val interface{}) (err error) {
	v := reflect.ValueOf(val).Kind()
	if v == reflect.String {
		switch o.oType {
		case reflect.Int:
			o.value, err = strconv.ParseInt(val.(string), 10, 0)
			if err != nil {
				o.value = nil
			}
			return
		case reflect.Int8:
			o.value, err = strconv.ParseInt(val.(string), 10, 8)
			if err != nil {
				o.value = nil
			}
			return
		case reflect.Int16:
			o.value, err = strconv.ParseInt(val.(string), 10, 16)
			if err != nil {
				o.value = nil
			}
			return
		case reflect.Int32:
			o.value, err = strconv.ParseInt(val.(string), 10, 32)
			if err != nil {
				o.value = nil
			}
			return
		case reflect.Int64:
			o.value, err = strconv.ParseInt(val.(string), 10, 64)
			if err != nil {
				o.value = nil
			}
			return
		case reflect.Float32:
			o.value, err = strconv.ParseFloat(val.(string), 32)
			if err != nil {
				o.value = nil
			}
			return
		case reflect.Float64:
			o.value, err = strconv.ParseFloat(val.(string), 64)
			if err != nil {
				o.value = nil
			}
			return
		case reflect.Bool:
			o.value, err = strconv.ParseBool(val.(string))
			if err != nil {
				o.value = nil
			}
			return
		case reflect.String:
			o.value = val
			return
		}
	}
	if v == o.oType {
		o.value = val
		return
	}
	return IncompatibleTypes{
		msg:   "The value can not be converted to the option base type",
		valt:  v,
		baset: o.oType,
	}
}

// Get return the value associated to the key
func (c *GetConf) Get(key string) interface{} {
	if o, ok := c.allOptions[key]; ok != false {
		return o.value
	}
	return nil
}

// GetString will return the value associated to the key as a string
func (c *GetConf) GetString(key string) string {
	defer dontPanic()
	if val, ok := c.allOptions[key]; ok && val.value != nil {
		return val.value.(string)
	}
	return ""
}

// GetInt will return the value associated to the key as an int64
func (c *GetConf) GetInt(key string) (int64, error) {
	defer dontPanic()
	if val, ok := c.allOptions[key]; ok && val.value != nil {
		return val.value.(int64), nil
	}
	return 0, fmt.Errorf("Key %s not found", key)
}

// GetBool will return the value associated to the key as a bool
func (c *GetConf) GetBool(key string) (bool, error) {
	defer dontPanic()
	if val, ok := c.allOptions[key]; ok && val.value != nil {
		return val.value.(bool), nil
	}
	return false, fmt.Errorf("Key %s not found", key)
}

// GetFloat will return the value associated to the key as a float64
func (c *GetConf) GetFloat(key string) (float64, error) {
	defer dontPanic()
	if val, ok := c.allOptions[key]; ok && val.value != nil {
		return val.value.(float64), nil
	}
	return 0, fmt.Errorf("Key %s not found", key)
}

// GetAll return a map with the options and its values
// The values are of type interface{} so they have to be casted
func (c *GetConf) GetAll() map[string]interface{} {
	opts := make(map[string]interface{})
	for _, x := range c.allOptions {
		opts[x.name] = x.value
	}
	return opts
}

// dontPanic prevent for quitting if a panic occurs. It logs the
// panic and continue
func dontPanic() {
	if err := recover(); err != nil {
		log.Printf("PANIC Detected: %+v", err)
	}
}

// Stop make sure the connetions and channels are released
func (c *GetConf) Stop() {
	watchStopChan <- true
}
