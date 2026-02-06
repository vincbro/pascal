PROJECT_NAME="pascal"
DOCKER_IMAGE="vincbrod/pascal"
VERSION=$(cat version.txt)

echo "Pushing $PROJECT_NAME:$VERSION docker image..."
sudo docker push $DOCKER_IMAGE:latest
sudo docker push $DOCKER_IMAGE:$VERSION
echo "$PROJECT_NAME docker image pushed"
