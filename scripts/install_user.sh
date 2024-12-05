#/bin/sh
ctr image del docker.io/flycash/webook:v0.0.1
docker rmi flycash/webook:v0.0.1
docker builder prune -a -f
kubectl delete deploy user-service

cd /usr/local/go_project/w-book/app/user-service && go build -o user-service /usr/local/go_project/w-book/app/user-service/cmd/user.go && docker build -t flycash/webook:v0.0.1 .

mkdir -p /images
docker save -o /images/user-service.tar flycash/webook:v0.0.1
ctr -n=k8s.io i import /images/user-service.tar

kubectl apply -f /usr/local/go_project/w-book/helm/deployment.yaml

rm -rf

