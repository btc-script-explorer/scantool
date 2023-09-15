#! /bin/bash

echo ""

CONTAINER_ID=`grep -m 1 '/var/lib/docker/containers/' /proc/self/mountinfo | awk -F/var/lib/docker/containers/ '{print $2}' | awk -F/ '{print $1}'`
if [ ${#CONTAINER_ID} -gt 0 ]; then
	echo " Docker Container ID = $CONTAINER_ID"
else
	echo "Failed to get the container ID."
	exit
fi

SHORT_ID=`echo $CONTAINER_ID | cut -c 1-12`
if [ ${#SHORT_ID} -lt 12 ]; then
	echo "Failed to get the container short ID."
	exit
fi

IP_ADDRESS=`awk "/$SHORT_ID/" /etc/hosts | awk '{print $1}'`
if [ ${#IP_ADDRESS} -gt 0 ]; then
	echo "Container IP Address = $IP_ADDRESS"
else
	echo "Failed to get the container IP address."
	exit
fi

echo ""
echo "Web Interface: http://$IP_ADDRESS/web/"
echo "     REST API: http://$IP_ADDRESS/rest/v1/"

./scantool --addr=$IP_ADDRESS --port=80 --config-file=./scantool-config-dir/scantool.conf

