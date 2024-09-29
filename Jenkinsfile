def TARGET_REGISTRY="registry.dcim.co"
def TARGET_IMAGE_TAG="dcim/${TARGET}-collector"
def OUTPUT_FILENAME="${TARGET}-collector.img"


pipeline {
    agent any

    stages {
        stage('Prepare') {
            when {
                expression {
                    return fileExists('jenkins-prepare.sh')
                }
            }
            steps {
                sh "sh jenkins-prepare.sh"
            }
        }
        stage('Build') {
            steps {
                sh "docker build -t ${TARGET_REGISTRY}/${TARGET_IMAGE_TAG} ./"
            }
        }
        stage('Release'){
            parallel { 
                stage('Archive script artifacts') {
                    when {
                        expression {
                            return fileExists('jenkins-artifacts.sh')
                        }
                    }
                    steps {
                        sh "TARGET_REGISTRY=\"${TARGET_REGISTRY}\" TARGET_IMAGE_TAG=\"${TARGET_IMAGE_TAG}\" sh jenkins-artifacts.sh"
                        archiveArtifacts artifacts: "build/*"
                    }
                }
                stage('Archive Artifacts'){
                    steps {
                        sh "docker save ${TARGET_REGISTRY}/${TARGET_IMAGE_TAG} | pigz > ${OUTPUT_FILENAME}.gz"
                        archiveArtifacts artifacts: "${OUTPUT_FILENAME}.gz"
                    }
                }
                stage('Push to server'){
                    steps {
                        sh "docker push ${TARGET_REGISTRY}/${TARGET_IMAGE_TAG}"
                    }
                }
            }
        }
    }
    post {
        always {
            deleteDir()
        }
    }
}