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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func randomHex(t *testing.T) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

// TestNewAuthenticates needs root credentials of a throwaway server with access control, given by
// MONGO_AUTH_TEST_URL, _USER and _PASSWORD; it creates and drops its own users and databases.
func TestNewAuthenticates(t *testing.T) {
	url, rootUser, rootPassword := os.Getenv("MONGO_AUTH_TEST_URL"), os.Getenv("MONGO_AUTH_TEST_USER"), os.Getenv("MONGO_AUTH_TEST_PASSWORD")
	if testing.Short() || url == "" || rootUser == "" || rootPassword == "" {
		t.Skip("needs MONGO_AUTH_TEST_URL, MONGO_AUTH_TEST_USER and MONGO_AUTH_TEST_PASSWORD, not in -short")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	root, err := mongo.Connect(ctx, options.Client().ApplyURI(url).SetAuth(options.Credential{Username: rootUser, Password: rootPassword, AuthSource: "admin"}))
	if err != nil {
		t.Fatal(err)
	}
	// Registered first so it runs after the user cleanups, which need the connection.
	t.Cleanup(func() { _ = root.Disconnect(context.Background()) })

	suffix := randomHex(t)
	database, otherDatabase := "budget_authtest_"+suffix, "other_authtest_"+suffix
	user, otherUser := "budget-"+suffix, "other-"+suffix
	password, otherPassword := randomHex(t), randomHex(t)
	createUser := func(name, pw, role, db string) {
		// Registered before creation and tolerant of a missing user, so a half-done setup is cleaned up too.
		t.Cleanup(func() {
			err := root.Database("admin").RunCommand(context.Background(), bson.D{{Key: "dropUser", Value: name}}).Err()
			var cmdErr mongo.CommandError
			if err != nil && !(errors.As(err, &cmdErr) && cmdErr.Code == 11) {
				t.Errorf("unable to drop user %v: %v", name, err)
			}
			if err := root.Database(db).Drop(context.Background()); err != nil {
				t.Errorf("unable to drop database %v: %v", db, err)
			}
		})
		err := root.Database("admin").RunCommand(ctx, bson.D{
			{Key: "createUser", Value: name},
			{Key: "pwd", Value: pw},
			{Key: "roles", Value: bson.A{bson.D{{Key: "role", Value: role}, {Key: "db", Value: db}}}},
		}).Err()
		if err != nil {
			t.Fatal(err)
		}
	}
	readUser, readPassword := "budget-read-"+suffix, randomHex(t)
	createUser(user, password, "readWrite", database)
	createUser(otherUser, otherPassword, "readWrite", otherDatabase)
	createUser(readUser, readPassword, "read", database)

	cases := []struct {
		name     string
		user     string
		password string
		wantErr  bool
	}{
		{"correct user", user, password, false},
		{"no credentials", "", "", true},
		{"user of another database", otherUser, otherPassword, true},
		{"wrong password", user, password + "-wrong", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := mongoConf(url)
			conf.MongoUser, conf.MongoPassword, conf.MongoDatabase = c.user, c.password, database
			svcCtx, stop := context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}
			defer wg.Wait()
			defer stop()
			db, err := New(conf, svcCtx, wg)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if err != nil {
				for _, pw := range []string{password, otherPassword, readPassword, rootPassword} {
					if strings.Contains(err.Error(), pw) {
						t.Fatal("error contains a password")
					}
				}
				if !strings.HasPrefix(err.Error(), "mongo startup check failed: ") {
					t.Errorf("unexpected error: %v", err)
				}
				t.Log(err)
				return
			}
			budget := models.Budget{BudgetIdentifier: "authtest", UserId: "user", Role: "role"}
			if err = db.SetBudget(context.Background(), budget); err != nil {
				t.Fatalf("write after successful startup failed: %v", err)
			}
			list, err := db.ListBudgets(context.Background(), 10, 0, "authtest", "user", nil)
			if err != nil || len(list) != 1 {
				t.Errorf("read after successful startup: %v, %v", list, err)
			}
		})
	}

	// A read-only user passes the startup check but not index creation; New must not leave that client connected.
	t.Run("read-only user", func(t *testing.T) {
		pools := &poolCounter{}
		t.Cleanup(func() { newClientOptions = clientOptions })
		newClientOptions = func(conf configuration.Config) (*options.ClientOptions, error) {
			opts, err := clientOptions(conf)
			if err != nil {
				return nil, err
			}
			return opts.SetPoolMonitor(pools.monitor()), nil
		}
		conf := mongoConf(url)
		conf.MongoUser, conf.MongoPassword, conf.MongoDatabase = readUser, readPassword, database
		svcCtx, stop := context.WithCancel(context.Background())
		defer stop()
		wg := &sync.WaitGroup{}
		_, err := New(conf, svcCtx, wg)
		if err == nil {
			wg.Wait()
			t.Fatal("expected index creation to fail")
		}
		if strings.Contains(err.Error(), "startup check") {
			t.Fatalf("failed at the startup check instead of index creation: %v", err)
		}
		for _, pw := range []string{password, otherPassword, readPassword, rootPassword} {
			if strings.Contains(err.Error(), pw) {
				t.Fatal("error contains a password")
			}
		}
		t.Log(err)
		if created, closed := pools.counts(); created == 0 || closed != created {
			t.Errorf("pools created %d, closed %d: the client was not disconnected", created, closed)
		}
	})
}
