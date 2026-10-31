/* groovylint-disable CompileStatic, NestedBlockDepth, UnnecessaryGString */
pipeline {
  agent {
    kubernetes {
      yaml '''
        apiVersion: v1
        kind: Pod
        spec:
          containers:
          - name: buildkit
            image: moby/buildkit:v0.31.0
            securityContext:
              privileged: true
      '''
    }
  }
  options {
    skipDefaultCheckout(true)
  }
  stages {
    stage('Checkout') {
      steps {
        checkout([
          $class: 'GitSCM',
          branches: scm.branches,
          userRemoteConfigs: scm.userRemoteConfigs,
          extensions: [[
            $class: 'SubmoduleOption',
            disableSubmodules: false,
            recursiveSubmodules: true,
            parentCredentials: true,
            trackingSubmodules: false,
            reference: '',
            timeout: 30
          ]]
        ])
      }
    }
    stage('Build') {
      environment {
        REGISTRY = "${env.GHCR_REGISTRY}"
        REPOSITORY = 'auth'
      }
      steps {
        container('buildkit') {
          script {
            String dockerConfig = '/tmp/registry-creds'
            String branch = env.BRANCH_NAME.replace('/', '-')
            String hash = sh(script: '''
              git config --global --add safe.directory "$WORKSPACE"
              git rev-parse --short=10 HEAD
            ''', returnStdout: true).take(10)

            withCredentials([usernamePassword(
              credentialsId: 'ghcr-creds',
              usernameVariable: 'USERNAME',
              passwordVariable: 'TOKEN'
            )]) {
              sh """
                export BUILDKIT_HTTP_TIMEOUT=300
                export DOCKER_CONFIG=${dockerConfig}
                mkdir -p ${dockerConfig}

                AUTH=\$(printf '%s:%s' '${USERNAME}' '${TOKEN}' | base64 -w 0)
                echo '{"auths":{"${REGISTRY}":{"auth":"'\${AUTH}'"}}}' > ${dockerConfig}/config.json

                buildctl build \
                  --frontend dockerfile.v0 \
                  --local context=. \
                  --local dockerfile=. \
                  --output type=image,name=${REGISTRY}/${REPOSITORY}:latest,push=true \
                  --output type=image,name=${REGISTRY}/${REPOSITORY}:${branch}-latest,push=true \
                  --output type=image,name=${REGISTRY}/${REPOSITORY}:${hash},push=true
              """
            }
          }
        }
      }
    }
  }
}
