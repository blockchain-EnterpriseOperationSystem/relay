/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package redis

import (
	"fmt"
	"github.com/Loopring/relay/config"
	"github.com/Loopring/relay/log"
	"github.com/garyburd/redigo/redis"
	"time"
	"errors"
)

type RedisCacheImpl struct {
	options config.RedisOptions
	pool    *redis.Pool
}

func (impl *RedisCacheImpl) Initialize(cfg interface{}) {
	options := cfg.(config.RedisOptions)
	impl.options = options

	impl.pool = &redis.Pool{
		IdleTimeout: time.Duration(options.IdleTimeout) * time.Second,
		MaxIdle:     options.MaxIdle,
		MaxActive:   options.MaxActive,
		Dial: func() (redis.Conn, error) {
			address := fmt.Sprintf("%s:%s", options.Host, options.Port)
			var (
				c   redis.Conn
				err error
			)
			if len(options.Password) > 0 {
				c, err = redis.Dial("tcp", address, redis.DialPassword(options.Password))
			} else {
				c, err = redis.Dial("tcp", address)
			}

			if err != nil {
				log.Fatal(err.Error())
				return nil, err
			}

			return c, nil
		},
	}
}

func (impl *RedisCacheImpl) Get(key string) ([]byte, error) {
	conn := impl.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("get", key)

	if nil == reply {
		if nil == err {
			err = fmt.Errorf("no this key:%s", key)
		}
		return []byte{}, err
	} else {
		return reply.([]byte), err
	}
}

func (impl *RedisCacheImpl) Exists(key string) (bool, error) {
	conn := impl.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("exists", key)

	if err != nil {
		return false, err
	} else {
		exists := reply.(int64)
		if exists == 1 {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func (impl *RedisCacheImpl) Set(key string, value []byte, ttl int64) error {
	conn := impl.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("set", key, value); err != nil {
		return err
	}
	if _, err := conn.Do("expire", key, ttl); err != nil {
		return err
	}
	return nil
}

func (impl *RedisCacheImpl) Del(key string) error {
	conn := impl.pool.Get()
	defer conn.Close()

	_, err := conn.Do("del", key)

	return err
}

func (impl *RedisCacheImpl) HMSet(key string, args ...[]byte) error {
	conn := impl.pool.Get()
	defer conn.Close()

	if len(args) % 2 != 0 {
		return errors.New("the length of `args` must be even")
	}
	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range args {
		vs = append(vs, v)
	}
	_, err := conn.Do("hmset", vs...)

	return err
}

func (impl *RedisCacheImpl) HMGet(key string, fields ...[]byte) ([][]byte,error) {
	conn := impl.pool.Get()
	defer conn.Close()

	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range fields {
		println()
		vs = append(vs, v)
	}
	reply, err := conn.Do("hmget", vs...)

	res := [][]byte{}

	if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _,r := range rs {
			if nil == r {
				res = append(res, []byte{})
			} else {
				res = append(res, r.([]byte))
			}
		}
	} else {
		log.Errorf("HMGet err:%s", err.Error())
	}
	return res,err
}

func (impl *RedisCacheImpl) HGetAll(key string) ([][]byte,error) {
	conn := impl.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("hgetall", key)

	res := [][]byte{}
	if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _,r := range rs {
			res = append(res, r.([]byte))
		}
	}
	return res,err
}

func (impl *RedisCacheImpl) HExists(key string, field []byte) (bool,error) {
	conn := impl.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("hexists", key, field)

	if nil == err && nil != reply {
		exists := reply.(int)
		return exists > 0, nil
	}

	return false,err
}

func (impl *RedisCacheImpl) SAdd(key string, members ...[]byte) error {
	conn := impl.pool.Get()
	defer conn.Close()

	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range members {
		vs = append(vs, v)
	}
	_, err := conn.Do("sadd", vs...)

	return err
}

func  (impl *RedisCacheImpl) SMembers(key string) ([][]byte,error) {
	conn := impl.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("smembers", key)

	res := [][]byte{}
	if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _,r := range rs {
			res = append(res, r.([]byte))
		}
	}
	return res,err
}