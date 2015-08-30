package getconf

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

func PingEtcd(conn *etcd.Client) bool {
	_, err := conn.Get("/", false, false)
	res := err == nil
	return res
}

func etcdDial(url []string, rc chan *etcd.Client) chan *etcd.Client {
	go func() {
		tries := 0
		for {
			c := etcd.NewClient(url)
			// check if connected
			if PingEtcd(c) {
				log.Printf("[INF] connected to etcd server: %+v", url)
				rc <- c
				break
			}
			log.Printf("[ERR] Retry %d. Connection FAILED to etcd server: %+v", tries, url)
			time.Sleep(5 * time.Second)
			tries++
			if tries == 10 {
				close(rc)
				break
			}
		}
	}()
	return rc
}

// getEtcd will get the variables found in the etcd server and will watch the required Options for changes
func (c *GetConf) getEtcd() error {
	// check etcd
	// nodes := Get("nodes")
	// var node_arr []string
	// if nodes != "" {
	// 	for _, n := range strings.Split(nodes, ",") {
	// 		node_arr = append(node_arr, n)
	// 	}
	// } else {
	// 	node_arr = nil
	// }
	// etcdConn := etcd.NewClient(node_arr)
	responseChan = make(chan *etcd.Response)
	watchStopChan = make(chan bool, 1)

	rc := etcdDial([]string{c.etcdURL}, make(chan *etcd.Client))
	etcdConn = <-rc

	if etcdConn == nil {
		return errors.New("connection to etcd failed. Cannot monitor")
	}

	go c.waitResponses()

	for _, v := range c.allOptions {
		if v.noetcd == true || v.etcdName == "" {
			log.Printf("getconf.getEtcd: var %s not watched", v.name)
			continue
		}
		// If there is no value, check etcd
		if v.value == nil {
			// option not defined neither in env nor in flags. Query etcd for value
			resp, err := etcdConn.Get(v.etcdName, false, false)
			if err != nil {
				log.Printf("getconf.getEtcd: error getting var %s -> %s", v.etcdName, err)
			} else {
				c.Set(v, resp.Node.Value)
			}
		}
		go c.setupWatch(v.etcdName)
	}
	return nil
}

func (c *GetConf) waitResponses() {
	for {
		select {
		case resp, ok := <-responseChan:
			if ok {
				lkey := strings.TrimLeft(resp.Node.Key, "/")
				// Get Option
				for _, v := range c.allOptions {
					if v.etcdName == lkey {
						log.Printf("getconf.etcd.watcher: value changed. setting %+v", resp)
						c.Set(v, resp.Node.Value)
						c.ConfChanged <- map[string]interface{}{"key": v.name, "value": resp.Node.Value}
					}
				}
			} else {
				log.Printf("getconf: channel closed. Connection lost?")
				c.recoverEtcd()
				return
			}
		}
	}
}

func (c *GetConf) recoverEtcd() {
	if !PingEtcd(etcdConn) {
		// Wait for all watch goroutines to exit
		//log.Printf("recoverEtcd: waiting for all watch goroutines to exit")
		//wg.Wait()
		// setup again from the beginning...
		log.Printf("recoverEtcd: calling c.getEtcd()")
		err := c.getEtcd()
		if err != nil {
			log.Printf("getconf: cannot reconnect: %+v", err)
		}
		log.Printf("recoverEtcd done")
	}
}

func (c *GetConf) setupWatch(en string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("getconf.getEtcd.watcher channel error: %v", r)
		}
	}()
	if PingEtcd(etcdConn) {
		log.Printf("getconf.getEtcd.watcher Watching: %s", en)
		_, err := etcdConn.Watch(en, 0, false, responseChan, watchStopChan)
		if err != etcd.ErrWatchStoppedByUser {
			log.Printf("getconf.getEtcd.watcher error: %s", err.Error())
		}
	}
}
