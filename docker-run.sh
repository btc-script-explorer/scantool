#! /bin/bash

# docker support requires:
# - access to /var/lib/docker/volumes/
# - Docker Engine installed
# - use of a config file to start scantool

# Make sure the included a config file and that it exists.
CONFIG_FILE=$1
if [ $# -lt 1 ]; then
	echo "Config file required."
	exit
fi

if [ ! -f $CONFIG_FILE ]; then
	echo "$CONFIG_FILE does not exist."
	exit
fi

# Check to make sure the user can create a volume and copy files into its data directory.
DOCKER_DIR="/var/lib/docker"
DOCKER_VOLUMES_DIR="$DOCKER_DIR/volumes"
VOLUME_DIR="$DOCKER_VOLUMES_DIR/scantool-config-dir"
VOLUME_DATA_DIR="$VOLUME_DIR/_data/"
if [ ! -d $VOLUME_DATA_DIR ]; then

	docker volume create scantool-config-dir > /dev/null

	if [ ! -d $DOCKER_DIR ]; then
		echo ""
		echo "$DOCKER_DIR does not exist. Is Docker Engine installed properly? Are you running a recent version of docker?"
		echo ""
		docker version
		exit
	fi

	SYSTEM_DOCKER_GROUP=`awk "/docker/" /etc/group`
	if [ ${#SYSTEM_DOCKER_GROUP} -eq 0 ]; then
		echo ""
		echo "Group docker does not exist. Is Docker Engine installed properly? Are you running a recent version of docker?"
		echo ""
		docker version
		exit
	fi

	DOCKER_OWNER=`stat -c "%U" $DOCKER_DIR`
	DOCKER_GROUP=`stat -c "%G" $DOCKER_DIR`

	echo ""
	echo "Write access to $VOLUME_DATA_DIR is required."
	echo "Run the following commands, then run this script again."
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
		echo "You will probably need to log out and log in again before the group membership takes effect."
	fi

	echo ""
	exit
fi

cp $CONFIG_FILE $VOLUME_DATA_DIR

# Create a container name.
VERSION=`cat ./VERSION`
CONTAINER_VERSION=`echo "$VERSION" | sed "s/\./_/g"`
CONTAINER_NAME="scantool_$CONTAINER_VERSION"

# Check to make sure the container is not already running.
# If it is running, inform the user.
RUNNING_CONTAINER=`echo \`docker container ls\` | awk "/$CONTAINER_NAME/"`
if [ ${#RUNNING_CONTAINER} -ne 0 ]; then
#	docker stop $CONTAINER_NAME > /dev/null
	echo "Container $CONTAINER_NAME is already running."
	echo "Run \"docker stop $CONTAINER_NAME\" to stop it, then run this script again."
	exit
fi

# If the container exists, start it up.
EXISTING_CONTAINER=`echo \`docker container ls -a\` | awk "/$CONTAINER_NAME/"`
if [ ${#EXISTING_CONTAINER} -ne 0 ]; then
	docker container start -i $CONTAINER_NAME
	exit
fi

# Load the images, create a new container and run it.
docker load -i scantool-$VERSION-docker-image.tar > /dev/null
docker run --name $CONTAINER_NAME -p 80:80 -v scantool-config-dir:/scantool-config-dir scantool:$VERSION

#docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' scantool:$VERSION
#docker inspect -f '{{.Id}}' scantool

