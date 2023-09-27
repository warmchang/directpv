# DirectPV on Red Hat OpenShift
DirectPV runs under project `directpv` in Red Hat OpenShift. Project `directpv` is automatically created after successful DirectPV installation.

## Settings required
* Add privileges to `directpv` namespace and DirectPV service account by adding `system:serviceaccount:directpv:directpv-min-io` to `users` section of YAML generated by command `$ oc edit scc privileged`

## Limitations
* DirectPV does not support volume snapshot feature as per CSI specification. DirectPV is specifically meant for use cases like MinIO where the data availability and resiliency is taken care by the application itself. Additionally, with the AWS S3 versioning APIs and internal healing, snapshots is not a requirement.
* DirectPV does not support `ReadWriteMany` volume access mode. The workloads using DirectPV run local to the node and are provisioned from local storage drives in the node. This allows the workloads to directly access data without any additional network hops, unlike remote volumes, network PVs, etc. The additional network hops may lead to poor performance and increases the complexity. With `ReadWriteOnce` access mode, DirectPV provides high performance storage for Pods.