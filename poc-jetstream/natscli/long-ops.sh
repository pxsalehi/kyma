#!/bin/zsh
declare -A pods
pods["pod1"]="./natscli bench --stream=bench1 --pub=2 --sub=10 --js --msgs=8640 --replicas=3 --size=10240 --storage=file --publishInterval=5000 bench-subj-1"
pods["pod2"]="./natscli bench --stream=bench2 --pub=2 --sub=10 --js --msgs=8640 --replicas=3 --size=10240 --storage=file --publishInterval=5000 bench-subj-2"
pods["pod3"]="./natscli bench --stream=bench3 --pub=2 --sub=10 --js --msgs=8640 --replicas=3 --size=10240 --storage=file --publishInterval=5000 bench-subj-3"

for key in ${(k)pods}; do
    PODNAME=${key}; BENCHCMD=${pods[${key}]}; cat nats-bench.taml | envsubst | kubectl apply -f-
done

