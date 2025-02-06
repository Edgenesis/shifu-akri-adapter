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
      - image: edgehub/deviceshifu-http-http:nightly
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
      - name: plc
        image: edgenesis/siemens-s7:v0.0.1
        ports:
          - containerPort: 11111
        env:
          - name: PLC_ADDRESS
            value: "{{ .Address }}"
          - name: PLC_RACK
            value: "0"
          - name: PLC_SLOT
            value: "1"
          - name: PLC_PORT
            value: "{{ .Properties.Port }}"
          - name: PLC_CONTAINER_PORT
            value: "11111"
          - name: PYTHONUNBUFFERED
            value: "1"
      volumes:
      - name: deviceshifu-config
        configMap:
          name: {{ .ConfigmapName }}
      - name: edgedevice-certificate
        configMap:
          name: edgedevice-{{ .Name }}-certificate
          optional: true
      serviceAccountName: edgedevice-sa
