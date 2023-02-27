# Rebuilds the index for JTT's API
# ./create_index.sh <tag ID>

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
  task=${1};
else
  echo "Please supply a task ID."
  exit 1
fi

aws ecs execute-command  \
    --region $region \
    --profile ${profile} \
    --cluster JTT-Production-1 \
    --task $task \
    --container jtt-$service \
    --interactive \
    --command "/bin/bash"
exit 0