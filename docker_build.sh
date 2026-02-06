PROJECT_NAME="pascal"
DOCKER_IMAGE="vincbrod/pascal"
VERSION=$(cat version.txt)

echo "Building $PROJECT_NAME docker image..."
sudo docker build -t $DOCKER_IMAGE:latest -t $DOCKER_IMAGE:$VERSION .
echo "$PROJECT_NAME docker image done"
