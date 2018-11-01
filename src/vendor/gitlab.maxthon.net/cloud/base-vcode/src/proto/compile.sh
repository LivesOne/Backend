#!/usr/bin/env bash
#! Goland 命令行可用
protoc vcode.proto --go_out=plugins=grpc:.
