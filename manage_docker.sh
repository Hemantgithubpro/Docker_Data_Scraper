#!/bin/bash

# Go to your docker-compose project directory
cd /home/hemant/StockMarketData/exp1/ || exit

case "$1" in
    start)
        docker compose up -d
        ;;
    stop)
        docker compose stop
        ;;
    *)
        echo "Usage: $0 {start|stop}"
        exit 1
        ;;
esac
