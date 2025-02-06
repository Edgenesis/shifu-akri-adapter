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
      - image: edgehub/camera-python:v0.0.3
        name: camera-python
        ports:
        - containerPort: 11112
        volumeMounts:
        - name: deviceshifu-config
          mountPath: "/etc/edgedevice/config"
          readOnly: true
        env:
        - name: EDGEDEVICE_NAME
          value: "edgedevice-camera"
        - name: EDGEDEVICE_NAMESPACE
          value: "devices"
        - name: IP_CAMERA_ADDRESS
          value: "{{ .Address }}"
        - name: IP_CAMERA_HTTP_PORT
          value: "80"
        - name: IP_CAMERA_RTSP_PORT
          value: "{{ .Properties.Port }}"
        - name: IP_CAMERA_USERNAME
          valueFrom:
            secretKeyRef:
              name: deviceshifu-secret
              key: username
              optional: false
        - name: IP_CAMERA_PASSWORD
          valueFrom:
            secretKeyRef:
              name: deviceshifu-secret
              key: password
              optional: false
        - name: IP_CAMERA_CONTAINER_PORT
          value: "11112"
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
