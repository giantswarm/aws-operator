package asg

type asg struct {
	// Name is the ASG name related to this mapping.
	Name string
	// ActiveLifecycleHook indicates the ASG has an active lifecycle hook
	// associated. We need to know this in case of a CF stack update of a HA
	// Masters setup. Then we want to drain the next ASG which has an active
	// lifecycle hook, so that we can drain one at a time.
	ActiveLifecycleHook bool
}
