#! /bin/bash

echo ""

CONTAINER_ID=`grep -m 1 '/var/lib/docker/containers/' /proc/self/mountinfo | awk -F/var/lib/docker/containers/ '{print $2}' | awk -F/ '{print $1}'`
if [ ${#CONTAINER_ID} -gt 0 ]; then
	echo "Container ID = $CONTAINER_ID"
else
	echo "Failed to get the container ID."
	exit
fi

SHORT_ID=`echo $CONTAINER_ID | cut -c 1-12`
if [ ${#SHORT_ID} -gt 0 ]; then
	echo "    Short ID = $SHORT_ID"
else
	echo "Failed to get the container short ID."
	exit
fi

IP_ADDRESS=`awk "/$SHORT_ID/" /etc/hosts | awk '{print $1}'`
if [ ${#IP_ADDRESS} -gt 0 ]; then
	echo "  IP Address = $IP_ADDRESS"
else
	echo "Failed to get the container IP address."
	exit
fi

./scantool --addr=$IP_ADDRESS --port=80 --config-file=./scantool-config-dir/scantool.conf

