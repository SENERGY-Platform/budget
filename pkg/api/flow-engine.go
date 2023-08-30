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

package api

import (
	"github.com/SENERGY-Platform/budget/pkg/api/util"
	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func init() {
	endpoints = append(endpoints, CheckFlowEngineEndpoints)
}

func CheckFlowEngineEndpoints(router *httprouter.Router, _ configuration.Config, control *controller.Controller) {
	router.POST("/check/analytics/flow-engine", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		parsed, err := util.ParseRequest(request.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}

		code, err := control.CheckFlowEngine(parsed)
		if err != nil {
			http.Error(writer, err.Error(), code)
		}
	})
}
