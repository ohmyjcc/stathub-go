/*
 * A smart Hub for holding server stat
 * https://www.likexian.com/
 *
 * Copyright 2015-2019, Li Kexian
 * Released under the Apache License, Version 2.0
 *
 */

package main

import (
	"bytes"
	"fmt"
	"github.com/likexian/simplejson-go"
	"github.com/likexian/gokit/assert"
	"math/rand"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	confFile = "../bin/test.conf"
	testFile = "../bin/test.sh"
	testText = `#!/bin/bash
	go build -v -o ../bin/stathub
	cd ../bin
	rm -rf cert data log
	./stathub -c test.conf --init-server
	sed -ie 's/\/usr\/local\/stathub/./g' test.conf
	rm -rf test.confe
	./stathub -c test.conf
	`
)

func startServer() {
	var stdout bytes.Buffer
	cmd := exec.Command(testFile)
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	cmd.Run()
	fmt.Println(string(stdout.Bytes()))
}

func TestApiStat(t *testing.T) {
	os.Remove(confFile)

	err := WriteFile(testFile, testText)
	assert.Nil(t, err)

	err = os.Chmod(testFile, 0755)
	assert.Nil(t, err)

	go startServer()

	for {
		if FileExists(confFile) {
			time.Sleep(1 * time.Second)
			break
		}
		time.Sleep(1 * time.Second)
	}

	CLIENT_CONF, err := GetConfig(confFile)
	assert.Nil(t, err)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		stat := Stat{}
		diskWarn := ""
		if rand.Intn(10) == 1 {
			diskWarn = "/data 98%;"
		}
		now := time.Now().Unix() - int64(rand.Intn(86400)*rand.Intn(3))
		for j := 0; j < 2; j++ {
			stat = Stat{
				Id:        Md5(fmt.Sprintf("%d", i), ""),
				TimeStamp: now + int64(j*60),
				HostName:  fmt.Sprintf("ns%d", i),
				OSRelease: "CentOS 6.5 32Bit",
				CPUName:   "Intel(R) Core(TM)2 Duo CPU P8600 @ 2.40GHz",
				CPUCore:   uint64(rand.Intn(100) + int(1)),
				Uptime:    uint64(rand.Intn(86400 * 365)),
				Load:      fmt.Sprintf("%d %d %d", rand.Intn(100), rand.Intn(100), rand.Intn(100)),
				CPURate:   rand.Float64() * 100,
				MemRate:   rand.Float64() * 100,
				SwapRate:  rand.Float64() * 100,
				DiskRate:  rand.Float64() * 100,
				DiskWarn:  diskWarn,
				DiskRead:  stat.DiskRead + uint64(rand.Intn(10000000)),
				DiskWrite: stat.DiskWrite + uint64(rand.Intn(10000000)),
				NetRead:   stat.NetRead + uint64(rand.Intn(10000000)),
				NetWrite:  stat.NetWrite + uint64(rand.Intn(10000000)),
			}
			data := simplejson.New(stat)
			result, _ := data.Dumps()
			err := httpSend(CLIENT_CONF.ServerUrl, CLIENT_CONF.ServerKey, result)
			assert.Nil(t, err)
		}
	}

	cmd := exec.Command("bash", "-c", "kill -9 `ps aux|grep [t]est.conf|awk '{print $2}'`")
	cmd.Run()

	os.Remove(confFile)
	os.Remove(testFile)
}
