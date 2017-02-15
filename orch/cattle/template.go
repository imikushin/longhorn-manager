package cattle

const (
	dockerComposeText = `
version: '2'
services:
  ## Replicas {{range $i, $replica := .Replicas}}

  # replica-{{$i}}
  replica-{{$i}}:
    image: ${LONGHORN_IMAGE}
    entrypoint:
    - longhorn
    command:
    - replica
    - --listen=0.0.0.0:9502
    - --sync-agent=false
    - --size={{$.Size}}
    - /volume/{{$.Name}}
    volumes:
    - /volume/{{$.Name}}
    - /var/lib/rancher/longhorn/backups:/var/lib/rancher/longhorn/backups   #TODO :shared
    labels:
      io.rancher.sidekicks: sync-agent-{{$i}} #, replica-api-{{$i}}
      io.rancher.container.hostname_override: container_name
      io.rancher.scheduler.affinity:container_label_soft_ne: io.rancher.longhorn.replica.volume={{$.Name}}
      io.rancher.scheduler.affinity:container_soft: ${ORC_CONTAINER}
      io.rancher.resource.disksize.{{$.Name}}: {{$.Size}}
      io.rancher.longhorn.replica.volume: {{$.Name}}

  sync-agent-{{$i}}:
    image: ${LONGHORN_IMAGE}
    entrypoint:
    - longhorn
    network_mode: container:replica-{{$i}}
    working_dir: /volume/{{$.Name}}
    volumes_from:
    - replica-{{$i}}
    command:
    - sync-agent
    - --listen=0.0.0.0:9504

  #replica-api-{{$i}}:
  #  image: ${ORC_IMAGE}
  #  privileged: true
  #  pid: host
  #  network_mode: container:replica-{{$i}}
  #  volumes_from:
  #  - replica-{{$i}}
  #  command:
  #  - longhorn-agent
  #  - --replica

  # end replica-{{$i}} {{end}}

  ## Controller {{with .Controller}}
  controller:
    image: ${LONGHORN_IMAGE}
    command:
    - launch
    - controller
    - --listen=0.0.0.0:9501
    - --frontend=tgt
    # {{range $.Replicas}}
    - --replica=tcp://{{.Name}}.{{StackName $.Name}}:9502
    # {{end}}
    - {{$.Name}}
    links:
    # {{range $.Replicas}}
    - {{.Name}}
    # {{end}}
    privileged: true
    volumes:
    - /dev:/host/dev
    - /proc:/host/proc
    labels:
      # io.rancher.sidekicks: controller-agent
      io.rancher.container.hostname_override: container_name
      io.rancher.scheduler.affinity:container: ${ORC_CONTAINER}

  #controller-agent:
  #  image: ${ORC_IMAGE}
  #  network_mode: container:controller
  #  volumes_from: [controller]
  #  command:
  #  - longhorn-agent
  #  - --controller

  # end controller {{end}}

  ## Metadata
  volmd:
    image: alpine
    command: [sh]
    stdin_open: true
    tty: true

  ## End
`

	rancherComposeText = `
version: '2'
services:
  volmd:
    metadata:
      volume:
        name: {{$.Name}}
        size: {{$.Size}}
        numberOfReplicas: {{$.NumberOfReplicas}}
        staleReplicaTimeout: {{$.StaleReplicaTimeout.Hours}}
`
)
