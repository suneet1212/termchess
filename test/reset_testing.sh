#!/usr/bin/env bash

# check if we have sudo access, if not exit
if [[ $EUID -ne 0 ]]; then
    echo "run as root: sudo $0 $*" >&2
    exit 1
fi

docker compose down

ip netns del internet
ip netns del r1
ip netns del r2

