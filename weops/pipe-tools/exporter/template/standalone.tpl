apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nginx-exporter-{{VERSION}}
  namespace: nginx
spec:
  serviceName: nginx-exporter-{{VERSION}}
  replicas: 1
  selector:
    matchLabels:
      app: nginx-exporter-{{VERSION}}
  template:
    metadata:
      annotations:
        telegraf.influxdata.com/interval: 1s
        telegraf.influxdata.com/inputs: |+
          [[inputs.cpu]]
            percpu = false
            totalcpu = true
            collect_cpu_time = true
            report_active = true

          [[inputs.disk]]
            ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

          [[inputs.diskio]]

          [[inputs.kernel]]

          [[inputs.mem]]

          [[inputs.processes]]

          [[inputs.system]]
            fielddrop = ["uptime_format"]

          [[inputs.net]]
            ignore_protocol_stats = true

          [[inputs.procstat]]
          ## pattern as argument for nginxrep (ie, nginxrep -f <pattern>)
            pattern = "exporter"
        telegraf.influxdata.com/class: opentsdb
        telegraf.influxdata.com/env-fieldref-NAMESPACE: metadata.namespace
        telegraf.influxdata.com/limits-cpu: '300m'
        telegraf.influxdata.com/limits-memory: '300Mi'
      labels:
        app: nginx-exporter-{{VERSION}}
        exporter_object: nginx
        object_version: {{VERSION}}
        pod_type: exporter
    spec:
      nodeSelector:
        node-role: worker
      shareProcessNamespace: true
      containers:
      - name: nginx-exporter-{{VERSION}}
        image: registry-svc:25000/library/nginx-exporter:latest
        imagePullPolicy: Always
        securityContext:
          allowPrivilegeEscalation: false
          runAsUser: 0
        args:
          - -nginx.timeout=3s
          - -nginx.retry-interval=3s
          - -nginx.vts
        env:
        - name: NGINX_SCRAPE_URI
          value: "http://nginx-{{VERSION}}.nginx:80/stub_status"
        - name: NGINX_VTS_SCRAPE_URI
          value: "http://nginx-{{VERSION}}.nginx:80/vts_status/format/json"
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
          limits:
            cpu: 500m
            memory: 500Mi
        ports:
        - containerPort: 9113

---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: nginx-exporter-{{VERSION}}
  name: nginx-exporter-{{VERSION}}
  namespace: nginx
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9113"
    prometheus.io/path: '/metrics'
spec:
  ports:
  - port: 9113
    protocol: TCP
    targetPort: 9113
  selector:
    app: nginx-exporter-{{VERSION}}
