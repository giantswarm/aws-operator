[![CircleCI](https://circleci.com/gh/giantswarm/apiextensions.svg?&style=shield&circle-token=880450a6e0265218c2b1f8540e280599500bb1a6)](https://circleci.com/gh/giantswarm/apiextensions)

# apiextensions

Package apiextensions provides generated Kubernetes clients for the Giant Swarm
infrastructure.

## Contributing

### Adding a New Group and/or Version

This is example skeleton for adding new group and/or version.

- Replace `GROUP` with new group name and `VERSION` with new version name.
- Create a new package `/pkg/apis/GROUP/VERSION/`.
- Inside the package create a file `doc.go` (content below).
- Inside the package create a file `register.go` (content below).
- Edit the last argument of `generate-groups.sh` call inside
  `./scripts/gen.sh`. It has format `existingGroup:existingVersion
  GROUP:VERSION`.
- Add a new object (described in [next paragraph](#adding-a-new-custom-object)).

Example `doc.go` content.

```go
// +k8s:deepcopy-gen=package,register

// +groupName=GROUP.giantswarm.io
package VERSION
```

Example `register.go` content.

```go
package VERSION

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	group   = "GROUP.giantswarm.io"
	version = "VERSION"
)

// knownTypes is the full list of objects to register with the scheme. It
// should contain pointers of zero values of all custom objects and custom
// object lists in the group version.
var knownTypes = []runtime.Object{
		//&Object{},
		//&ObjectList{},
}

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   group,
	Version: version,
}

var (
	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme is used by the generated client.
	AddToScheme = schemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, knownTypes...)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
```

### Adding a New Custom Object

This is example skeleton for adding new object.

- Make sure group and version of the object to add exists (described in
  [previous paragraph](#adding-a-new-group-andor-version)).
- Replace `NewObj` with your object name.
- Put struct definitions inside a proper package denoted by group and version
  in file named `new_obj_types.go`. Replace `new_obj` with lowercased,
  snakecased object name.
- Add `NewObj` and `NewObjList` to `knownTypes` slice in `register.go`
- Generate client by calling `./scripts/gen.sh`.
- Commit generated code and all edits to `./scripts/gen.sh`.

```go
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NewObj godoc.
type NewObj struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.NewObjMeta `json:"metadata"`
	Spec              NewObjSpec `json:"spec"`
}

// NewObjSpec godoc.
type NewObjSpec struct {
	FieldName string `json:"fieldName", yaml:"fieldName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NewObjList godoc.
type NewObjList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []NewObj `json:"items"`
}
```

### Changing Existing Custom Object

- Make the desired changes.
- Update generated client by calling `./scripts/gen.sh`.
- Commit all changes, including generated code.

### Naming Convention

Custom object structs are placed in packages corresponding to the endpoints in
Kubernetes API. E.g. structs in package
`github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1` are created
from objects under `/apis/cluster.giantswarm.io/v1alpha1/` endpoint.

As this is common to have name collisions between field type names in different
custom objects sharing the same group and version we prefix all type names
referenced inside custom object with custom object name.

Example:

```go
type NewObj struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              NewObjSpec `json:"spec"`
}

type NewObjSpec struct {
	Cluster       NewObjCluster       `json:"cluster" yaml:"cluster"`
	VersionBundle NewObjVersionBundle `json:"versionBundle" yaml:"versionBundle"`
}

type NewObjCluster struct {
	Calico       NewObjCalico       `json:"calico" yaml:"calico"`
	DockerDaemon NewObjDockerDaemon `json:"dockerDaemon" yaml:"dockerDaemon"`
}
```
