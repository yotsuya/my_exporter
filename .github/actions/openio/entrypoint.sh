#!/bin/sh -l
set -x

/openio-docker-init.sh &
PID=$!

# TODO: openioの起動完了をチェックする方法調べる
sleep 45

/my_exporter

# OpenIOコンテナイメージのバージョンにもよるが、単にopenio-docker-init.shをkillするだけでは
# gracefulに停止しない場合がある。ここではコンテナは使い捨てなので実用上は問題ないのだが、
# 画面がエラー出力であふれるのが嫌なので、先にgridinit_cmd stopを実行しておく。
gridinit_cmd stop
kill $PID
wait $PID
