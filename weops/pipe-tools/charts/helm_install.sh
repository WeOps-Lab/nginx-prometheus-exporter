#!/bin/bash
# 部署监控对象
object=nginx
object_versions=("1.25.1" "1.24" "1.23" "1.22" "1.21" "1.20" "1.19" "1.18" "1.17" "1.16" "1.14" "1.12" "1.10.2-r4" "1.9.15-1
" "1.8.0-4")

value_file="bitnami_values.yaml"


# 设置起始端口号
port=30080

for version in "${object_versions[@]}"; do
    version_suffix="v$(echo "$version" | grep -Eo '[0-9]{1,2}\.[0-9]{1,2}' | tr '.' '-')"
    helm install $object-standalone-$version_suffix --namespace $object --create-namespace -f ./values/$value_file ./$object \
    --set image.tag=$version \
    --set commonLabels.object_version=$version_suffix \
    --set service.nodePorts.http=$port
    ((port++))

    sleep 1
done

