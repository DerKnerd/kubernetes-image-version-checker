# Kubernetes Image Version Checker
The Kubernetes Image Version Checker (k8s IVC) is small go binary, that checks if any of your Deployments, DaemonSets, StatefulSets and CronJobs in your cluster have a new version for the container images available. Currently, only images on Dockerhub are supported. It is possible, though, to configure a custom registry host, which will be stripped away.

It notifies the configured email address when new versions are available.

## Installation
You can install the k8s IVC as cronjob type in k8s or download this repo and run the binary by hand. The container image is hosted on hub.docker.com under the name `iulbricht/kubernetes-deployment-version-checker`.

## Configuration
There are a few configuration options. These option control the mailing system and a few image related options.

Variable               | Description
---------------------- | ------
`MAILING_TO`           | The mail address the updates should be sent to
`MAILING_FROM`         | The mail address sending the updates
`MAILING_USERNAME`     | The username for the mail server
`MAILING_PASSWORD`     | The password for the mail server
`MAILING_HOST`         | The host of the mail server
`MAILING_PORT`         | The port of the mail server
`IGNORE_NAMESPACES`    | A comma separated list of namespaces to skip. When using microk8s recommend namespaces to exclude are `kube-system`, `kube-public`, `ingress` and `kube-node-lease`
`CUSTOM_REGISTRY_HOST` | The host of a proxy registry, like Sonatype Nexus. This host is automatically removed from the images
`MODE`                 | If set to `out` the configuration from `~/.kube-config` will be used, if left unset the kubernetes secret mounted at `/var/run/secrets/kubernetes.io/serviceaccount`

## Example k8s Cronjob
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: kubernetes-deployment-version-checker
spec:
  schedule: "0 0 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: kubernetes-deployment-version-checker
              image: iulbricht/kubernetes-deployment-version-checker:2
              env:
                - name: MAILING_TO
                  value: notify@example.com
                - name: MAILING_FROM
                  value: noreply@example.com
                - name: MAILING_USERNAME
                  value: noreply@example.com
                - name: MAILING_PASSWORD
                  value: password
                - name: MAILING_HOST
                  value: mail.example.com
                - name: MAILING_PORT
                  value: "587"
                - name: IGNORE_NAMESPACES
                  value: kube-system,kube-public,ingress,kube-node-lease
                - name: CUSTOM_REGISTRY_HOST
                  value: registry.example.com
          restartPolicy: OnFailure
```

## License
Like all other projects I create, the k8s IVC is distributed under the MIT License.