# 作业要求：编写 Kubernetes 部署脚本将 httpserver 部署到 Kubernetes 集群，以下是你可以思考的维度。
* 优雅启动: deploy/deployment.yaml postStart
* 优雅终止: deploy/deployment.yaml preStop
* 资源需求和 QoS 保证: deploy/deployment.yaml resources
* 探活: deploy/deployment.yaml livenessProbe
* 日常运维需求，日志等级: httpserver/conf/config.yaml log
* 配置和代码分离: httpserver/main.go

`kubectl create -f deploy`

![k8s资源](/Users/chenying/go/src/github.com/homework/module8/readme.png)
