#!/bin/sh -l
set -x

/openio-docker-init.sh &
PID=$!

# TODO: openioの起動完了をチェックする方法調べる
sleep 60

/my_exporter

kill $PID
wait $PID
