/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package redisclient new a redis client
package redisclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTLS "yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	// timeout : allow TCP reconnection for 3 times(1, 2, 4)
	dialTimeout         = 8 * time.Second
	readTimeout         = 8 * time.Second
	writeTimeout        = 8 * time.Second
	idleTimeout         = 300 * time.Second
	defaultDialTimeout  = 8
	defaultReadTimeout  = 8
	defaultWriteTimeout = 8
	defaultIdleTimeout  = 300
	defaultRedisConn    = 20
	// TTL -
	TTL           = 1 * time.Minute
	maxRetryTimes = 3
	// DefaultCAFile is the default ca file for tls client
	DefaultCAFile = "/home/sn/resource/redis-secret/ca.pem"
	// DefaultCertFile is the default cert file for tls client
	DefaultCertFile = "/home/sn/resource/redis-secret/cert.pem"
	// DefaultKeyFile is the default key file for tls client
	DefaultKeyFile = "/home/sn/resource/redis-secret/key.pem"
	// redisStringFile is the temp file to store string type data of redis
	redisStringFile = "/tmp/redis-string"
	// redisStringFile is the temp file to store slice type data of redis
	redisSliceFile = "/tmp/redis-slice"
	redisSeparator = "%WITH%"

	// the detection is performed every 5 seconds.
	healthCheckIntervalTime = 5
	// 2 * 60min * 60s / 5 second, trigger every 5 minutes
	twoHoursCount             = 2 * 60 * 60 / healthCheckIntervalTime
	success                   = 0
	fail                      = 1
	redisValueIndex           = 2
	redisReconnectionInternal = 10 * time.Second
	// DefaultRedisContextTimeout -
	DefaultRedisContextTimeout = time.Second

	redisRetryTimes    = 3
	redisRetryInterval = 100 * time.Millisecond
)

var (
	errMode = errors.New("serverMode is not single or cluster")
	// RedisClient -
	redisClient = &Client{
		client:    &redis.Client{},
		option:    redisClientOption{},
		connected: false,
		RWMutex:   sync.RWMutex{},
	}
	// defaultTimeoutConf is the default timeout conf
	defaultTimeoutConf = TimeoutConf{
		DialTimeout:  defaultDialTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}
)

var (
	mu       sync.RWMutex
	redisCmd *Client
)

// Option -
type Option func(*redisClientOption)

type redisClientOption struct {
	tlsConfig       *tls.Config
	serverAddr      string
	dialTimeout     time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	password        string
	serverMode      string
	enableTLS       bool
	hotloadConfFunc func() (string, TimeoutConf, error)
	enableAlarm     bool
}

// RedisOperation -
type RedisOperation struct {
	Key    string
	Value  string
	Method string
	TTL    time.Duration
}

// Client -
type Client struct {
	client    redis.Cmdable
	option    redisClientOption
	connected bool
	sync.RWMutex
}

// Config is the config of redis client
type Config struct {
	ClusterID   string      `json:"clusterID,omitempty" valid:",optional"`
	ServerAddr  string      `json:"serverAddr,omitempty" valid:",optional"`
	ServerMode  string      `json:"serverMode,omitempty" valid:",optional"`
	Password    string      `json:"password,omitempty" valid:",optional"`
	EnableTLS   bool        `json:"enableTLS,omitempty" valid:",optional"`
	TimeoutConf TimeoutConf `json:"timeoutConf,omitempty" valid:",optional"`
}

// TimeoutConf A variety of timeout configurations
type TimeoutConf struct {
	DialTimeout  int `json:"dialTimeout,omitempty" valid:",optional"`
	ReadTimeout  int `json:"readTimeout,omitempty" valid:",optional"`
	WriteTimeout int `json:"writeTimeout,omitempty" valid:",optional"`
	IdleTimeout  int `json:"idleTimeout,omitempty" valid:",optional"`
}

// NewRedisClientParam parameters of a new redis client
type NewRedisClientParam struct {
	ServerMode      string
	ServerAddr      string
	Password        string
	Timeout         TimeoutConf
	EnableTLS       bool `json:"enableTLS,omitempty" valid:",optional"`
	HotloadConfFunc func() (string, TimeoutConf, error)
}

// GetRedisCmd -
func GetRedisCmd() *Client {
	mu.Lock()
	client := redisCmd
	mu.Unlock()
	return client
}

// SetRedisCmd -
func SetRedisCmd(client *Client) {
	mu.Lock()
	redisCmd = client
	mu.Unlock()
}

// SetEnableTLS -
func SetEnableTLS(enableTLS bool) Option {
	return func(c *redisClientOption) {
		c.enableTLS = enableTLS
	}
}

// ZCard -
func (c *Client) ZCard(ctx context.Context, key string) *redis.IntCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.ZCard(ctx, key)
}

// ZRange -
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.ZRange(ctx, key, start, stop)
}

// ZRem -
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.ZRem(ctx, key, members...)
}

// ZAdd -
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.ZAdd(ctx, key, members...)
}

