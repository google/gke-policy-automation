#!/bin/bash

# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The script populates Artifact Registry with a GKE Policy Automation image
# and creates the Cloud Run Job (currenty not available in Terraform).

WHT='\033[1;37m'
RED='\033[1;31m'
NC='\033[0m'
VARS=( "REGION" )

varCheck() {
  if [ -z "$REGION" ]; then
    echo -e "${RED}[ERROR] GKE_PA_REGION variable is not set${NC}"
    return 1
  fi
  if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}[ERROR] GKE_PA_PROJECT_ID variable is not set${NC}"
    return 1
  fi
  if [ -z "$JOB_NAME" ]; then
    echo -e "${RED}[ERROR] GKE_PA_JOB_NAME variable is not set${NC}"
    return 1
  fi
  if [ -z "$SA_EMAIL" ]; then
    echo -e "${RED}[ERROR] GKE_PA_SA_EMAIL variable is not set${NC}"
    return 1
  fi
  if [ -z "$SECRET_NAME" ]; then
    echo -e "${RED}[ERROR] GKE_PA_SECRET_NAME variable is not set${NC}"
    return 1
  fi
}

cmdCheck() {
  if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}[ERROR] gcloud command could not be found${NC}"
    return 1
  fi
  if ! command -v docker &> /dev/null; then
    echo -e "${RED}[ERROR] docker command could not be found${NC}"
    return 1
  fi
  return 0
}

preCheck() {
  varCheck
  if [ $? -ne 0 ]; then
    return 1
  fi
  cmdCheck
  if [ $? -ne 0 ]; then
    return 1
  fi
  return 0
}

runAndCheck() {
  if [ $# -lt 1 ]; then
    return 1
  fi
  eval "$1"
  if [ $? -ne 0 ]; then
    exit 1
  fi
  return 0
}

preCheck
if [ $? -ne 0 ]; then
  exit 1
fi

echo -e "${WHT}[INFO] Pulling GKE Policy Automation docker image${NC}"
runAndCheck "docker pull ghcr.io/google/gke-policy-automation:latest"

echo -e "${WHT}[INFO] Configuring docker credential helper${NC}"
runAndCheck "gcloud auth configure-docker ${GKE_PA_REGION}-docker.pkg.dev"

echo -e "${WHT}[INFO] Pushing GKE Policy Automation image to the Artifact Registry${NC}"
runAndCheck "docker tag ghcr.io/google/gke-policy-automation:latest ${GKE_PA_REGION}-docker.pkg.dev/${GKE_PA_PROJECT_ID}/gke-policy-automation/gke-policy-automation:latest"
runAndCheck "docker push ${REGION}-docker.pkg.dev/${PROJECT_ID}/gke-policy-automation/gke-policy-automation:latest"

echo -e "${WHT}[INFO] Creating Cloud Run Job${NC}"
runAndCheck "gcloud beta run jobs create ${GKE_PA_JOB_NAME} \
  --image ${GKE_PA_REGION}-docker.pkg.dev/${GKE_PA_PROJECT_ID}/gke-policy-automation/gke-policy-automation:latest \
  --command=/gke-policy,check \
  --args=-c,/etc/secrets/config.yaml \
  --set-secrets /etc/secrets/config.yaml=${GKE_PA_SECRET_NAME}:latest \
  --service-account=${GKE_PA_SA_EMAIL} \
  --set-env-vars=GKE_POLICY_LOG=INFO \
  --region=${GKE_PA_REGION} \
  --project=${GKE_PA_PROJECT_ID}"

if [ $? -eq 0 ]; then
  echo -e "${WHT}[INFO] Script was executed successfuly${NC}"
else
  echo -e "${RED}[ERROR] Script was executed with errors${NC}"
fi
