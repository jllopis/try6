getconf
======

**v0.3.1**

Go library to load configuration variables from OS environment, command line and/or [etcd](https://github.com/coreos/etcd) server.

## Requirements

* **go** ~> 1.4.2

## Installation

    go get github.com/jllopis/getconf

## Quick Start

1. Include the package *bitbucket.org/jllopis/getconf* in your file
2. Create a *struct* to hold the variables
3. Call `getconf.New(myconf interface{}, envPrefix string, useEtcd bool, etcdURL string) *GetConf`
   where:
     - *myconf* is the *struct* holding your variables
     - *envPrefix* is the prefix to be used when calling the environment (can be empty "")
     - *useEtcd* instruct to watch variables in etcd server. Must be set *true* for watching or *false* otherwise
     - *etcdURL* the URL where the *etcd* server listen
4. If using command line flags, call `getconf.Parse()` to get them. Be sure to not call `flags.Parse()` directly as it is done by *getconf*
5. If needed, listen for changes binding to the **GetConf.ConfChanged** channel
6. Use the variables through the **get** methods provided

```go
package main

import (
	"bitbucket.org/jllopis/getconf"
	"log"
)

type MyConf struct {
	Key1 string  `getconf:"etcd app/keys/conf/key1, env T_ENVKEY, flag flagkey, noetcd"`
	Key2 bool    `getconf:"-"`
	Key3 float64 `getconf:"etcd /app/keys/conf/my3rdkey"`
	Key4 int
}

func main() {
	//getconf.New(conf struct, envPrefix string, useEtcd bool, etcdURL string) *GetConf
	conf := getconf.New(&MyConf{}, "T", true, "")
	defer conf.Stop()
	conf.Parse()

	for k, v := range conf.GetAll() {
		log.Printf("Key: %s\tValue: %v\n", k, v)
	}

	// watch for changes from etcd
	for resp := range conf.ConfChanged {
		log.Printf("Configuration Changed:  Key: %s\tValue: %s\n", resp["key"], resp["value"])
		// We can see that the stored variables has also changed
		log.Printf("Key: %s\tValue: %v\n", resp["key"], conf.Get(resp["key"].(string)))
	}
```

And call it by

```bash
(go1.2) $ T_KEY2=true T_KEY1="hello" go run main.go -key4 23 -key3 25.6
2013/12/18 19:15:06 Key: Key1	Value: hello
2013/12/18 19:15:06 Key: Key2	Value: true
2013/12/18 19:15:06 Key: Key3	Value: 25.6
2013/12/18 19:15:06 Key: Key4	Value: 23
(go1.2) $
```

If using **etcd** you are responsible of reacting upon variable changes notified by the etcd response channel.

## Conventions

The options can be defined in:

1. environment
2. etcd
3. command line flags if called `getconf.Parse()` function

The order is the specified, meaning that the last option will win (if you set an environment variable it can be ovewritten by a command line flag).

If a key has no value when connecting to **etcd** we will query the server for a value before setting the watch. That way you can populate all your vars from *etcd* but explicitly set variables will not be overriten. They'll only be if updated on **etcd**.

To be parsed, you must define a struct in your program that will define the name and the type of the variables. The struct members **must** be uppercase (exported) otherwise _reflection_ will not work.

The struct can be any lenght and supported types are:

* int, int8, int16, int32, int64
* float32, float64
* string
* bool

Any other type will be discarded.

If a value can not be matched to the variable type, it will be discarded and the variable set to **nil**.

### tags

There are some tags that can be used:

- **-**: If a dash is found the variable will not be watched for changes in etcd.
- **env**: Set the variable name in the environment. If a **prefix** is provided, it will be prependet with an underscore (*prefix_*).
- **flag**: Set the variable name for flag parsing (command line options). If no specified, it will default to the struct name lowercase.
- **etcd**: **Required** to watch the variable in etcd. Sets the full path to the variable to be watched (see sample above).
- **noetcd**: If found the variable will not be watched in etcd even if a etcd name is specified.

### etcd

The variables in **etcd** **must** have a tag *etcd* in the *Config struct*. It there is no such tag the variable will be omitted and not watched.

To be notified, a channel is provided in the **Conf struct**: **ConfChanged**. Every change in the watched variables will be notified through this channel using a map `map[string]string` with two elements "key" and "value". See the example above for instructions about how to use it.

When a change occurs, you can access the variable through the usual **Get** functions.

To be polite, you can defer the getconf.Stop() function to close channels and connections before quitting the application.

### environment

The variables can have a prefix provided by the user. This is useful to prevent collisions. So you can set

    FB_VAR1="a value"

and at the same time

    YZ_VAR1=233

being _prefixes_ "FB" and "YZ".

The prefix **must** be followed by an underscore **'_'** that separates the prefix from the variable name. This name must be the same defined in the struct but **must** be UPPERCASE. Lower and Mixed case environment variables will not be taken into account.

This default can be overwritten setting the tag **env** in the variable tags when defining the Struct.

### command line flags

Command line flags are standard variables from the _go_ **flag** package. The flags **must** be all lowercase and the lowercase name must match the ones defined in the struct. This variables will be readed upon request by calling `getconf.Parse()` method.

In command line, a _boolean_ flag acts as a switch, that is, it will take the value of **true** if present and **false** otherwise. If a boolean flag has associated a value it will error:

    $ ./executable -correct -notcorrect false

In the example we can see that _correct_ will set the _Correct_ var to _true_ while _notcorrect_ will error. **IMPORTANT**: in this case, the flags following the offending _notcorrect_ *will not be parsed* and will be unassigned (nil).

This default can be overwritten setting the tag **env** in the variable tags when defining the Struct.

