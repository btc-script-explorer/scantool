#! /bin/bash

# docker support requires:
# - access to /var/lib/docker/volumes/
# - Docker Engine installed
# - use of a config file to start scantool

CONFIG_FILE=$1
if [ $# -lt 1 ]; then
	echo "Config file required."
	exit
fi
if [ ! -f $CONFIG_FILE ]; then
	echo "$CONFIG_FILE does not exist."
	exit
fi

RUNNING_CONTAINER=`echo \`docker container ls -a\` | awk "/scantool/"`
if [ ${#RUNNING_CONTAINER} -ne 0 ]; then
	docker stop scantool > /dev/null
	docker container rm scantool:latest > /dev/null
fi

VERSION=`cat ./VERSION`
DOCKER_IMAGE_FILE="script-analytics-tool-$VERSION-docker-image.tar"
docker load -i $DOCKER_IMAGE_FILE

exit

docker volume create scantool-config-dir > /dev/null

DOCKER_DIR="/var/lib/docker"
DOCKER_VOLUMES_DIR="$DOCKER_DIR/volumes"
VOLUME_DIR="$DOCKER_VOLUMES_DIR/scantool-config-dir"
VOLUME_DATA_DIR="$VOLUME_DIR/_data/"
if [ ! -d $VOLUME_DATA_DIR ]; then

	if [ ! -d $DOCKER_DIR ]; then
		echo ""
		echo "$DOCKER_DIR does not exist. Is Docker Engine installed properly? Are you running a recent version?"
		echo ""
		docker version
		exit
	fi

	SYSTEM_DOCKER_GROUP=`awk "/docker/" /etc/group`
	if [ ${#SYSTEM_DOCKER_GROUP} -eq 0 ]; then
		echo ""
		echo "Group docker does not exist. Is Docker Engine installed properly? Are you running a recent version?"
		echo ""
		docker version
		exit
	fi

	DOCKER_OWNER=`stat -c "%U" $DOCKER_DIR`
	DOCKER_GROUP=`stat -c "%G" $DOCKER_DIR`

	echo ""
	echo "Access to $VOLUME_DATA_DIR is required."
	echo "Run the following commands."
	echo ""
	if [ "$DOCKER_GROUP" != "docker" ]; then
		echo "sudo chown $DOCKER_OWNER:docker $DOCKER_DIR"
		echo "sudo chown $DOCKER_OWNER:docker $DOCKER_VOLUMES_DIR"
		echo "sudo chown $DOCKER_OWNER:docker $VOLUME_DIR"
		echo "sudo chown $DOCKER_OWNER:docker $VOLUME_DATA_DIR"
	fi

	echo "sudo chmod g+rx $DOCKER_DIR"
	echo "sudo chmod g+rx $DOCKER_VOLUMES_DIR"
	echo "sudo chmod g+rx $VOLUME_DIR"
	echo "sudo chmod g+rwx $VOLUME_DATA_DIR"

	IN_DOCKER=`echo \`groups\` | awk "/docker/"`
	if [ ${#IN_DOCKER} -eq 0 ]; then
		echo "sudo usermod -a -G docker $USER"
	fi

	echo "./docker-run.sh $1"

	if [ ${#IN_DOCKER} -eq 0 ]; then
		echo ""
		echo "You will probably need to log out and log in again before the group membership takes effect."
	fi

	echo ""
	exit
fi

cp $CONFIG_FILE $VOLUME_DATA_DIR
docker run --name scantool -p 80:80 -v scantool-config-dir:/scantool-config-dir scantool

echo ""
echo "To stop scantool run:"
echo "docker stop scantool"

IP_ADDRESS=`docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' scantool`
echo ""
echo "Web site is at http://$IP_ADDRESS/web/"
echo "REST API is at http://$IP_ADDRESS/rest/v1/"
echo ""

#docker inspect -f '{{.Id}}' scantool

