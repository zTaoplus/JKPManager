#!/bin/bash

# 获取所有开头为 `jovyan-` 的 pod ID
kernel_ids=$(kubectl get pod -n tablegpt-kernels -o=jsonpath='{range .items[*]}{@.metadata.name}{"\n"}{end}' | grep '^jovyan-' | sed 's/^jovyan-//')

# 循环遍历每个 kernel_id，并发送 DELETE 请求
for kernel_id in $kernel_ids; do
    echo "Deleting kernel: $kernel_id"
    curl -X DELETE http://127.0.0.1:8888/api/kernels/$kernel_id
done
