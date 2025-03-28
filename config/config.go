// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package config

import (
	"flag"
	"fmt"
	"os"
)

const ClusterGroup = "coltt-stream"
const NodeNamePrefix = "coltt-node"

var ClusterAddrFlag = flag.String("cluster-addr", "", "Cluster listening address")
var ClusterPeersFlag = flag.String("cluster-peers", "", "Comma separated list of clusters")
var LeafServerFlag = flag.String("leaf-servers", "", "Comma separated list of leaf servers")
var DataRootDir = os.TempDir()

type ConfigMap struct {
	CacheKey string `toml:"cache_key"`
	NodeID   uint64 `toml:"node_id"`
	// when false, detect cluster mode
	// use single instance, must standalone=true
	Standalone bool      `toml:"standalone"`
	JetStream  JetStream `toml:"jetstream"`
	RootLayer  RootLayer `toml:"rootlayer"`
}

type JetStream struct {
	URLs                 []string `toml:"urls"`
	SubjectPrefix        string   `toml:"subject_prefix"`
	StreamPrefix         string   `toml:"stream_prefix"`
	ServerConfigFile     string   `toml:"server_config"`
	SeedFile             string   `toml:"seed_file"`
	CredsUser            string   `toml:"user_name"`
	CredsPassword        string   `toml:"user_password"`
	CAFile               string   `toml:"ca_file"`
	CertFile             string   `toml:"cert_file"`
	KeyFile              string   `toml:"key_file"`
	BindAddress          string   `toml:"bind_address"`
	ConnectRetries       int      `toml:"connect_retries"`
	ReconnectWaitSeconds int      `toml:"reconnect_wait_seconds"`
}

type RootLayer struct {
	BindAddress              string `toml:"bind_address"`
	ProfAddr                 string `toml:"prof_addr"`
	KeepAliveTimeOut         int    `toml:"keep_alive_time_out"`
	KeepAliveTime            int    `toml:"keep_alive_time"`
	EnforcementPolicyMinTime int    `toml:"enforcement_policy_min_time"`
	PemFile                  string `toml:"pem_file"`
	KeyFile                  string `toml:"key_file"`
	MaxRecvMsgSize           int    `toml:"max_recv_msg_size"`
	MaxSendMsgSize           int    `toml:"max_send_msg_size"`
}

var Config = &ConfigMap{
	CacheKey: "22ENpk1CTyMsbKlkATzRPydsrZRDu657mltVvAQSMJc=",
	NodeID:   0,
	JetStream: JetStream{
		URLs:                 []string{},
		SubjectPrefix:        "coltt-change-log",
		StreamPrefix:         "coltt-changes",
		ServerConfigFile:     "",
		SeedFile:             "",
		CredsPassword:        "",
		CredsUser:            "",
		BindAddress:          ":-1",
		ConnectRetries:       5,
		ReconnectWaitSeconds: 2,
	},
	RootLayer: RootLayer{
		BindAddress:              ":50051",
		ProfAddr:                 ":10224",
		KeepAliveTimeOut:         0,
		KeepAliveTime:            0,
		EnforcementPolicyMinTime: 0,
		PemFile:                  "",
		KeyFile:                  "",
		MaxRecvMsgSize:           0,
		MaxSendMsgSize:           0,
	},
}

func (c *ConfigMap) NodeName() string {
	return fmt.Sprintf("%s-%d", NodeNamePrefix, c.NodeID)
}
