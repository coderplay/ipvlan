#!/bin/sh

touch /run/docker/plugins/ipvlan.sock
docker-compose up -d