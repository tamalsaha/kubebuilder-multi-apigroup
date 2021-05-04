# kubebuilder-multi-apigroup

ref: https://kubebuilder.io/migration/multi-group.html

```
# create kubebuilder project
$ kubebuilder init --domain x-helm.dev --skip-go-version-check

# convert project to multi group layout
$ kubebuilder edit --multigroup=true

# add 2 kinds to first api group
$ kubebuilder create api --group module --version v1alpha1 --kind Workflow
$ kubebuilder create api --group module --version v1alpha1 --kind Action

# add 2nd api group
$ kubebuilder create api --group core --version v1alpha1 --kind Release

# add 3rd api group
$ kubebuilder create api --group chart --version v1alpha1 --kind Repository
```