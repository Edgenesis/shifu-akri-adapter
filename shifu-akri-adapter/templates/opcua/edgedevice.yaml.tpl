apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-{{ .Name }}
  namespace: {{ .Namespace }}
spec:
  sku: "{{ .SKU }}-test"
  connection: Ethernet
  address: {{ .Address }}
  protocol: OPCUA
  protocolSettings:
    OPCUASetting:
      SecurityMode: None
      ConnectionTimeoutInMilliseconds: 5000
      AuthenticationMode: Anonymous
