#! /bin/bash

################################################
#                                              #
#   docker files are currently not supported   #
#                                              #
################################################

CONTAINER_ID=`grep -m 1 '/var/lib/docker/containers/' /proc/self/mountinfo | awk -F/var/lib/docker/containers/ '{print $2}' | awk -F/ '{print $1}'`
if [ ${#CONTAINER_ID} -eq 0 ]; then
	echo ""
	echo "Failed to get the container ID."
	echo ""
	exit
fi

SHORT_ID=`echo $CONTAINER_ID | cut -c 1-12`
if [ ${#SHORT_ID} -lt 12 ]; then
	echo ""
	echo "Failed to get the container short ID."
	echo ""
	exit
fi

IP_ADDRESS=`awk "/$SHORT_ID/" /etc/hosts | awk '{print $1}'`
if [ ${#IP_ADDRESS} -eq 0 ]; then
	echo ""
	echo "Failed to get the container IP address based on container ID $CONTAINER_ID, shortened to $SHORT_ID."
	cat /etc/hosts
	echo ""
	exit
fi

./scantool --addr=$IP_ADDRESS --port=80 --config-file=./scantool.conf

