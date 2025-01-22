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
	"log"
	"net/http"

	"github.com/SENERGY-Platform/budget/pkg/api/util"
	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"github.com/julienschmidt/httprouter"
)

func init() {
	endpoints = append(endpoints, CheckImportDeployEndpoints)
}

func CheckImportDeployEndpoints(router *httprouter.Router, c configuration.Config, control *controller.Controller) {
	router.POST("/check/import/deploy", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		parsed, err := util.ParseRequest(request.Body, c.Debug)
		if err != nil {
			log.Println("ERROR: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
		}

		code, err := control.CheckImportDeploy(parsed)
		if err != nil {
			log.Println("ERROR: " + err.Error())
			http.Error(writer, err.Error(), code)
		}
	})
}
