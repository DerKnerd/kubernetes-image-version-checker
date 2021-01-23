// Uses Declarative syntax to run commands inside a container.
pipeline {
    triggers {
        pollSCM("*/5 * * * *")
    }
    agent {
        kubernetes {
            yaml '''
apiVersion: v1
kind: Pod
spec:
  imagePullSecrets:
    - name: dev-imanuel-jenkins-regcred
  volumes:
    - name: docker-sock
      hostPath:
        path: /var/run/docker.sock
  containers:
  - name: docker
    image: registry.imanuel.dev/library/docker:20.10.2
    command:
    - cat
    tty: true
    volumeMounts:
    - mountPath: /var/run/docker.sock
      name: docker-sock
'''
            defaultContainer 'docker'
        }
    }
    stages {
        stage('Push') {
            steps {
                container('docker') {
                    sh "docker build -t registry-hosted.imanuel.dev/tools/kubernetes-deployment-version-checker:$BUILD_NUMBER -f ./Dockerfile ."
                    sh "docker tag registry-hosted.imanuel.dev/tools/kubernetes-deployment-version-checker:$BUILD_NUMBER iulbricht/kubernetes-deployment-version-checker:$BUILD_NUMBER"

                    withDockerRegistry(credentialsId: 'nexus.imanuel.dev', url: 'https://registry-hosted.imanuel.dev') {
                        sh "docker push registry-hosted.imanuel.dev/tools/kubernetes-deployment-version-checker:$BUILD_NUMBER"
                    }
                    withDockerRegistry(credentialsId: 'hub.docker.com', url: '') {
                        sh "docker push iulbricht/kubernetes-deployment-version-checker:$BUILD_NUMBER"
                    }
                }
            }
        }
    }
}
