package cattle

const composeTemplate = `
## Replicas {{range $i, $replica := .Replicas}}

# replica{{$i}}
replica{{$i}}:
  image: ${LONGHORN_IMAGE}
  entrypoint:
  - longhorn
  command:
  - replica
  - --listen
  - 0.0.0.0:9502
  - --sync-agent=false
  - /volume/{{$.Name}}
  volumes:
  - /volume/{{$.Name}}
  - /var/lib/rancher/longhorn/backups:/var/lib/rancher/longhorn/backups   #TODO :shared
  labels:
    io.rancher.sidekicks: replica-api, sync-agent
    io.rancher.container.hostname_override: container_name
    io.rancher.scheduler.affinity:container_label_ne: io.rancher.longhorn.replica.volume={{$.Name}}
    io.rancher.scheduler.affinity:container_soft: ${ORC_CONTAINER}
    io.rancher.resource.disksize.{{$.Name}}: {{$.Size}}
    io.rancher.longhorn.replica.volume: {{$.Name}}
  metadata:
    volume:
      name: {{$.Name}}
      size: {{$.Size}}

sync-agent:
  image: ${LONGHORN_IMAGE}
  entrypoint:
  - longhorn
  net: container:replica
  working_dir: /volume/{{$.Name}}
  volumes_from:
  - replica
  command:
  - sync-agent
  - --listen
  - 0.0.0.0:9504

replica-api:
  image: ${ORC_IMAGE}
  privileged: true
  pid: host
  net: container:replica
  volumes_from:
  - replica
  command:
  - longhorn-agent
  - --replica
# end replica{{$i}} {{end}}

## Controller {{with .Controller}}
controller:
  image: ${LONGHORN_IMAGE}
  command:
  - launch
  - controller
  - --listen
  - 0.0.0.0:9501
  - --frontend
  - tgt
  - {{$.Name}}
  privileged: true
  volumes:
  - /dev:/host/dev
  - /proc:/host/proc
  labels:
    io.rancher.sidekicks: controller-agent
    io.rancher.container.hostname_override: container_name
    io.rancher.scheduler.affinity:container: ${ORC_CONTAINER}
  metadata:
    volume:
      name: {{$.Name}}
      size: {{$.Size}}

controller-agent:
  image: ${ORC_IMAGE}
  net: container:controller
  volumes_from: [controller]
  command:
  - longhorn-agent
  - --controller
# end controller {{end}}

## End
`
