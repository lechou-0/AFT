#!/bin/bash

#  Copyright 2019 U.C. Berkeley RISE Lab
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# Start the process.
if [[ "$ROLE" = "aft" ]]; then
  cd $AFT_HOME/cmd/aft

  go build
  ./aft
elif [[ "$ROLE" = "manager" ]]; then
  cd $AFT_HOME/cmd/gc
  go build

  REPLICA_IPS=`cat ../replicas.txt | awk 'BEGIN{ORS=","}1'`
  GC_IPS=`cat ../gcs.txt | awk 'BEGIN{ORS=","}1'`

  python3 ft-server.py &

  ./gc -replicaList $REPLICA_IPS -gcReplicaList $GC_IPS
elif [[ "$ROLE" = "bench" ]]; then
  cd $AFT_HOME/cmd/benchmark
  go build
  python3 benchmark_server.py
elif [[ "$ROLE" = "lb" ]]; then
  cd $GOPATH/src/k8s.io/klog
  git checkout v0.4.0

  cd $AFT_HOME/cmd/lb
  go build
  ./lb
elif [[ "$ROLE" = "gc" ]]; then
  cd $AFT_HOME/cmd/gc/server
  go build

  ./server
fi
