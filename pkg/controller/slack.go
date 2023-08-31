/*
 * Copyright 2023 InfAI (CC SES)
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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
)

func (c *Controller) SendSlackMessage(message string) error {
	if len(c.config.SlackWebhookUrl) == 0 {
		log.Println("WARNING: SlackWebhookUrl not configured, assuming thats okay")
		return nil
	}
	j := make(map[string]string)
	j["text"] = message
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.config.SlackWebhookUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		err2 := errors.New("unexpected status code " + strconv.Itoa(resp.StatusCode) + ": " + string(body))
		if err != nil {
			return errors.Join(err, err2)
		}
		return err2
	}
	return nil
}
