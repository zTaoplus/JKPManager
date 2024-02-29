# JKPManager
Jupyter kernels Pool manager


A http server for jupyter kernels pool management


### ENV
| NAME                    | Default Value               | Description                                              |
| ----------------------- | --------------------------- | -------------------------------------------------------- |
| JKP_EG_ENDPOINT         | http://127.0.0.1:8888       | jupyter server enterprise gateway.                       |
| JKP_MAX_PENDING_KERNELS | 10                          | The maximum number of Jupyter kernels idle.              |
| JKP_NFS_VOLUME_SERVER   | 10.0.0.29                   | NFS SERVER HOST                                          |
| JKP_NFS_MOUNT_PATH      | /data/tablegpt-test/shared/ | NFS MOUNTED PATH                                         |
| JKP_WORKING_DIR         | /mnt/shared/                | Jupyter kernel Working Dir PATH                          |
| JKP_KERNEL_IMAGE        | elyra/kernel-py:3.2.2       | Jupyter Kernel Running Image                             |
| JKP_KERNEL_NAMESPACE    | tablegpt-kernels            | Jupyter Kernel Running namespace in k8s                  |
| JKP_SERVER_PORT         | 8080                        | JKP Manager Http Server Running Port                     |
| JKP_ACTIVATION_INTERVAL | 1200                        | The activation time interval of the kernel WS connection |
| JKP_REDIS_HOST          | 127.0.0.1                   | Redis HOST                                               |
| JKP_REDIS_PORT          | 6379                        | Redis Port                                               |
| JKP_REDIS_DB            | 0                           | Redis DB                                                 |
| JKP_REDIS_KEY           | tablegpt-test:kernels:idle  | redis key to save kernels info                           |



API 
Parameters
| Name   | Required | Description                            |
| ------ | -------- | -------------------------------------- |
| userId | true     | Create a directory and cd into the directory's userId. |
``` 
POST /api/kernels/pop/ HTTP/1.1
Content-Type: application/json
Host: 127.0.0.1:8080

{
	"userId":"zt"
}

```
Response

```json
{
	"id": "7d7e2f30-60bf-415c-836c-1137256965d9",
	"name": "python_kubernetes",
	"last_activity": "2024-02-29T06:37:19.189801Z",
	"execution_state": "starting",
	"connections": 0
}
```