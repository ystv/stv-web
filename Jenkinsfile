pipeline {
    agent any

    environment {
        REGISTRY_ENDPOINT = credentials('docker-registry-endpoint')
    }

    stages {
        stage('Update Components') {
            steps {
                sh 'docker pull golang:1.20.4-alpine3.16'
            }
        }
        stage('Build') {
            steps {
                sh 'docker build -t $REGISTRY_ENDPOINT/ystv/stv-web:$BUILD_ID .'
            }
        }
        stage('Registry Upload') {
            steps {
                sh 'docker push $REGISTRY_ENDPOINT/ystv/stv-web:$BUILD_ID' // Uploaded to registry
            }
        }
        stage('Deploy') {
            stages {
                stage('Staging') {
                    when {
                        branch 'master'
                        not {
                            expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ }
                        }
                    }
                    environment {
                        TARGET_PATH = credentials('moss-server-path-stv')
                    }
                    steps {
                        script {
                            sh 'docker pull $REGISTRY_ENDPOINT/ystv/stv-web:$BUILD_ID'
                            sh 'docker rm -f ystv-stv-web'
                            sh 'docker run -d -p 6691:6691 --name ystv-stv-web -v $TARGET_PATH/db:/db -v $TARGET_PATH/toml:/toml --restart=always $REGISTRY_ENDPOINT/ystv/stv-web:$BUILD_ID'
                            sh 'docker image prune -a -f --filter "label=site=stv-web"'
                        }
                    }
                }
                /*stage('Production') {
                    when {
                        expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ } // Checking if it is main semantic version release
                    }
                    environment {
                        APP_ENV = credentials('computing-env')
                        TARGET_SERVER = credentials('prod-server-address')
                        TARGET_PATH = credentials('prod-server-path')
                    }
                    steps {
                        sshagent(credentials : ['prod-server-key']) {
                            script {
                                sh 'rsync -av $APP_ENV deploy@$TARGET_SERVER:$TARGET_PATH/computing/.env'
                                sh '''ssh -tt deploy@$TARGET_SERVER << EOF
                                    docker pull $REGISTRY_ENDPOINT/ystv/computing:$BUILD_ID
                                    docker rm -f ystv-computing
                                    docker run -d -p 7075:7075 --env-file $TARGET_PATH/computing/.env --name ystv-computing --restart=always $REGISTRY_ENDPOINT/ystv/computing:$BUILD_ID
                                    docker image prune -a -f --filter "label=site=computing"
                                    exit 0
                                EOF'''
                            }
                        }
                    }
                }*/
            }
        }
    }
    post {
        success {
            echo 'Very cash-money'
        }
        failure {
            echo 'That is not ideal, cheeky bugger'
        }
        always {
            sh "docker image prune -f --filter label=site=stv-web --filter label=stage=builder" // Removing the local builder image
            sh 'docker image prune -a -f --filter "label=site=stv-web"' // remove old image
        }
    }
}
