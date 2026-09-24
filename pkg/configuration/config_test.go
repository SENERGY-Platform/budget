/*
 *    Copyright 2026 InfAI (CC SES)
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package configuration

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

var mongoEnv = []string{"MONGO_URL", "MONGO_USER", "MONGO_PASSWORD", "MONGO_AUTH_SOURCE", "MONGO_DATABASE", "MONGO_BUDGET_COLLECTION", "MONGO_REPL_SET"}

// unsetMongoEnv clears the variables for the test; the loader ignores empty values and t.Setenv restores them.
func unsetMongoEnv(t *testing.T) {
	for _, v := range mongoEnv {
		t.Setenv(v, "")
	}
}

func mongoFields(c Config) map[string]any {
	return map[string]any{
		"MongoUrl":              c.MongoUrl,
		"MongoUser":             c.MongoUser,
		"MongoPassword":         c.MongoPassword,
		"MongoAuthSource":       c.MongoAuthSource,
		"MongoDatabase":         c.MongoDatabase,
		"MongoBudgetCollection": c.MongoBudgetCollection,
		"MongoReplSet":          c.MongoReplSet,
	}
}

// captureStdout returns what f prints, since the loader reports used variables there.
func captureStdout(t *testing.T, f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()
	out := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		out <- string(b)
	}()
	f()
	_ = w.Close()
	return <-out
}

func TestLoadMongoDefaults(t *testing.T) {
	unsetMongoEnv(t)
	conf, err := Load("../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"MongoUrl":              "mongodb://localhost:27017",
		"MongoUser":             "",
		"MongoPassword":         "",
		"MongoAuthSource":       "admin",
		"MongoDatabase":         "budget",
		"MongoBudgetCollection": "budgets",
		"MongoReplSet":          true,
	}
	if got := mongoFields(conf); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestLoadMongoFromEnv(t *testing.T) {
	unsetMongoEnv(t)
	t.Setenv("MONGO_URL", "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0")
	t.Setenv("MONGO_USER", "budget")
	t.Setenv("MONGO_PASSWORD", "p@ss:w/rd")
	t.Setenv("MONGO_AUTH_SOURCE", "users")
	t.Setenv("MONGO_DATABASE", "budget_test")
	var conf Config
	var err error
	printed := captureStdout(t, func() { conf, err = Load("../../config.json") })
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"MongoUrl":              "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0",
		"MongoUser":             "budget",
		"MongoPassword":         "p@ss:w/rd",
		"MongoAuthSource":       "users",
		"MongoDatabase":         "budget_test",
		"MongoBudgetCollection": "budgets",
		"MongoReplSet":          true,
	}
	if got := mongoFields(conf); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if strings.Contains(printed, "p@ss:w/rd") {
		t.Errorf("password printed: %q", printed)
	}
	if !strings.Contains(printed, "MONGO_DATABASE") {
		t.Errorf("used variables not reported: %q", printed)
	}
}
