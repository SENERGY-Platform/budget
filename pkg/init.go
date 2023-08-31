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

package pkg

import (
	"context"
	"github.com/SENERGY-Platform/budget/pkg/api"
	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"github.com/SENERGY-Platform/budget/pkg/database"
	"github.com/SENERGY-Platform/budget/pkg/validator"
	"sync"
)

// starts services and goroutines; returns a waiting group which is done as soon as all go routines are stopped
func Start(parent context.Context, config configuration.Config, onError func(err error, wg *sync.WaitGroup)) (wg *sync.WaitGroup, err error) {
	wg = &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(parent)
	db, err := database.New(config, ctx, wg)
	if err != nil {
		return wg, err
	}

	control := controller.New(config, db)

	if config.CheckAndQuit {
		v, err := validator.New(config, control)
		if err != nil {
			return wg, err
		}
		wg.Add(1)
		go func() {
			err := v.Run()
			wg.Done()
			cancel()
			if err != nil {
				onError(err, wg)
			}
		}()
	} else {
		err = api.Start(ctx, wg, config, control)
	}
	return
}
