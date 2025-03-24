load('ext://ko', 'ko_build')

k8s_yaml(local('curl -L https://github.com/cert-manager/cert-manager/releases/download/v1.17.0/cert-manager.yaml'))
k8s_yaml('.local/certificate.yaml')
k8s_yaml('.local/rbac.yaml')

ko_build('backend',
         'github.com/tlscert/tlscert/server/cmd')

k8s_yaml('.local/k8s.yaml')

k8s_resource(
    workload='backend',
    port_forwards=8080
)