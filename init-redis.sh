#!/bin/bash
sysctl -w vm.overcommit_memory=1
redis-server --port ${REDIS_PORT} --appendonly yes