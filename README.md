# Deadline Job Operator

To run controller locally, run below commands.

```bash
make generate
make manifests
make install run
```

To create DeadlineJob resource, run below commands.

```bash
kubectl apply -f config/samples/job_v1alpha1_deadlinejob.yaml
```

After DeadlineJob resource is created, you can see logs like this.

```
2021-10-26T23:20:32.560Z        INFO    controller_deadlinejob  === Reconcile DeadlineJob       {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.561Z        INFO    controller_deadlinejob  Phase: PENDING  {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.561Z        INFO    controller_deadlinejob  It's time to execute the job.   {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.575Z        INFO    controller_deadlinejob  === Reconcile DeadlineJob       {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.575Z        INFO    controller_deadlinejob  Phase: RUNNING  {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.607Z        INFO    controller_deadlinejob  Pod launched:   {"namespace": "default", "DeadlineJob": "deadlinejob-sample", "name": "deadlinejob-sample-pod"}
2021-10-26T23:20:32.624Z        INFO    controller_deadlinejob  === Reconcile DeadlineJob       {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.624Z        INFO    controller_deadlinejob  Phase: RUNNING  {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.647Z        INFO    controller_deadlinejob  === Reconcile DeadlineJob       {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
2021-10-26T23:20:32.647Z        INFO    controller_deadlinejob  Phase: EXCEEDED_DEADLINE        {"namespace": "default", "DeadlineJob": "deadlinejob-sample"}
```

You can check DeadlineJob resource as follows.

```
$ kubectl get deadlinejobs.job.nuka137.com
NAME                 AGE
deadlinejob-sample   80s

$ kubectl describe 
Name:         deadlinejob-sample
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  job.nuka137.com/v1alpha1
Kind:         DeadlineJob
Metadata:
  Creation Timestamp:  2021-10-26T23:20:32Z
  Generation:          1
  Managed Fields:
    API Version:  job.nuka137.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:command:
        f:jobEnd:
        f:jobStart:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-10-26T23:20:32Z
    API Version:  job.nuka137.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:phase:
    Manager:         main
    Operation:       Update
    Time:            2021-10-26T23:20:32Z
  Resource Version:  24539878
  UID:               be0f1340-dd37-4749-b5bd-23d3fdaedd2e
Spec:
  Command:    sleep 120000
  Job End:    2021-10-11T12:43:20Z
  Job Start:  2021-10-11T06:24:00Z
Status:
  Phase:  EXCEEDED_DEADLINE
Events:   <none>
```


Run commands to cleanup.

```bash
kubectl delete -f config/samples/job_v1alpha1_deadlinejob.yaml
kubectl delete crd deadlinejobs.job.nuka137.com
```

