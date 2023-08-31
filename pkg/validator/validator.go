/*
 *    Copyright 2023 InfAI (CC SES)
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

package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Validator struct {
	config      configuration.Config
	controller  *controller.Controller
	clientToken *OpenidToken
}

var validations = []func(controller *controller.Controller, token string, userid string, roles []string) (budgetIdentifier string, maxBudget uint64, actualBudget uint64, err error){}

func New(config configuration.Config, controller *controller.Controller) (v *Validator, err error) {
	v = &Validator{config: config, controller: controller}
	err = v.setClientToken()
	return
}

func (v *Validator) Run() (err error) {
	users, err := v.getAllUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		roles, err := v.getUserRoles(user.Id)
		if err != nil {
			return err
		}
		var token *OpenidToken
		for _, validation := range validations {
			if token == nil || token.Expired() {
				token, err = v.exchangeUserToken(user.Id)
				if err != nil {
					return err
				}
			}
			budgetIdentifier, maxBudget, actualBudget, err := validation(v.controller, "Bearer "+token.AccessToken, user.Id, roles)
			if err != nil {
				return err
			}
			if v.config.Debug {
				log.Printf("%s \t %s \t %s \t %d/%d\n", user.Username, user.Id, budgetIdentifier, actualBudget, maxBudget)
			}
			if actualBudget > maxBudget {
				err := v.notifyOverBudget(budgetIdentifier, user, maxBudget, actualBudget)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *Validator) setClientToken() (err error) {
	now := time.Now()
	resp, err := http.PostForm(v.config.KeycloakUrl+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":     {v.config.KeycloakClientId},
		"client_secret": {v.config.KeycloakClientSecret},
		"grant_type":    {"client_credentials"},
	})
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return handleHttpErr(resp)
	}
	v.clientToken = &OpenidToken{}
	err = json.NewDecoder(resp.Body).Decode(v.clientToken)
	if err != nil {
		return
	}
	v.clientToken.requestTime = now
	return
}

func (v *Validator) exchangeUserToken(userId string) (token *OpenidToken, err error) {
	now := time.Now()
	resp, err := http.PostForm(v.config.KeycloakUrl+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":         {v.config.KeycloakClientId},
		"client_secret":     {v.config.KeycloakClientSecret},
		"grant_type":        {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"requested_subject": {userId},
	})
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return nil, handleHttpErr(resp)
	}
	token = &OpenidToken{}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return
	}
	token.requestTime = now
	return
}

func (v *Validator) getAllUsers() (users []KeycloakUser, err error) {
	limit := 1000
	users = []KeycloakUser{}
	for {
		if v.clientToken == nil || v.clientToken.Expired() {
			err = v.setClientToken()
			if err != nil {
				return
			}
		}

		req, err := http.NewRequest(http.MethodGet, v.config.KeycloakUrl+"/auth/admin/realms/master/users?max="+strconv.Itoa(limit)+"&first="+strconv.Itoa(len(users)), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+v.clientToken.AccessToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, handleHttpErr(resp)
		}
		newUsers := []KeycloakUser{}
		err = json.NewDecoder(resp.Body).Decode(&users)
		if err != nil {
			return nil, err
		}
		users = append(users, newUsers...)
		if len(newUsers) < limit {
			break
		}
	}

	return
}

func (v *Validator) getUserRoles(userId string) (roles []string, err error) {
	if v.clientToken == nil || v.clientToken.Expired() {
		err = v.setClientToken()
		if err != nil {
			return
		}
	}

	req, err := http.NewRequest(http.MethodGet, v.config.KeycloakUrl+"/auth/admin/realms/master/users/"+userId+"/role-mappings/realm", nil)
	req.Header.Set("Authorization", "Bearer "+v.clientToken.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, handleHttpErr(resp)
	}
	roleMappings := []KeycloakRealmRoleMapping{}
	err = json.NewDecoder(resp.Body).Decode(&roleMappings)
	if err != nil {
		return nil, err
	}
	roles = []string{}
	for _, mapping := range roleMappings {
		roles = append(roles, mapping.Name)
	}
	return
}

func (v *Validator) notifyOverBudget(budgetIdentifier string, user KeycloakUser, maxBudget uint64, actualBudget uint64) error {
	msg := fmt.Sprintf("ALERT! User %s %s overuses budget %s by %d", user.Username, user.Id, budgetIdentifier, actualBudget-maxBudget)
	log.Println(msg)
	return v.controller.SendSlackMessage(msg)
}

func handleHttpErr(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return errors.New(strconv.Itoa(resp.StatusCode) + " " + string(body))
}
