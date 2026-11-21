# Local deployment
Source ~/.profile and run `bash deploy.sh`. The build now writes dist/package.txt;
deploy-out/package.txt should receive the same content. Authentication is
supplied through DEPLOY_TOKEN. This script copies local files only.
