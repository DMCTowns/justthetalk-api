# Deploys portal to AWS repository
# ./deploy.sh [environment [tag ]]
# e.g. ./deploy.sh development latest

set -e

# Keep track of the last executed command
trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG

# Function to read environment variables
read_var() {
  if [ -z "$1" ]; then
    echo "Environment variable name is required"
    return
  fi

  local ENV_FILE='.env'
  if [ ! -z "$2" ]; then
    ENV_FILE="$2"
  fi

  local VAR=$(grep $1 "$ENV_FILE" | xargs)
  IFS="=" read -ra VAR <<< "$VAR"
  echo ${VAR[1]}
}

# Read in parameters
aws=$(read_var AWS_ACCOUNT_ID)
profile=$(read_var AWS_PROFILE)
profile=${profile:-'default'}
region=$(read_var AWS_REGION)
region=${region:-'eu-west-2'}
platform=$(read_var PLATFORM)
service=$(read_var SERVICE)
version=$(read_var APP_VERSION)

aws=$(echo $aws|tr -d '\r')

awsroot="${aws}.dkr.ecr.${region}.amazonaws.com"

echo ${awsroot}

# Get passed arguments
if [ $# -gt 0 ]; then
  environment=${1};
else
  environment='development';
fi
# Check validity of environment
if [[ ! $environment =~ ^(production|staging|development)$ ]]; then
  echo "Invalid second argument '${environment}'. Please use one of 'production', 'staging' or 'development'."
  exit 1
fi

if [ $# -gt 1 ]; then
  tag=${2};
else
  tag="latest";
  if [[ "$environment" == "production" ]]; then
    if [[ -n "$version" ]]; then
      tag="v${version}"
    fi
  fi
fi

# Log in to aws
aws ecr get-login-password --region ${region} --profile ${profile} | docker login --username AWS --password-stdin "${awsroot}"


# Create repository (if required)
status=0
aws ecr describe-repositories --repository-names ${service}-${environment} --region ${region} --profile ${profile} --registry-id ${aws} 2>&1 > /dev/null || status=$?
if [[ ! "${status}" -eq 0 ]]; then
  aws --region ${region} --profile ${profile} ecr create-repository --repository-name ${service}-${environment} 2>&1 > /dev/null
fi

# Build image
echo "Building image ${platform}/${service}"
docker build -t "${platform}/${service}" ./

# Tag image
docker tag "${platform}/${service}" "${platform}/${service}:${tag}"
echo "Tagging image ${awsroot}/${service}-${environment}:${tag}"
docker tag "${platform}/${service}" "${awsroot}/${service}-${environment}:${tag}"

# Push image
echo "Pushing image ${awsroot}/${service}-${environment}:${tag}"
docker push "${awsroot}/${service}-${environment}:${tag}"

# Deploy image to dev
if [ "${environment}" == 'development' ]; then
  if [ "${tag}" == 'latest' ]; then
    echo "Deploying service"
    aws --region ${region} --profile ${profile} ecs update-service --cluster "JTT-Production-1" --service "jtt-${service}-service-${environment}" --force-new-deployment >/dev/null
  fi
fi
# echo "Done"
exit 0