// Ping -
func (c *Client) Ping(ctx context.Context) *redis.StatusCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.Ping(ctx)
}

// Expire -
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.Expire(ctx, key, expiration)
}

// Get -
func (c *Client) Get(ctx context.Context, key string) *redis.StringCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.Get(ctx, key)
}

// Del -
func (c *Client) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	c.RLock()
	cli := c.client
	c.RUnlock()
	return cli.Del(ctx, keys...)
}

// ZADDMetricsToRedis -
func ZADDMetricsToRedis(key string, metrics interface{}, limit int64, expireTime time.Duration) error {
	redisCmd := GetRedisCmd()
	if redisCmd == nil {
		log.GetLogger().Errorf("redis client is nil")
		return errors.New("redis client is nil")
	}
	count, err := redisCmd.ZCard(context.TODO(), key).Result()
	if err != nil {
		log.GetLogger().Errorf("failed to ZCard metrics key %s from redis, err: %s", key, err.Error())
		return err
	}
	// if count reach limit, delete the earliest metric
	if count >= limit {
		earliestValues, err := redisCmd.ZRange(context.TODO(), key, 0, 0).Result()
		if err != nil {
			log.GetLogger().Errorf("failed to ZRange metrics key %s from redis, err: %s", key, err.Error())
			return err
		}
		_, err = redisCmd.ZRem(context.TODO(), key, earliestValues[0]).Result()
		if err != nil {
			log.GetLogger().Errorf("failed to ZRem metrics key %s to redis, err: %s", key, err.Error())
			return err
		}
	}
	// Add a new value to the sorted set, with the score being the current timestamp
	score := time.Now().Unix()
	err = redisCmd.ZAdd(context.TODO(), key, redis.Z{Score: float64(score), Member: metrics}).Err()
	if err != nil {
		log.GetLogger().Errorf("failed to ZAdd metrics key %s to redis, err: %s", key, err.Error())
		return err
	}
	redisCmd.Expire(context.TODO(), key, expireTime)
	return nil
}

// New create a redis client
func New(newClientParam NewRedisClientParam, stopCh <-chan struct{}, options ...Option) (*Client, error) {
	o := getNewRedisOption(newClientParam)
	for _, option := range options {
		option(&o)
	}

	var redisCMD redis.Cmdable
	switch newClientParam.ServerMode {
	case "single":
		redisCMD = newSingleClient(o)
	case "cluster":
		redisCMD = newClusterClient(o)
	default:
		utils.ClearStringMemory(o.password)
		return nil, errMode
	}

	if redisCMD == nil {
		return nil, errors.New("failed to new redis cmd")
	}
	finished := make(chan int)
	go connectRedis(redisCMD, finished, o)
	select {
	case i, ok := <-finished:
		if ok && i == fail {
			return nil, errors.New("failed to connect redis server")
		}
	case <-time.After(o.dialTimeout):
		log.GetLogger().Errorf("dialing redis server error with incorrect ip address:%s.", o.serverAddr)
		return nil, errors.New("dialing redis server timeout")
	}
	redisClient = &Client{
		client:    redisCMD,
		option:    o,
		connected: true,
		RWMutex:   sync.RWMutex{},
	}
	return redisClient, nil
}

func getNewRedisOption(param NewRedisClientParam) redisClientOption {
	o := redisClientOption{
		serverAddr: param.ServerAddr,
		password:   param.Password,
		serverMode: param.ServerMode,
	}
	if param.Timeout.DialTimeout > 0 {
		o.dialTimeout = time.Duration(param.Timeout.DialTimeout) * time.Second
		log.GetLogger().Infof("new dialTimeout: %d", param.Timeout.DialTimeout)
	} else {
		o.dialTimeout = dialTimeout
	}
	if param.Timeout.ReadTimeout > 0 {
		o.readTimeout = time.Duration(param.Timeout.ReadTimeout) * time.Second
		log.GetLogger().Infof("new readTimeout: %d", param.Timeout.ReadTimeout)
	} else {
		o.readTimeout = readTimeout
	}
	if param.Timeout.WriteTimeout > 0 {
		o.writeTimeout = time.Duration(param.Timeout.WriteTimeout) * time.Second
		log.GetLogger().Infof("new writeTimeout: %d", param.Timeout.WriteTimeout)
	} else {
		o.writeTimeout = writeTimeout
	}
	if param.Timeout.IdleTimeout > 0 {
		o.idleTimeout = time.Duration(param.Timeout.IdleTimeout) * time.Second
		log.GetLogger().Infof("new idleTimeout: %d", param.Timeout.IdleTimeout)
	} else {
		o.idleTimeout = idleTimeout
	}
	return o
}

func newSingleClient(o redisClientOption) redis.Cmdable {
	options := &redis.Options{
		PoolSize:        defaultRedisConn,
		Addr:            o.serverAddr,
		Password:        o.password,
		DialTimeout:     o.dialTimeout,
		ReadTimeout:     o.readTimeout,
		WriteTimeout:    o.writeTimeout,
		ConnMaxIdleTime: o.idleTimeout,
		MaxRetries:      maxRetryTimes,
	}
	if o.enableTLS {
		tlsConfig, err := buildCfg(DefaultCAFile, DefaultCertFile, DefaultKeyFile)
		if err != nil {
			utils.ClearStringMemory(options.Password)
			log.GetLogger().Errorf("failed to build single client tls config: %s", err.Error())
			return nil
		}
		options.TLSConfig = tlsConfig
	}
	return redis.NewClient(options)
}

