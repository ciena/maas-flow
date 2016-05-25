package main

import (
	"github.com/fzzy/radix/redis"
	"log"
	"net/url"
	"os"
)

// Tracker used to track if a node has been post deployed provisioned
type Tracker interface {
	Get(key string) (bool, error)
	Set(key string) error
	Clear(key string) error
}

// RedisTracker redis implementation of the tracker interface
type RedisTracker struct {
	client *redis.Client
}

func (t *RedisTracker) Get(key string) (bool, error) {
	reply := t.client.Cmd("get", key)
	if reply.Err != nil {
		return false, reply.Err
	}
	if reply.Type == redis.NilReply {
		return false, nil
	}

	value, err := reply.Bool()
	return value, err
}

func (t *RedisTracker) Set(key string) error {
	reply := t.client.Cmd("set", key, true)
	return reply.Err
}

func (t *RedisTracker) Clear(key string) error {
	reply := t.client.Cmd("del", key)
	return reply.Err
}

// MemoryTracker in memory implementation of the tracker interface
type MemoryTracker struct {
	data map[string]bool
}

func (m *MemoryTracker) Get(key string) (bool, error) {
	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return false, nil
}

func (m *MemoryTracker) Set(key string) error {
	m.data[key] = true
	return nil
}

func (m *MemoryTracker) Clear(key string) error {
	delete(m.data, key)
	return nil
}

// NetTracker constructs an implemetation of the Tracker interface. Which implementation selected
//            depends on the environment. If a link to a redis instance is defined then this will
//            be used, else an in memory version will be used.
func NewTracker() Tracker {
	// Check the environment to see if we are linked to a redis DB
	if os.Getenv("AUTODB_ENV_REDIS_VERSION") != "" {
		tracker := new(RedisTracker)
		if spec := os.Getenv("AUTODB_PORT"); spec != "" {
			port, err := url.Parse(spec)
			checkError(err, "[error] unable to lookup to redis database : %s", err)
			tracker.client, err = redis.Dial(port.Scheme, port.Host)
			checkError(err, "[error] unable to connect to redis database : '%s' : %s", port, err)
			log.Println("[info] Using REDIS to track provisioning status of nodes")
			return tracker
		} else {
			log.Fatalf("[error] looks like we are configured for REDIS, but no PORT defined in environment")
		}
	}

	// Else fallback to an in memory tracker
	tracker := new(MemoryTracker)
	tracker.data = make(map[string]bool)
	log.Println("[info] Using memory based structures to track provisioning status of nodes")
	return tracker
}