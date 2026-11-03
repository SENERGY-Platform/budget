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

package controller

import (
	"strings"
	"testing"

	"github.com/SENERGY-Platform/budget/pkg/configuration"
)

func TestSendSlackMessageErrorHidesWebhook(t *testing.T) {
	// The space makes the URL unparsable, so http.Post fails at once with a *url.Error that carries the URL.
	c := &Controller{config: &configuration.ConfigStruct{SlackWebhookUrl: "http://hooks .slack.com/services/T000/B000/not-a-real-token"}}
	err := c.SendSlackMessage("test")
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "not-a-real-token") {
		t.Errorf("webhook in error: %v", err)
	}
}
