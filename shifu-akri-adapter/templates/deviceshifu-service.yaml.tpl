apiVersion: v1
kind: Service
metadata:
  labels:
    app: deviceshifu-{{ .Name }}-deployment
  name: deviceshifu-{{ .Name }}
  namespace: deviceshifu
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: deviceshifu-{{ .Name }}-deployment
  type: LoadBalancer
