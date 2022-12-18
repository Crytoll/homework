# 作业要求

把我们的 httpserver 服务以 Istio Ingress Gateway 的形式发布出来。以下是你需要考虑的几点：

* 如何实现安全保证；
* 七层路由规则；
* 考虑 open tracing 的接入。

# 作业内容

## 安装 istio

```bash
cd  /tmp
wget https://github.com/istio/istio/releases/download/1.13.3/istio-1.13.3-linux-amd64.tar.gz
tar -zxvf istio-1.13.3-linux-amd64.tar.gz
mv istio-1.13.3 /usr/local/
cd /usr/local/
ln -s istio-1.13.3 istio
ln -sf /usr/local/istio/bin/istioctl /usr/local/sbin/
istioctl install --set profile=demo -y
kubectl get pod -n istio-system
kubectl get svc -n kube-system
```

## 启动应用服务

为 default 命名空间打上标签 istio-injection=enabled：

```bash
kubectl label namespace default istio-injection=enabled
```

使用 `kubectl` 部署应用：
```bash
kubectl apply -f deploy/
```

确认所有服务和 Pod 都已经正确定义和启动，并获取 istio-ingressgateway 的service IP：
```bash
kubectl get pod|grep httpserver
kubectl get svc,ep|grep httpserver
kubectl get svc -n istio-system
```

## 访问测试

```bash
curl https://httpserver.k8snb.com?user=crytoll
curl -H "user:crytoll" https://httpserver.k8snb.com?user=crytoll
```


## 安装 Jaeger

```bash
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.13/samples/addons/jaeger.yaml
```

## tracing 测试

部署 service0、service1及相关 istio 资源：

```bash
kubectl apply -f service0/specs/deployment.yaml
kubectl apply -f service1/specs/deployment.yaml
kubectl apply -f tracing/istio-specs.yaml
```

查看运行情况：
```bash
kubectl get deploy,pod,svc,virtualservice|egrep 'service[0|1]'
```


临时端口转发 jaeger service 暴露访问：
```bash
kubectl port-forward svc/tracing -n istio-system 8000:80 --address=0.0.0.0
```