func connectRedis(redisCmd redis.Cmdable, finished chan<- int, o redisClientOption) {
	if finished == nil {
		return
	}
	var err error
	for i := 0; i < maxRetryTimes; i++ {
		if redisCmd == nil {
			log.GetLogger().Errorf("redis is not ready")
			continue
		}
		_, err = redisCmd.Ping(context.Background()).Result()
		if err == nil {
			finished <- success
			return
		}
	}
	// The key relies on go's GC for memory cleanup
	log.GetLogger().Errorf("dialing redis server error: %s", err.Error())
	finished <- fail
	return
}

func newClusterClient(o redisClientOption) redis.Cmdable {
	options := &redis.ClusterOptions{
		PoolSize:        defaultRedisConn,
		Addrs:           strings.Split(o.serverAddr, ","),
		Password:        o.password,
		DialTimeout:     o.dialTimeout,
		ReadTimeout:     o.readTimeout,
		WriteTimeout:    o.writeTimeout,
		ConnMaxIdleTime: o.idleTimeout,
		MaxRetries:      maxRetryTimes,
	}
	if o.enableTLS {
		tlsConfig, err := buildCfg(DefaultCAFile, DefaultCertFile, DefaultKeyFile)
		if err != nil {
			utils.ClearStringMemory(options.Password)
			log.GetLogger().Errorf("failed to build redis ClusterClient tls config: %s", err.Error())
			return nil
		}
		options.TLSConfig = tlsConfig
	}
	return redis.NewClusterClient(options)
}

func buildCfg(caFile string, certFile string, keyFile string) (*tls.Config, error) {
	var pools *x509.CertPool
	var err error
	pools, err = commonTLS.GetX509CACertPool(caFile)
	if err != nil {
		log.GetLogger().Errorf("failed to get X509 CACert Pool: %s", err.Error())
		return nil, err
	}

	var certs []tls.Certificate
	if certs, err = commonTLS.LoadServerTLSCertificate(certFile, keyFile, "", "LOCAL", false); err != nil {
		log.GetLogger().Errorf("failed to load Server TLS Certificate: %s", err.Error())
		return nil, err
	}

	clientAuth := tls.NoClientCert
	tlsConfig := &tls.Config{
		RootCAs:      pools,
		Certificates: certs,
		ClientAuth:   clientAuth,
	}
	return tlsConfig, nil
}

// CheckRedisConnectivity -
func CheckRedisConnectivity(clientRedisConfig *NewRedisClientParam, client *Client, stopCh <-chan struct{}) {
	if stopCh == nil {
		log.GetLogger().Errorf("stopCh is nil")
		return
	}
	ticker := time.NewTicker(redisReconnectionInternal)
	for {
		select {
		case <-ticker.C:
			if err := checkAndReconnectRedis(clientRedisConfig, client, stopCh); err != nil {
				log.GetLogger().Errorf("failed to check or reconnect redis client, err:%s", err.Error())
			}
		case <-stopCh:
			log.GetLogger().Errorf("module process exit")
			ticker.Stop()
			return
		}
	}
}

func checkAndReconnectRedis(clientRedisConfig *NewRedisClientParam, client *Client, stopCh <-chan struct{}) error {
	log.GetLogger().Debug("redis check redis connection start")
	if client != nil {
		_, err := (*client).Ping(context.TODO()).Result()
		if err == nil {
			log.GetLogger().Debug("redis periodically checks availability")
			return nil
		}
	}
	newClient, err := initClient(clientRedisConfig, stopCh)
	if err != nil {
		return err
	}
	if client != nil {
		client = newClient
	}
	SetRedisCmd(newClient)
	return nil
}

func initClient(clientRedisConfig *NewRedisClientParam, stopCh <-chan struct{}) (*Client, error) {
	c, err := New(NewRedisClientParam{
		ServerMode: clientRedisConfig.ServerMode,
		ServerAddr: clientRedisConfig.ServerAddr,
		Password:   clientRedisConfig.Password,
		Timeout:    clientRedisConfig.Timeout,
	}, stopCh, SetEnableTLS(clientRedisConfig.EnableTLS),
		SetGetRealTimeServerAddrFunc(clientRedisConfig.HotloadConfFunc))
	if err != nil {
		log.GetLogger().Errorf("failed to new a redis Client, %s", err.Error())
		return nil, err
	}
	return c, nil
}

// SetGetRealTimeServerAddrFunc hot update server address when disconnected
func SetGetRealTimeServerAddrFunc(getServerAddr func() (string, TimeoutConf, error)) Option {
	return func(c *redisClientOption) {
		c.hotloadConfFunc = getServerAddr
	}
}
