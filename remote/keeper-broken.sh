#!/bin/bash

nodes=(9801 9802 9803)

while True
do
    leader=0
    for node in ${nodes[@]}  
    do
        isleader=`curl -s  http://localhost:${node}/leader`
        if [[ $isleader == "true" ]]
        then
            leader=$(($leader+1))
            echo $node $leader
        fi
    done

    if (($leader > 1))
    then
        echo "it's broken"
        exit 1
    fi
    sleep 1
done
