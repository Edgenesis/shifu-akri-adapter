apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-{{ .Name }}
  namespace: {{ .Namespace }}
spec:
  sku: "IP Camera"
  connection: Ethernet
  address: "0.0.0.0:11112"
  protocol: HTTP
