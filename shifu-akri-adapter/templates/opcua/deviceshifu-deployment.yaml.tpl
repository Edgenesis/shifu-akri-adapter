apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: deviceshifu-{{ .Name }}-deployment
  name: deviceshifu-{{ .Name }}-deployment
  namespace: deviceshifu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: deviceshifu-{{ .Name }}-deployment
  template:
    metadata:
      labels:
        app: deviceshifu-{{ .Name }}-deployment
    spec:
      containers:
      - image: edgehub/deviceshifu-http-opcua:nightly
        name: deviceshifu-http
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: deviceshifu-config
          mountPath: "/etc/edgedevice/config"
          readOnly: true
        - name: edgedevice-certificate
          mountPath: "/etc/edgedevice/certificate"
          readOnly: true
        env:
        - name: EDGEDEVICE_NAME
          value: "edgedevice-{{ .Name }}"
        - name: EDGEDEVICE_NAMESPACE
          value: "{{ .Namespace }}"
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
      volumes:
      - name: deviceshifu-config
        configMap:
          name: {{ .ConfigmapName }}
      - name: edgedevice-certificate
        configMap:
          name: edgedevice-{{ .Name }}-certificate
          optional: true
      serviceAccountName: edgedevice-sa
