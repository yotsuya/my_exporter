#!/bin/sh

/openio-docker-init.sh > /dev/null 2>&1 &

/my_exporter &

# TODO: ちゃんとしたテスト
for i in `seq 5`
do
    sleep 5
    echo "## ${i}"
    curl -s http://localhost:9999/metrics > /dev/null
done
