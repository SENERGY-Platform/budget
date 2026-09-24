/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package database

import (
	"context"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/log"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMain(m *testing.M) {
	log.InitForTest()
	os.Exit(m.Run())
}

func mongoConf(url string) configuration.Config {
	return &configuration.ConfigStruct{
		MongoUrl:              url,
		MongoAuthSource:       "admin",
		MongoDatabase:         "budget",
		MongoBudgetCollection: "budgets",
		MongoTimeout:          "10s",
	}
}

// unusedAddr returns a loopback address that was just free, so no server answers there.
func unusedAddr(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	if err = l.Close(); err != nil {
		t.Fatal(err)
	}
	return addr
}

func TestClientOptionsAuthWhenUserGiven(t *testing.T) {
	conf := mongoConf("mongodb://localhost:27017")
	conf.MongoUser, conf.MongoPassword = "budget", "s3cr3t"
	opts, err := clientOptions(conf)
	if err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "budget", Password: "s3cr3t", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

func TestClientOptionsNoAuthWhenUserEmpty(t *testing.T) {
	// A password without a user must not switch auth on.
	conf := mongoConf("mongodb://localhost:27017")
	conf.MongoPassword = "s3cr3t"
	opts, err := clientOptions(conf)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Auth != nil {
		t.Errorf("auth = %+v, want nil", opts.Auth)
	}
}

func TestClientOptionsUserReplacesURICredentials(t *testing.T) {
	conf := mongoConf("mongodb://old:oldpw@localhost:27017/?authSource=other&authMechanism=SCRAM-SHA-1")
	conf.MongoUser, conf.MongoPassword = "new", "newpw"
	opts, err := clientOptions(conf)
	if err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "new", Password: "newpw", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

func TestClientOptionsURIPassedUnchanged(t *testing.T) {
	uri := "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0&readPreference=primary"
	opts, err := clientOptions(mongoConf(uri))
	if err != nil {
		t.Fatal(err)
	}
	if got := opts.GetURI(); got != uri {
		t.Errorf("uri = %q, want %q", got, uri)
	}
	if want := []string{"mongo-0.mongo:27017", "mongo-1.mongo:27017"}; !reflect.DeepEqual(opts.Hosts, want) {
		t.Errorf("hosts = %v, want %v", opts.Hosts, want)
	}
	if opts.ReplicaSet == nil || *opts.ReplicaSet != "rs0" {
		t.Errorf("replica set = %v, want rs0", opts.ReplicaSet)
	}
}

func TestClientOptionsRejects(t *testing.T) {
	cases := map[string]func(c configuration.Config){
		"empty database":        func(c configuration.Config) { c.MongoDatabase = "" },
		"blank database":        func(c configuration.Config) { c.MongoDatabase = "  " },
		"user without password": func(c configuration.Config) { c.MongoUser = "budget" },
		// The old value of a service that prefixed the scheme itself.
		"uri without scheme": func(c configuration.Config) {
			c.MongoUrl, c.MongoUser, c.MongoPassword = "localhost:27017", "budget", "s3cr3t"
		},
	}
	for name, modify := range cases {
		t.Run(name, func(t *testing.T) {
			conf := mongoConf("mongodb://localhost:27017")
			modify(conf)
			_, err := clientOptions(conf)
			if err == nil {
				t.Fatal("expected an error")
			}
			if strings.Contains(err.Error(), "s3cr3t") {
				t.Errorf("error leaks the password: %v", err)
			}
		})
	}
}

func TestNewRejectsInvalidConfigBeforeConnecting(t *testing.T) {
	cases := map[string]func(c configuration.Config){
		"empty database":        func(c configuration.Config) { c.MongoDatabase = "" },
		"user without password": func(c configuration.Config) { c.MongoUser = "budget" },
	}
	for name, modify := range cases {
		t.Run(name, func(t *testing.T) {
			conf := mongoConf("mongodb://" + unusedAddr(t) + "/?directConnection=true")
			modify(conf)
			wg := &sync.WaitGroup{}
			_, err := New(conf, context.Background(), wg)
			if err == nil {
				t.Fatal("expected an error")
			}
			if strings.Contains(err.Error(), "startup check") {
				t.Errorf("config was not checked before connecting: %v", err)
			}
			wg.Wait()
		})
	}
}

// poolCounter counts connection pools the driver opens and closes.
type poolCounter struct {
	mu              sync.Mutex
	created, closed int
}

func (p *poolCounter) monitor() *event.PoolMonitor {
	return &event.PoolMonitor{Event: func(e *event.PoolEvent) {
		p.mu.Lock()
		defer p.mu.Unlock()
		switch e.Type {
		case event.PoolCreated:
			p.created++
		case event.PoolClosedEvent:
			p.closed++
		}
	}}
}

func (p *poolCounter) counts() (int, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.created, p.closed
}

func TestConnectFailsWithoutServerAndDisconnects(t *testing.T) {
	opts, err := clientOptions(mongoConf("mongodb://" + unusedAddr(t) + "/?directConnection=true"))
	if err != nil {
		t.Fatal(err)
	}
	pools := &poolCounter{}
	opts.SetPoolMonitor(pools.monitor())
	start := time.Now()
	client, err := connect(context.Background(), opts, "budget", 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected an error")
	}
	if client != nil {
		t.Error("expected no client on failure")
	}
	if !strings.HasPrefix(err.Error(), "mongo startup check failed: ") {
		t.Errorf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("connect took %v, the timeout was not applied", elapsed)
	}
	if created, closed := pools.counts(); created == 0 || closed != created {
		t.Errorf("pools created %d, closed %d: the client was not disconnected", created, closed)
	}
}

func TestNewFailsWithoutServer(t *testing.T) {
	conf := mongoConf("mongodb://" + unusedAddr(t) + "/?directConnection=true&serverSelectionTimeoutMS=500")
	wg := &sync.WaitGroup{}
	db, err := New(conf, context.Background(), wg)
	if err == nil {
		t.Fatal("expected an error")
	}
	if db != nil {
		t.Error("expected no db on failure")
	}
	if !strings.HasPrefix(err.Error(), "mongo startup check failed: ") {
		t.Errorf("unexpected error: %v", err)
	}
	// Nothing may be left waiting for the service context.
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("wait group still busy after a failed start")
	}
}
