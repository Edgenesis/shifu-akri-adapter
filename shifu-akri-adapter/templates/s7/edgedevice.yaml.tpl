apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-{{ .Name }}
  namespace: {{ .Namespace }}
spec:
  sku: "S7 PLC"
  connection: "Ethernet"
  address: "0.0.0.0:11111"
  protocol: "HTTP"
