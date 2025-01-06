String registryEndpoint = 'registry.comp.ystv.co.uk'

def branch = env.BRANCH_NAME.replaceAll("/", "_")
def image
String imageName = "ystv/stv-web:${branch}-${env.BUILD_ID}"

pipeline {
  agent {
    label 'docker'
  }

  environment {
    REGISTRY_ENDPOINT = credentials('docker-registry-endpoint')
  }

  stages {
    stage('Build image') {
      steps {
        script {
          GIT_COMMIT_HASH = sh (script: "git log -n 1 --pretty=format:'%H'", returnStdout: true)
          docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
            image = docker.build(imageName, "--build-arg STV_WEB_VERSION_ARG=${env.BRANCH_NAME}-${env.BUILD_ID} --build-arg STV_WEB_COMMIT_ARG=${GIT_COMMIT_HASH} .")
          }
        }
      }
    }

    stage('Push image to registry') {
      steps {
        script {
          docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
            image.push()
            if (env.BRANCH_IS_PRIMARY) {
              image.push('latest')
            }
          }
        }
      }
    }

    stage('Deploy') {
      stages {
        stage('Development') {
          when {
            expression { env.BRANCH_IS_PRIMARY }
          }
          steps {
            build(job: 'Deploy Nomad Job', parameters: [
              string(name: 'JOB_FILE', value: 'stv-dev.nomad'),
              text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${imageName}")
            ])
          }
        }

        stage('Production') {
          when {
            // Checking if it is semantic version release.
            expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ }
          }
          steps {
            build(job: 'Deploy Nomad Job', parameters: [
              string(name: 'JOB_FILE', value: 'stv-prod.nomad'),
              text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${imageName}")
            ])
          }
        }
      }
    }
  }
}
