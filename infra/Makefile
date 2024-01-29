IMAGE_TAG := ${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${APP_NAME}

build:
	docker build --platform linux/amd64 -t ${IMAGE_TAG} ..

push: build
	gcloud auth configure-docker ${IMAGE_TAG} --quiet
	docker push ${IMAGE_TAG}

deploy: push
	gcloud run deploy ${REPO_NAME}-${APP_NAME} \
	  --image ${IMAGE_TAG} \
	  --region ${REGION} \
	  --project=${PROJECT_ID} \
	  --allow-unauthenticated \
	  --set-env-vars REDIS_URL=${REDIS_URL},GIN_MODE=${GIN_MODE},API_KEY=${API_KEY},KLAVIYO_API_KEY=${KLAVIYO_API_KEY}

run: build
	docker run -p 8080:8080 \
	  -e REDIS_URL=${REDIS_URL} \
	  -e GIN_MODE=release \
	  ${IMAGE_TAG}

