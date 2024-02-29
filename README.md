# JPKManager
Jupyter kernels Pool manager


A http server for jupyter kernels pool management


### ENV
| NAME                    | Default Value               | Description                                              |
| ----------------------- | --------------------------- | -------------------------------------------------------- |
| JPK_EG_ENDPOINT         | http://127.0.0.1:8888       | jupyter server enterprise gateway.                       |
| JPK_MAX_PENDING_KERNELS | 10                          | The maximum number of Jupyter kernels idle.              |
| JPK_NFS_VOLUME_SERVER   | 10.0.0.29                   | NFS SERVER HOST                                          |
| JPK_NFS_MOUNT_PATH      | /data/tablegpt-test/shared/ | NFS MOUNTED PATH                                         |
| JPK_WORKING_DIR         | /mnt/shared/                | Jupyter kernel Working Dir PATH                          |
| JPK_KERNEL_IMAGE        | elyra/kernel-py:3.2.2       | Jupyter Kernel Running Image                             |
| JPK_KERNEL_NAMESPACE    | tablegpt-kernels            | Jupyter Kernel Running namespace in k8s                  |
| JPK_SERVER_PORT         | 8080                        | JPK Manager Http Server Running Port                     |
| JPK_ACTIVATION_INTERVAL | 1200                        | The activation time interval of the kernel WS connection |
| JPK_REDIS_HOST          | 127.0.0.1                   | Redis HOST                                               |
| JPK_REDIS_PORT          | 6379                        | Redis Port                                               |
| JPK_REDIS_DB            | 0                           | Redis DB                                                 |
| JPK_REDIS_KEY           | tablegpt-test:kernels:idle  | redis key to save kernels info                           |



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
