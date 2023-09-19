#! /bin/bash

SCRIPT_NAME=`basename "$0"`

# need a better way to read in parameters
# currently, there is a required order

# Make sure a config file was passed in and that it exists.
if [ $# -lt 1 ]; then
	echo ""
	echo "Config file required."
	echo "Usage: $SCRIPT_NAME <path-to-config-file> [ OPTIONS ]"
	echo ""
	echo "Options"
	echo "    --docker, -d: run the docker image"
	echo ""
	echo "To run on this host:"
	echo "    $SCRIPT_NAME <path-to-config-file>"
	echo ""
	echo "To run the docker image:"
	echo "    $SCRIPT_NAME <path-to-config-file> --docker"
	echo ""
	exit
fi

CONFIG_FILE=$1
if [ ! -f $CONFIG_FILE ]; then
	echo ""
	echo "$CONFIG_FILE does not exist."
	echo ""
	exit
fi

# copy the config file into the release directory
cp $CONFIG_FILE ./scantool.conf

# If we are not running a docker container, we can just start up the application and exit.
if [ $# -lt 2 ]; then
	./scantool --config-file=./scantool.conf
	exit
fi


####################
#                  #
#   docker setup   #
#                  #
####################

DOCKER_EXEC_FILE=`which docker`
if [ ${#DOCKER_EXEC_FILE} -eq 0 ]; then
	echo ""
	echo "Docker does not appear to be installed."
	echo "Please install the latest version of Docker Engine."
	echo ""
	exit
fi

# Create a container name.
VERSION=`cat ./VERSION`
CONTAINER_VERSION=`echo "$VERSION" | sed "s/\./_/g"`
CONTAINER_NAME="scantool_$CONTAINER_VERSION"

# Check to make sure the container is not already running.
# If it is running, inform the user.
RUNNING_CONTAINER=`echo \`docker container ls\` | awk "/$CONTAINER_NAME/"`
if [ ${#RUNNING_CONTAINER} -ne 0 ]; then
	echo ""
	echo "Container $CONTAINER_NAME is already running."
	echo "Run the following commands to continue:"
	echo ""
	echo "    docker stop $CONTAINER_NAME"
	echo "    ./$SCRIPT_NAME $1"
	echo ""
	exit
fi

# Delete any existing container of the same name.
EXISTING_CONTAINER=`echo \`docker container ls -a\` | awk "/$CONTAINER_NAME/"`
if [ ${#EXISTING_CONTAINER} -ne 0 ]; then
	docker container rm $CONTAINER_NAME > /dev/null
fi

# Load the images, create a new container and run it.
docker build -t scantool:$VERSION . > /dev/null
#docker build -t scantool:latest . > /dev/null

echo ""
echo ""
echo ""
echo "To stop:"
echo "    docker stop $CONTAINER_NAME"

docker run --name $CONTAINER_NAME -p 80:80 scantool:$VERSION

#docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' scantool:$VERSION
#docker inspect -f '{{.Id}}' scantool

