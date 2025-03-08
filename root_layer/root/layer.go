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

package root

import (
	"context"
	"time"

	"github.com/sjy-dv/coltt/core"
	"google.golang.org/grpc"
)

type RootCore struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Core   *core.Core
	S      *grpc.Server
}

type rpcLayer struct {
	Cproto *coreProtoConn
}

type coreProtoConn struct {
	rpcLayer
}

const (
	B  = 1
	KB = 1024
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

const DefaultMsgSize = 104858000 // 10mb
const DefaultKeepAliveTimeout = 10 * time.Second
const DefaultKeepAlive = 60 * time.Second
const DefaultEnforcementPolicyMinTime = 5 * time.Second

var UncaughtPanicError = "uncaught panic error: %v"
