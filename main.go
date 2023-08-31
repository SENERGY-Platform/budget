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

package main

import (
	"context"
	"flag"
	"github.com/SENERGY-Platform/budget/pkg"
	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	config, err := configuration.Load(*configLocation)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	onError := func(err error, wg *sync.WaitGroup) {
		log.Println("FATAL: " + err.Error())
		cancel()
		if wg != nil {
			go func() { // safety routine shuts program down if wg does not finish
				d := 10 * time.Second
				<-time.After(d)
				log.Println("Could not shutdown in " + d.String() + ", halting now")
				os.Exit(1)
			}()
			wg.Wait()
		}
		os.Exit(1)
	}

	wg, err := pkg.Start(ctx, config, onError)
	if err != nil {
		onError(err, wg)
	}

	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		sig := <-shutdown
		log.Println("received shutdown signal", sig)
		cancel()
	}()

	wg.Wait()
}
