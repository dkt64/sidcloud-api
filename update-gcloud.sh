sh ./docker-image.sh
sh ./docker-tag-gcloud.sh
sh ./docker-push-gcloud.sh
gcloud compute instances reset sidcloud-1