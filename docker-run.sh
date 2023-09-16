#! /bin/bash

# docker support requires:
# - access to /var/lib/docker/volumes/
# - Docker Engine installed
# - use of a config file to start scantool

# Make sure a config file was passed in and that it exists.
CONFIG_FILE=$1
if [ $# -lt 1 ]; then
	echo "Config file required."
	exit
fi

if [ ! -f $CONFIG_FILE ]; then
	echo "$CONFIG_FILE does not exist."
	exit
fi

DOCKER_EXEC_FILE=`which docker`
if [ ${#DOCKER_EXEC_FILE} -eq 0 ]; then
	echo ""
	echo "Docker does not appear to be installed."
	echo "Please install the latest version of Docker Engine."
	echo ""
	exit
fi

VOLUME_NAME="scantool-config-dir"

# If the user has the right permissions and software installed, we can create a volume and copy files into its data directory.
EXISTING_VOLUME=`echo \`docker volume ls\` | awk "/$VOLUME_NAME/"`
if [ ${#EXISTING_VOLUME} -eq 0 ]; then
	docker volume create $VOLUME_NAME > /dev/null
fi

DOCKER_DIR="/var/lib/docker"
DOCKER_VOLUMES_DIR="$DOCKER_DIR/volumes"
VOLUME_DIR="$DOCKER_VOLUMES_DIR/$VOLUME_NAME"
VOLUME_DATA_DIR="$VOLUME_DIR/_data/"
VOLUME_CONFIG_FILE="$VOLUME_DATA_DIR/scantool.conf"

cp $CONFIG_FILE VOLUME_CONFIG_FILE
if [ ! -f $VOLUME_CONFIG_FILE ]; then

	echo ""
	echo "Failed to pass the config file to the docker container."
	echo "There could be several causes of this problem."

	echo ""
	echo "Is the docker service running?"

	DOCKER_OWNER=""
	DOCKER_GROUP=""
	DIRS_TO_CHECK=("$DOCKER_DIR" "$DOCKER_VOLUMES_DIR" "$VOLUME_DIR" "$VOLUME_DATA_DIR")
	for D in ${DIRS_TO_CHECK[@]}; do
		if [ -d "$D" ]; then
			DIR_OWNER=`stat -c "%U" $D`
			DIR_GROUP=`stat -c "%G" $D`

			if [ ${#DOCKER_OWNER} -eq 0 ]; then
				DOCKER_OWNER=$DIR_OWNER
			fi

			if [ ${#DOCKER_GROUP} -eq 0 ]; then
				DOCKER_GROUP=$DIR_GROUP
			fi

			# if these directories are owned by multiple users and/or groups, do not proceed
#			if [[ ( "$DIR_OWNER" != "$DOCKER_OWNER" ) || ( "$DIR_GROUP" != "$DOCKER_GROUP" ) || (( "$DOCKER_OWNER" != "root" ) && ( "$DOCKER_GROUP" == "root" )) ]]; then
if [ "1" == "2" ]; then
				echo ""
				echo "The file permissions for your installation of docker have been modified since installation."
				echo "It appears as though you are running a custom configuration of docker."
				echo "The administrator of the system must decide what is the best way to give user $USER the necessary permissions without impacting other users."
				echo "Read and write permissions are required for directory $VOLUME_DATA_DIR."
				echo ""
				exit
			fi
		else
			echo "$D is not readable by user $USER."
		fi
	done

DOCKER_GROUP="root"

	SYSTEM_DOCKER_GROUP=`awk "/docker/" /etc/group`
	if [[ ( "$DOCKER_OWNER" = "root" ) && ( "$DOCKER_GROUP" = "root" ) ]]; then
		if [ ${#SYSTEM_DOCKER_GROUP} -ne 0 ]; then
			DOCKER_GROUP="docker"
		else
			echo ""
			echo "No docker group exists and the root user owns the docker files."
			echo "An administrator must determine the best way to give user $USER the necessary permissions."
			echo ""
			exit
		fi
	fi

echo "$DOCKER_OWNER $DOCKER_GROUP"
#exit

	echo ""
	echo "The commands below MIGHT give you the permissions required to run the scantool docker image."
	echo "Make sure to CHECK WITH THE ADMINISTRATOR OF THE SYSTEM BEFORE RUNNING THESE COMMANDS."
	echo ""

	for D in ${DIRS_TO_CHECK[@]}; do
		echo "    sudo chown $DOCKER_OWNER:$DOCKER_GROUP $D"
	done

	for D in ${DIRS_TO_CHECK[@]}; do
		if [ "$D" == "$VOLUME_DATA_DIR" ]; then
			echo "    sudo chmod g+rwx $D"
		else
			echo "    sudo chmod g+rx $D"
		fi
	done

	if [ "$DOCKER_GROUP" == "docker" ]; then
		IN_DOCKER=`echo \`groups\` | awk "/$DOCKER_GROUP/"`
		if [ ${#IN_DOCKER} -eq 0 ]; then
			echo "    sudo usermod -a -G $DOCKER_GROUP $USER"
		fi
	fi

	SCRIPT_NAME=`basename "$0"`
	echo "    ./$SCRIPT_NAME $1"

	if [[ ( "$DOCKER_GROUP" == "docker" ) && ( ${#IN_DOCKER} -eq 0 ) ]]; then
		echo ""
		echo "    NOTE: You might need to log out and log in again before the group membership takes effect."
	fi

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
#	docker stop $CONTAINER_NAME > /dev/null
	echo "Container $CONTAINER_NAME is already running."
	echo "Run \"docker stop $CONTAINER_NAME\" to stop it, then run this script again."
	exit
fi

echo ""
echo "To stop: docker stop $CONTAINER_NAME"

EXISTING_CONTAINER=`echo \`docker container ls -a\` | awk "/$CONTAINER_NAME/"`
if [ ${#EXISTING_CONTAINER} -ne 0 ]; then
	# Start the existing container.
	docker container start -i $CONTAINER_NAME
else
	# Load the images, create a new container and run it.
	docker load -i scantool-$VERSION-docker-image.tar > /dev/null
	docker run --name $CONTAINER_NAME -p 80:80 -v $VOLUME_NAME:/$VOLUME_NAME scantool:$VERSION
fi

#docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' scantool:$VERSION
#docker inspect -f '{{.Id}}' scantool

