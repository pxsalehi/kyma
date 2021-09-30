#!/bin/zsh

declare -A pods
pods["pod1"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app1.bench1.subj.v1"
pods["pod2"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app2.bench2.subj.v1"
pods["pod3"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app3.bench3.subj.v1"
pods["pod4"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app4.bench1.subj.v1"
pods["pod5"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app5.bench2.subj.v1"
pods["pod6"]="./natscli bench --stream=default --pub=2 --sub=10 --js --msgs=1000000 --replicas=3 --size=512  --no-progress --storage=file default.app6.bench3.subj.v1"

for key in ${(k)pods}; do
    PODNAME=${key}; BENCHCMD=${pods[${key}]}; cat nats-bench.taml | envsubst | kubectl apply -f-
done

