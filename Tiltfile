load('ext://k8s_attach', 'k8s_attach')
load('ext://namespace', 'namespace_create')

k8s_yaml(kustomize("./config/default"))

# Categorize our resoruces in Tilt UI.
k8s_resource(
    workload='podinfo-controller-manager', 
    objects=[
        'podinfo-system:namespace',
        'myappresources.podinfo.podinfo.com:customresourcedefinition',
        'podinfo-controller-manager:serviceaccount',
        'podinfo-leader-election-role:role',
        'podinfo-manager-role:clusterrole',
        'podinfo-metrics-reader:clusterrole',
        'podinfo-proxy-role:clusterrole',
        'podinfo-leader-election-rolebinding:rolebinding',
        'podinfo-manager-rolebinding:clusterrolebinding',
        'podinfo-proxy-rolebinding:clusterrolebinding' ],
    labels=["Podinfo-Operator"], resource_deps=[], pod_readiness = 'ignore')

# Initially build, but also update docker file automagically.
docker_build(
    'controller',
    context='.',
)

# Rebuild manifest and generate on updates.
local_resource('generate', 
               cmd='make generate', 
               deps=['./cmd/', './api/v1alpha1/myappresource_types.go'], 
               labels=["make"])
local_resource('manifest', 
               cmd='make manifests', 
               deps=['./cmd/', './internal/controller/myappresource_controller.go', './api/v1alpha1/myappresource_types.go'], 
               labels=["make"])
